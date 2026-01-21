package storage

import (
	"os"
	"testing"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

func TestDB(t *testing.T) {
	tmpFile := "test_cloudguard.db"
	db, err := InitDB(tmpFile)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer os.Remove(tmpFile)

	// Test AddAccount
	roleARN := "arn:aws:iam::123:role/Test"
	accID, err := db.AddAccount(roleARN)
	if err != nil {
		t.Fatalf("AddAccount failed: %v", err)
	}

	// Test ListAccounts
	accounts, err := db.ListAccounts()
	if err != nil {
		t.Fatalf("ListAccounts failed: %v", err)
	}
	if len(accounts) != 1 || accounts[0].RoleARN != roleARN {
		t.Errorf("ListAccounts got %v, want 1 account with arn %s", accounts, roleARN)
	}

	// Test CreateScan
	scanID, err := db.CreateScan(accID)
	if err != nil {
		t.Fatalf("CreateScan failed: %v", err)
	}

	// Test SaveFindings
	f := models.Finding{
		ResourceID:     "i-123",
		ResourceType:   models.TypeEC2,
		RiskLevel:      "HIGH",
		Description:    "Test finding",
		Recommendation: "Fix it",
		GeneratedAt:    time.Now(),
	}
	err = db.SaveFindings(scanID, []models.Finding{f})
	if err != nil {
		t.Fatalf("SaveFindings failed: %v", err)
	}

	// Test GetLatestFindings
	findings, err := db.GetLatestFindings()
	if err != nil {
		t.Fatalf("GetLatestFindings failed: %v", err)
	}
	if len(findings) == 0 {
		t.Error("GetLatestFindings returned empty, want at least 1")
	}
	if findings[0].ResourceID != f.ResourceID {
		t.Errorf("Got finding resource %s, want %s", findings[0].ResourceID, f.ResourceID)
	}
}
