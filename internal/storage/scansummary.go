package storage

import (
	"fmt"
	"time"
)

// ScannerResult records what one scanner did on one scan. Previously only
// findings were persisted, so a scan that inspected real infrastructure but
// found nothing wrong was indistinguishable from a scan that failed entirely -
// both showed an empty dashboard.
type ScannerResult struct {
	Scanner   string // EC2 | S3 | RDS | Cost
	Status    string // ok | failed
	Resources int    // how many resources this scanner inspected
	Message   string // failure reason, empty when ok
}

// ScanSummary is the latest scan for a tenant, with per-scanner detail.
type ScanSummary struct {
	ScanID    int64
	AccountID int64
	RunAt     time.Time
	Results   []ScannerResult
}

func (db *DB) migrateScanSummary() error {
	q := `CREATE TABLE IF NOT EXISTS scan_results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		scan_id INTEGER NOT NULL,
		tenant_id TEXT NOT NULL,
		scanner TEXT NOT NULL,
		status TEXT NOT NULL,
		resources INTEGER DEFAULT 0,
		message TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.conn.Exec(q); err != nil {
		return fmt.Errorf("scan_results migration failed: %w", err)
	}
	_, err := db.conn.Exec(`CREATE INDEX IF NOT EXISTS idx_scan_results_scan ON scan_results(scan_id)`)
	return err
}

// SaveScannerResults records per-scanner outcomes for a scan.
func (db *DB) SaveScannerResults(scanID int64, tenantID string, results []ScannerResult) error {
	for _, r := range results {
		_, err := db.conn.Exec(
			`INSERT INTO scan_results (scan_id, tenant_id, scanner, status, resources, message)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			scanID, tenantID, r.Scanner, r.Status, r.Resources, r.Message,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetLatestScanSummary returns the most recent scan's per-scanner results for a tenant.
func (db *DB) GetLatestScanSummary(tenantID string) (*ScanSummary, error) {
	var scanID int64
	var runAt time.Time
	var accountID int64
	err := db.conn.QueryRow(
		`SELECT id, account_id, timestamp FROM scans WHERE tenant_id = ? ORDER BY id DESC LIMIT 1`,
		tenantID,
	).Scan(&scanID, &accountID, &runAt)
	if err != nil {
		return nil, err
	}

	rows, err := db.conn.Query(
		`SELECT scanner, status, resources, message FROM scan_results WHERE scan_id = ? ORDER BY id`,
		scanID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	s := &ScanSummary{ScanID: scanID, AccountID: accountID, RunAt: runAt}
	for rows.Next() {
		var r ScannerResult
		if err := rows.Scan(&r.Scanner, &r.Status, &r.Resources, &r.Message); err != nil {
			return nil, err
		}
		s.Results = append(s.Results, r)
	}
	return s, rows.Err()
}

// TotalResourcesScanned sums resources inspected in the tenant's latest scan.
func (s *ScanSummary) TotalResourcesScanned() int {
	if s == nil {
		return 0
	}
	n := 0
	for _, r := range s.Results {
		n += r.Resources
	}
	return n
}
