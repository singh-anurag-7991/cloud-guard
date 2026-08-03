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

		// Savings columns. Before these existed the dashboard invented a total
		// from (highCount*150 + mediumCount*100), which had no connection to the
		// customer's real bill. Findings now carry a priced, evidenced number.
		`ALTER TABLE findings ADD COLUMN region TEXT DEFAULT ''`,
		`ALTER TABLE findings ADD COLUMN monthly_saving_usd REAL DEFAULT 0`,
		`ALTER TABLE findings ADD COLUMN evidence TEXT DEFAULT ''`,
		`ALTER TABLE findings ADD COLUMN confidence TEXT DEFAULT ''`,
		`ALTER TABLE findings ADD COLUMN rule_id TEXT DEFAULT ''`,
		`ALTER TABLE findings ADD COLUMN fix_command TEXT DEFAULT ''`,

		// The ExternalId this account was connected with.
		//
		// Stored per account rather than looked up per tenant at scan time,
		// because the two can legitimately differ: accounts connected before
		// per-tenant IDs existed used the old shared value, and their AWS role
		// still has that value in its trust policy. Re-deriving would present
		// the wrong ExternalId and break every scan for existing customers.
		// Empty means "connected before this change" and falls back to the
		// legacy default.
		`ALTER TABLE accounts ADD COLUMN external_id TEXT DEFAULT ''`,
	}
	for _, aq := range alterQueries {
		db.conn.Exec(aq) // Ignore duplicate column error if already exists
	}

	if err := db.migrateAuth(); err != nil {
		return err
	}

	if err := db.migrateScanSummary(); err != nil {
		return err
	}

	if err := db.migratePageViews(); err != nil {
		return err
	}

	if err := db.migrateExternalIDs(); err != nil {
		return err
	}

	return nil
}

// AddAccount adds an AWS account for the default tenant.
func (db *DB) AddAccount(roleARN string) (int64, error) {
	return db.AddAccountForTenant("default", roleARN, "")
}

// AddAccountForTenant adds an AWS account under a specific tenant ID.
//
// externalID is recorded as it was at connect time. Storing it means a later
// change to the tenant's ExternalId cannot silently break an already-connected
// account whose AWS trust policy still names the old value.
func (db *DB) AddAccountForTenant(tenantID, roleARN, externalID string) (int64, error) {
	res, err := db.conn.Exec(
		"INSERT INTO accounts (tenant_id, role_arn, external_id) VALUES (?, ?, ?)",
		tenantID, roleARN, externalID,
	)
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
	// ExternalID is empty for accounts connected before per-tenant IDs existed.
	// The AWS client falls back to the legacy shared value in that case.
	ExternalID string
}

const accountColumns = "id, tenant_id, role_arn, COALESCE(external_id, '')"

// scanAccounts reads rows into AccountRecords. Shared so the two list queries
// cannot drift in column order — a mistake that compiles cleanly and then puts
// a role ARN in the ExternalID field.
func scanAccounts(rows *sql.Rows) ([]AccountRecord, error) {
	defer rows.Close()
	var accounts []AccountRecord
	for rows.Next() {
		var a AccountRecord
		if err := rows.Scan(&a.ID, &a.TenantID, &a.RoleARN, &a.ExternalID); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}

// ListAccounts lists all accounts (across all tenants).
func (db *DB) ListAccounts() ([]AccountRecord, error) {
	rows, err := db.conn.Query("SELECT " + accountColumns + " FROM accounts")
	if err != nil {
		return nil, err
	}
	return scanAccounts(rows)
}

// ListAccountsByTenant returns all connected accounts for a specific tenant.
func (db *DB) ListAccountsByTenant(tenantID string) ([]AccountRecord, error) {
	rows, err := db.conn.Query(
		"SELECT "+accountColumns+" FROM accounts WHERE tenant_id = ?", tenantID,
	)
	if err != nil {
		return nil, err
	}
	return scanAccounts(rows)
}

// DeleteAccountForTenant removes a connected AWS account and everything derived
// from it. Scoped by tenant_id so one customer can never delete another's data.
func (db *DB) DeleteAccountForTenant(tenantID string, accountID int64) error {
	res, err := db.conn.Exec(`DELETE FROM accounts WHERE id = ? AND tenant_id = ?`, accountID, tenantID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("account not found")
	}

	// Clean up the scans and findings that belonged to it.
	db.conn.Exec(`DELETE FROM findings WHERE tenant_id = ? AND scan_id IN (SELECT id FROM scans WHERE account_id = ? AND tenant_id = ?)`, tenantID, accountID, tenantID)
	db.conn.Exec(`DELETE FROM scan_results WHERE tenant_id = ? AND scan_id IN (SELECT id FROM scans WHERE account_id = ? AND tenant_id = ?)`, tenantID, accountID, tenantID)
	db.conn.Exec(`DELETE FROM scans WHERE account_id = ? AND tenant_id = ?`, accountID, tenantID)
	return nil
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

	stmt, err := tx.Prepare(`INSERT INTO findings
		(tenant_id, scan_id, resource_id, resource_type, risk_level, description, recommendation, generated_at,
		 region, monthly_saving_usd, evidence, confidence, rule_id, fix_command)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, f := range findings {
		tid := f.TenantID
		if tid == "" {
			tid = tenantID
		}
		_, err := stmt.Exec(tid, scanID, f.ResourceID, f.ResourceType, f.RiskLevel, f.Description, f.Recommendation, f.GeneratedAt,
			f.Region, f.MonthlySavingUSD, f.Evidence, f.Confidence, f.RuleID, f.FixCommand)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// findingColumns is shared between the two read paths so they can never drift.
const findingColumns = `tenant_id, resource_id, resource_type, risk_level, description, recommendation,
	generated_at, region, monthly_saving_usd, evidence, confidence, rule_id, fix_command`

func (db *DB) GetLatestFindings() ([]models.Finding, error) {
	return db.GetLatestFindingsByTenant("")
}

func (db *DB) GetLatestFindingsByTenant(tenantID string) ([]models.Finding, error) {
	// Ordered by money first: a customer scanning their bill wants the $40/mo
	// unattached volume above the $0.30/mo stale snapshot, not whichever the
	// scanner happened to write last.
	const order = " ORDER BY monthly_saving_usd DESC, generated_at DESC LIMIT 200"

	var rows *sql.Rows
	var err error
	if tenantID != "" {
		rows, err = db.conn.Query("SELECT "+findingColumns+" FROM findings WHERE tenant_id = ?"+order, tenantID)
	} else {
		rows, err = db.conn.Query("SELECT " + findingColumns + " FROM findings" + order)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []models.Finding
	for rows.Next() {
		var f models.Finding
		// Columns added by ALTER are NULL on rows written before the migration,
		// so the nullable wrappers stop old findings from breaking the dashboard.
		var region, evidence, confidence, ruleID, fixCmd sql.NullString
		var saving sql.NullFloat64
		if err := rows.Scan(&f.TenantID, &f.ResourceID, &f.ResourceType, &f.RiskLevel, &f.Description,
			&f.Recommendation, &f.GeneratedAt, &region, &saving, &evidence, &confidence, &ruleID, &fixCmd); err != nil {
			log.Println("Error scanning finding:", err)
			continue
		}
		f.Region = region.String
		f.MonthlySavingUSD = saving.Float64
		f.Evidence = evidence.String
		f.Confidence = confidence.String
		f.RuleID = ruleID.String
		f.FixCommand = fixCmd.String
		findings = append(findings, f)
	}
	return findings, nil
}
