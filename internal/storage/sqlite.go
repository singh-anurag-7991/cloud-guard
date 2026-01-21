package storage

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

type DB struct {
	conn *sql.DB
}

func InitDB(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
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

func (db *DB) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			role_arn TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS scans (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(account_id) REFERENCES accounts(id)
		);`,
		`CREATE TABLE IF NOT EXISTS findings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
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
	return nil
}

func (db *DB) AddAccount(roleARN string) (int64, error) {
	res, err := db.conn.Exec("INSERT INTO accounts (role_arn) VALUES (?)", roleARN)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) ListAccounts() ([]struct {
	ID      int64
	RoleARN string
}, error) {
	rows, err := db.conn.Query("SELECT id, role_arn FROM accounts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []struct {
		ID      int64
		RoleARN string
	}
	for rows.Next() {
		var a struct {
			ID      int64
			RoleARN string
		}
		if err := rows.Scan(&a.ID, &a.RoleARN); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (db *DB) CreateScan(accountID int64) (int64, error) {
	res, err := db.conn.Exec("INSERT INTO scans (account_id, timestamp) VALUES (?, ?)", accountID, time.Now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (db *DB) SaveFindings(scanID int64, findings []models.Finding) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO findings (scan_id, resource_id, resource_type, risk_level, description, recommendation, generated_at) VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, f := range findings {
		_, err := stmt.Exec(scanID, f.ResourceID, f.ResourceType, f.RiskLevel, f.Description, f.Recommendation, f.GeneratedAt)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (db *DB) GetLatestFindings() ([]models.Finding, error) {
	// Simple query: get all findings from the latest scan for each account?
	// MVP: Just get all findings from the last 24 hours or just the latest 50.
	rows, err := db.conn.Query("SELECT resource_id, resource_type, risk_level, description, recommendation, generated_at FROM findings ORDER BY generated_at DESC LIMIT 50")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []models.Finding
	for rows.Next() {
		var f models.Finding
		if err := rows.Scan(&f.ResourceID, &f.ResourceType, &f.RiskLevel, &f.Description, &f.Recommendation, &f.GeneratedAt); err != nil {
			log.Println("Error scanning finding:", err)
			continue
		}
		findings = append(findings, f)
	}
	return findings, nil
}
