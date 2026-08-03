package storage

import (
	"testing"
	"time"
)

func seedUser(t *testing.T, db *DB, email string) *User {
	t.Helper()
	u, err := db.CreateUser(email, "$2a$10$originalhashoriginalhashoriginalhashoriginalhashoriginal")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

func TestPasswordResetHappyPath(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "a@example.com")

	token, err := db.CreatePasswordReset(user.ID)
	if err != nil {
		t.Fatalf("create reset: %v", err)
	}

	gotID, ok := db.LookupPasswordReset(token)
	if !ok || gotID != user.ID {
		t.Fatalf("lookup returned (%d, %v), want (%d, true)", gotID, ok, user.ID)
	}

	if err := db.ConsumePasswordReset(token, "$2a$10$newhashnewhashnewhashnewhashnewhashnewhashnewhashnew"); err != nil {
		t.Fatalf("consume: %v", err)
	}

	after, err := db.GetUserByEmail("a@example.com")
	if err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if after.PasswordHash != "$2a$10$newhashnewhashnewhashnewhashnewhashnewhashnewhashnew" {
		t.Error("password hash was not updated")
	}
}

// TestResetTokenIsSingleUse is the property that makes a link in an email safe
// to send. Without it, anyone who later reads that message — a shared inbox, a
// forwarded thread, a backup — can take the account again.
func TestResetTokenIsSingleUse(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "b@example.com")

	token, err := db.CreatePasswordReset(user.ID)
	if err != nil {
		t.Fatalf("create reset: %v", err)
	}

	if err := db.ConsumePasswordReset(token, "$2a$10$firstfirstfirstfirstfirstfirstfirstfirstfirstfirstfi"); err != nil {
		t.Fatalf("first consume: %v", err)
	}

	if _, ok := db.LookupPasswordReset(token); ok {
		t.Error("a used token still validates")
	}
	if err := db.ConsumePasswordReset(token, "$2a$10$secondsecondsecondsecondsecondsecondsecondsecondsec"); err == nil {
		t.Error("a used token could be consumed a second time")
	}
}

// TestRequestingASecondLinkKillsTheFirst covers the common case where a user
// does not receive the first email and asks again. Two live keys to one account
// is one more than necessary.
func TestRequestingASecondLinkKillsTheFirst(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "c@example.com")

	first, err := db.CreatePasswordReset(user.ID)
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := db.CreatePasswordReset(user.ID)
	if err != nil {
		t.Fatalf("second: %v", err)
	}

	if _, ok := db.LookupPasswordReset(first); ok {
		t.Error("the older link still works after a new one was issued")
	}
	if _, ok := db.LookupPasswordReset(second); !ok {
		t.Error("the newest link does not work")
	}
}

// TestResetSignsOutExistingSessions matters because a person resetting their
// password may be doing it to remove someone else's access. Leaving that
// session alive would defeat the point entirely.
func TestResetSignsOutExistingSessions(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "d@example.com")

	sessionToken, err := db.CreateSession(user.ID, user.TenantID, time.Hour)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}
	if _, _, ok := db.LookupSession(sessionToken); !ok {
		t.Fatal("session should be valid before the reset")
	}

	resetToken, err := db.CreatePasswordReset(user.ID)
	if err != nil {
		t.Fatalf("create reset: %v", err)
	}
	if err := db.ConsumePasswordReset(resetToken, "$2a$10$afterafterafterafterafterafterafterafterafterafterxx"); err != nil {
		t.Fatalf("consume: %v", err)
	}

	if _, _, ok := db.LookupSession(sessionToken); ok {
		t.Error("a session survived the password reset — whoever was signed in still is")
	}
}

func TestUnknownAndExpiredTokensAreRejected(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "e@example.com")

	if _, ok := db.LookupPasswordReset("not-a-real-token"); ok {
		t.Error("an invented token validated")
	}
	if _, ok := db.LookupPasswordReset(""); ok {
		t.Error("an empty token validated")
	}

	token, err := db.CreatePasswordReset(user.ID)
	if err != nil {
		t.Fatalf("create reset: %v", err)
	}
	// Push it into the past rather than sleeping for an hour.
	if _, err := db.conn.Exec(
		`UPDATE password_resets SET expires_at = ? WHERE user_id = ?`,
		time.Now().UTC().Add(-time.Minute), user.ID,
	); err != nil {
		t.Fatalf("expire token: %v", err)
	}

	if _, ok := db.LookupPasswordReset(token); ok {
		t.Error("an expired token still validates")
	}
}

// TestRawTokenIsNotStored guards against a leaked backup handing over working
// reset links for every account, the same way password hashing does.
func TestRawTokenIsNotStored(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "f@example.com")

	token, err := db.CreatePasswordReset(user.ID)
	if err != nil {
		t.Fatalf("create reset: %v", err)
	}

	var stored string
	if err := db.conn.QueryRow(
		`SELECT token_hash FROM password_resets WHERE user_id = ? AND used_at IS NULL`, user.ID,
	).Scan(&stored); err != nil {
		t.Fatalf("read stored token: %v", err)
	}

	if stored == token {
		t.Error("the raw reset token is stored in the database")
	}
}

func TestRecentResetRequestCount(t *testing.T) {
	db := newTestDB(t)
	user := seedUser(t, db, "g@example.com")

	for i := 0; i < 3; i++ {
		if _, err := db.CreatePasswordReset(user.ID); err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
	}

	n, err := db.RecentResetRequestCount(user.ID, time.Hour)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 3 {
		t.Errorf("counted %d requests in the last hour, want 3", n)
	}
}
