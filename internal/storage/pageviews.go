package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// Page view counting.
//
// The number on the portfolio is a real count of real visits. A hardcoded
// "12,847 visitors" is the kind of detail a technical reader checks, and being
// caught inflating it costs more credibility than showing a small honest number.

func (db *DB) migratePageViews() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS page_views (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL,
			visitor_hash TEXT NOT NULL,
			day TEXT NOT NULL,
			viewed_at DATETIME NOT NULL
		)`,
		// One row per visitor per path per day. A refresh must not inflate the
		// count, and the unique index enforces that in the database rather than
		// relying on every caller to remember to check first.
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_page_views_unique
			ON page_views (path, visitor_hash, day)`,
		`CREATE INDEX IF NOT EXISTS idx_page_views_day ON page_views (day)`,
	}
	for _, s := range stmts {
		if _, err := db.conn.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

// VisitorHash turns an IP and user agent into an opaque identifier.
//
// The raw IP is never stored. An IP address is personal data under GDPR, and
// there is no reason to keep one just to count visits - the hash is enough to
// tell two visitors apart, and cannot be reversed back to the person.
func VisitorHash(ip, userAgent, salt string) string {
	sum := sha256.Sum256([]byte(ip + "|" + userAgent + "|" + salt))
	return hex.EncodeToString(sum[:16])
}

// RecordPageView counts one visit, ignoring repeats from the same visitor on
// the same path on the same day.
func (db *DB) RecordPageView(path, visitorHash string) error {
	now := time.Now().UTC()
	_, err := db.conn.Exec(
		`INSERT OR IGNORE INTO page_views (path, visitor_hash, day, viewed_at) VALUES (?, ?, ?, ?)`,
		path, visitorHash, now.Format("2006-01-02"), now,
	)
	return err
}

// PageViewStats is what the portfolio displays.
type PageViewStats struct {
	Total   int
	Today   int
	Last30d int
}

// GetPageViewStats returns visit counts for a path. Passing an empty path
// counts every page.
func (db *DB) GetPageViewStats(path string) (PageViewStats, error) {
	var s PageViewStats
	today := time.Now().UTC().Format("2006-01-02")
	cutoff := time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02")

	where := ""
	args := []interface{}{}
	if path != "" {
		where = " WHERE path = ?"
		args = append(args, path)
	}

	if err := db.conn.QueryRow(`SELECT COUNT(*) FROM page_views`+where, args...).Scan(&s.Total); err != nil {
		return s, err
	}

	todayWhere := " WHERE day = ?"
	todayArgs := []interface{}{today}
	if path != "" {
		todayWhere = " WHERE path = ? AND day = ?"
		todayArgs = []interface{}{path, today}
	}
	if err := db.conn.QueryRow(`SELECT COUNT(*) FROM page_views`+todayWhere, todayArgs...).Scan(&s.Today); err != nil {
		return s, err
	}

	recentWhere := " WHERE day >= ?"
	recentArgs := []interface{}{cutoff}
	if path != "" {
		recentWhere = " WHERE path = ? AND day >= ?"
		recentArgs = []interface{}{path, cutoff}
	}
	if err := db.conn.QueryRow(`SELECT COUNT(*) FROM page_views`+recentWhere, recentArgs...).Scan(&s.Last30d); err != nil {
		return s, err
	}

	return s, nil
}
