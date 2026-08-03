package storage

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"time"
)

// Password reset tokens.
//
// Until now there was no recovery at all: a user who forgot their password was
// locked out of their account permanently, with their connected AWS role and
// their findings behind it.

// ResetTokenTTL is deliberately short. A reset link is a temporary key to
// somebody's account — the longer it lives, the longer a forwarded or
// intercepted email stays dangerous.
const ResetTokenTTL = time.Hour

func (db *DB) migratePasswordResets() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS password_resets (
			token_hash TEXT PRIMARY KEY,
			user_id    INTEGER NOT NULL,
			expires_at DATETIME NOT NULL,
			used_at    DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_password_resets_user ON password_resets (user_id)`,
	}
	for _, s := range stmts {
		if _, err := db.conn.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

// hashResetToken is what actually goes in the database.
//
// The raw token is only ever in the email. Storing the hash means a leaked
// database backup does not hand over working reset links for every account —
// the same reason passwords are not stored in plain text.
//
// SHA-256 rather than bcrypt here on purpose: these tokens are 32 random bytes,
// so there is nothing to brute-force, and a reset check should not cost 100ms
// of key stretching.
func hashResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// CreatePasswordReset issues a token for a user and returns the raw value to
// put in the email. It is never recoverable after this call.
func (db *DB) CreatePasswordReset(userID int64) (string, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}

	// Invalidate anything outstanding. If somebody requests a second link
	// because the first did not arrive, the first should stop working — two
	// live keys to one account is one more than necessary.
	if _, err := db.conn.Exec(
		`UPDATE password_resets SET used_at = ? WHERE user_id = ? AND used_at IS NULL`,
		time.Now().UTC(), userID,
	); err != nil {
		return "", err
	}

	if _, err := db.conn.Exec(
		`INSERT INTO password_resets (token_hash, user_id, expires_at) VALUES (?, ?, ?)`,
		hashResetToken(token), userID, time.Now().UTC().Add(ResetTokenTTL),
	); err != nil {
		return "", err
	}

	return token, nil
}

// LookupPasswordReset returns the user a valid token belongs to.
//
// Returns ok=false for tokens that are unknown, expired, or already used —
// the caller cannot tell which, and should not: a reset page that says
// "expired" versus "never existed" tells an attacker whether they guessed a
// real token.
func (db *DB) LookupPasswordReset(token string) (userID int64, ok bool) {
	if token == "" {
		return 0, false
	}

	var expires time.Time
	var usedAt sql.NullTime
	err := db.conn.QueryRow(
		`SELECT user_id, expires_at, used_at FROM password_resets WHERE token_hash = ?`,
		hashResetToken(token),
	).Scan(&userID, &expires, &usedAt)
	if err != nil {
		return 0, false
	}
	if usedAt.Valid {
		return 0, false
	}
	if time.Now().UTC().After(expires) {
		return 0, false
	}
	return userID, true
}

// ConsumePasswordReset sets a new password and burns the token in one
// transaction.
//
// Both must happen together. If the password changed but the token stayed
// live, the link in the email would keep working — anyone who later read that
// message could take the account back.
func (db *DB) ConsumePasswordReset(token, newPasswordHash string) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	hash := hashResetToken(token)

	var userID int64
	var expires time.Time
	var usedAt sql.NullTime
	if err := tx.QueryRow(
		`SELECT user_id, expires_at, used_at FROM password_resets WHERE token_hash = ?`, hash,
	).Scan(&userID, &expires, &usedAt); err != nil {
		return err
	}
	if usedAt.Valid || time.Now().UTC().After(expires) {
		return sql.ErrNoRows
	}

	if _, err := tx.Exec(
		`UPDATE users SET password_hash = ? WHERE id = ?`, newPasswordHash, userID,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`UPDATE password_resets SET used_at = ? WHERE token_hash = ?`, time.Now().UTC(), hash,
	); err != nil {
		return err
	}

	// Every existing session for this user dies with the password change.
	//
	// Someone resetting a password may be doing it precisely because another
	// person has access. Leaving that person signed in would defeat the whole
	// exercise.
	if _, err := tx.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID); err != nil {
		return err
	}

	return tx.Commit()
}

// RecentResetRequestCount reports how many links were issued for a user inside
// a window, so the handler can refuse to spam somebody's inbox.
func (db *DB) RecentResetRequestCount(userID int64, within time.Duration) (int, error) {
	var n int
	err := db.conn.QueryRow(
		`SELECT COUNT(1) FROM password_resets WHERE user_id = ? AND created_at > ?`,
		userID, time.Now().UTC().Add(-within),
	).Scan(&n)
	return n, err
}
