package storage

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrEmailTaken is returned when a signup uses an already-registered email.
var ErrEmailTaken = errors.New("email already registered")

// ErrNoUser is returned when a lookup finds no matching user.
var ErrNoUser = errors.New("user not found")

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	TenantID     string
	CreatedAt    time.Time
}

func (db *DB) migrateAuth() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			tenant_id TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			user_id INTEGER NOT NULL,
			tenant_id TEXT NOT NULL,
			expires_at DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);`,
	}
	for _, q := range queries {
		if _, err := db.conn.Exec(q); err != nil {
			return fmt.Errorf("auth migration failed: %w", err)
		}
	}
	return nil
}

// NormalizeEmail lowercases and trims so "A@b.com " and "a@b.com" are one account.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// CreateUser inserts a new user with its own tenant. Caller supplies the bcrypt hash.
func (db *DB) CreateUser(email, passwordHash string) (*User, error) {
	email = NormalizeEmail(email)

	var exists int
	if err := db.conn.QueryRow(`SELECT COUNT(1) FROM users WHERE email = ?`, email).Scan(&exists); err != nil {
		return nil, err
	}
	if exists > 0 {
		return nil, ErrEmailTaken
	}

	// Each signup gets an isolated tenant so one customer never sees another's data.
	tenantID, err := randomToken(12)
	if err != nil {
		return nil, err
	}
	tenantID = "t_" + tenantID

	res, err := db.conn.Exec(
		`INSERT INTO users (email, password_hash, tenant_id) VALUES (?, ?, ?)`,
		email, passwordHash, tenantID,
	)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()

	// Mirror into tenants so billing/plan lookups have a row to work with.
	db.conn.Exec(`INSERT OR IGNORE INTO tenants (id, email) VALUES (?, ?)`, tenantID, email)

	return &User{ID: id, Email: email, PasswordHash: passwordHash, TenantID: tenantID}, nil
}

// GetUserByEmail looks up a user for login.
func (db *DB) GetUserByEmail(email string) (*User, error) {
	email = NormalizeEmail(email)
	u := &User{}
	err := db.conn.QueryRow(
		`SELECT id, email, password_hash, tenant_id FROM users WHERE email = ?`, email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.TenantID)
	if err == sql.ErrNoRows {
		return nil, ErrNoUser
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

// CountUsers reports how many accounts exist (used to show first-run hints).
func (db *DB) CountUsers() (int, error) {
	var n int
	err := db.conn.QueryRow(`SELECT COUNT(1) FROM users`).Scan(&n)
	return n, err
}

// CreateSession issues an opaque session token valid for ttl.
func (db *DB) CreateSession(userID int64, tenantID string, ttl time.Duration) (string, error) {
	token, err := randomToken(32)
	if err != nil {
		return "", err
	}
	_, err = db.conn.Exec(
		`INSERT INTO sessions (token, user_id, tenant_id, expires_at) VALUES (?, ?, ?, ?)`,
		token, userID, tenantID, time.Now().Add(ttl).UTC(),
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

// LookupSession resolves a session token to its tenant, rejecting expired ones.
func (db *DB) LookupSession(token string) (tenantID string, userID int64, ok bool) {
	if token == "" {
		return "", 0, false
	}
	var expires time.Time
	err := db.conn.QueryRow(
		`SELECT tenant_id, user_id, expires_at FROM sessions WHERE token = ?`, token,
	).Scan(&tenantID, &userID, &expires)
	if err != nil {
		return "", 0, false
	}
	if time.Now().UTC().After(expires) {
		db.conn.Exec(`DELETE FROM sessions WHERE token = ?`, token)
		return "", 0, false
	}
	return tenantID, userID, true
}

// DeleteSession logs a user out.
func (db *DB) DeleteSession(token string) {
	if token != "" {
		db.conn.Exec(`DELETE FROM sessions WHERE token = ?`, token)
	}
}

// PurgeExpiredSessions removes stale rows; safe to call periodically.
func (db *DB) PurgeExpiredSessions() {
	db.conn.Exec(`DELETE FROM sessions WHERE expires_at < ?`, time.Now().UTC())
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
