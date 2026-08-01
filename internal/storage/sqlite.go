package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	// Pure-Go SQLite driver (no CGO). Replaced mattn/go-sqlite3 because compiling
	// its C bindings needed gcc and peaked ~900MB RSS, which OOM-killed small EC2
	// instances and made every deploy a 10-12 minute build.
	_ "modernc.org/sqlite"

	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

type DB struct {
	conn *sql.DB
}

func InitDB(path string) (*DB, error) {
	// WAL lets the background scanner write while the dashboard reads; busy_timeout
	// makes concurrent writers wait instead of failing with "database is locked".
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)", path)

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		return nil, err
	}

	return db, nil
}

// Close cleanly shuts down the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS tenants (
			id TEXT PRIMARY KEY,
			clerk_id TEXT UNIQUE,
			email TEXT NOT NULL,
			company TEXT DEFAULT '',
			plan TEXT DEFAULT 'starter',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT DEFAULT 'default',
			role_arn TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS scans (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT DEFAULT 'default',
			account_id INTEGER,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(account_id) REFERENCES accounts(id)
		);`,
		`CREATE TABLE IF NOT EXISTS findings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id TEXT DEFAULT 'default',
			scan_id INTEGER,
			resource_id TEXT,
			resource_type TEXT,
			risk_level TEXT,
			description TEXT,
			recommendation TEXT,
			generated_at DATETIME,
			FOREIGN KEY(scan_id) REFERENCES scans(id)
		);`,
	}

	for _, q := range queries {
		if _, err := db.conn.Exec(q); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	// ALTER queries for backward compatibility with existing SQLite DB files
	alterQueries := []string{
		`ALTER TABLE accounts ADD COLUMN tenant_id TEXT DEFAULT 'default'`,
		`ALTER TABLE scans ADD COLUMN tenant_id TEXT DEFAULT 'default'`,
		`ALTER TABLE findings ADD COLUMN tenant_id TEXT DEFAULT 'default'`,
	}
	for _, aq := range alterQueries {
		db.conn.Exec(aq) // Ignore duplicate column error if already exists
	}

	return nil
}

// AddAccount adds an AWS account for the default tenant.
func (db *DB) AddAccount(roleARN string) (int64, error) {
	return db.AddAccountForTenant("default", roleARN)
}

// AddAccountForTenant adds an AWS account under a specific tenant ID.
func (db *DB) AddAccountForTenant(tenantID, roleARN string) (int64, error) {
	res, err := db.conn.Exec("INSERT INTO accounts (tenant_id, role_arn) VALUES (?, ?)", tenantID, roleARN)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// AccountRecord represents a connected AWS account record.
type AccountRecord struct {
	ID       int64
	TenantID string
	RoleARN  string
}

// ListAccounts lists all accounts (across all tenants).
func (db *DB) ListAccounts() ([]AccountRecord, error) {
	rows, err := db.conn.Query("SELECT id, tenant_id, role_arn FROM accounts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []AccountRecord
	for rows.Next() {
		var a AccountRecord
		if err := rows.Scan(&a.ID, &a.TenantID, &a.RoleARN); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

// ListAccountsByTenant returns all connected accounts for a specific tenant.
func (db *DB) ListAccountsByTenant(tenantID string) ([]AccountRecord, error) {
	rows, err := db.conn.Query("SELECT id, tenant_id, role_arn FROM accounts WHERE tenant_id = ?", tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []AccountRecord
	for rows.Next() {
		var a AccountRecord
		if err := rows.Scan(&a.ID, &a.TenantID, &a.RoleARN); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (db *DB) CreateScan(accountID int64) (int64, error) {
	return db.CreateScanForTenant("default", accountID)
}

func (db *DB) CreateScanForTenant(tenantID string, accountID int64) (int64, error) {
	res, err := db.conn.Exec("INSERT INTO scans (tenant_id, account_id, timestamp) VALUES (?, ?, ?)", tenantID, accountID, time.Now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) SaveFindings(scanID int64, findings []models.Finding) error {
	return db.SaveFindingsForTenant(scanID, "default", findings)
}

func (db *DB) SaveFindingsForTenant(scanID int64, tenantID string, findings []models.Finding) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO findings (tenant_id, scan_id, resource_id, resource_type, risk_level, description, recommendation, generated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, f := range findings {
		tid := f.TenantID
		if tid == "" {
			tid = tenantID
		}
		_, err := stmt.Exec(tid, scanID, f.ResourceID, f.ResourceType, f.RiskLevel, f.Description, f.Recommendation, f.GeneratedAt)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetLatestFindings() ([]models.Finding, error) {
	return db.GetLatestFindingsByTenant("")
}

func (db *DB) GetLatestFindingsByTenant(tenantID string) ([]models.Finding, error) {
	query := "SELECT tenant_id, resource_id, resource_type, risk_level, description, recommendation, generated_at FROM findings ORDER BY generated_at DESC LIMIT 50"
	var rows *sql.Rows
	var err error

	if tenantID != "" {
		query = "SELECT tenant_id, resource_id, resource_type, risk_level, description, recommendation, generated_at FROM findings WHERE tenant_id = ? ORDER BY generated_at DESC LIMIT 50"
		rows, err = db.conn.Query(query, tenantID)
	} else {
		rows, err = db.conn.Query(query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []models.Finding
	for rows.Next() {
		var f models.Finding
		if err := rows.Scan(&f.TenantID, &f.ResourceID, &f.ResourceType, &f.RiskLevel, &f.Description, &f.Recommendation, &f.GeneratedAt); err != nil {
			log.Println("Error scanning finding:", err)
			continue
		}
		findings = append(findings, f)
	}
	return findings, nil
}
