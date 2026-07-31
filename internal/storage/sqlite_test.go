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

func TestMultiTenantIsolation(t *testing.T) {
	tmpFile := "test_multitenant.db"
	db, err := InitDB(tmpFile)
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}
	defer db.Close()
	defer os.Remove(tmpFile)

	tenantA := "tenant-alpha"
	tenantB := "tenant-beta"

	// Add accounts for Tenant A and Tenant B
	accA, err := db.AddAccountForTenant(tenantA, "arn:aws:iam::111:role/Alpha")
	if err != nil {
		t.Fatalf("AddAccountForTenant A failed: %v", err)
	}

	accB, err := db.AddAccountForTenant(tenantB, "arn:aws:iam::222:role/Beta")
	if err != nil {
		t.Fatalf("AddAccountForTenant B failed: %v", err)
	}

	// Verify account isolation
	accountsA, err := db.ListAccountsByTenant(tenantA)
	if err != nil || len(accountsA) != 1 {
		t.Fatalf("Expected 1 account for tenantA, got %d (err: %v)", len(accountsA), err)
	}
	if accountsA[0].RoleARN != "arn:aws:iam::111:role/Alpha" {
		t.Errorf("Unexpected ARN for tenantA: %s", accountsA[0].RoleARN)
	}

	accountsB, err := db.ListAccountsByTenant(tenantB)
	if err != nil || len(accountsB) != 1 {
		t.Fatalf("Expected 1 account for tenantB, got %d (err: %v)", len(accountsB), err)
	}

	// Save findings for Tenant A
	scanA, _ := db.CreateScanForTenant(tenantA, accA)
	db.SaveFindingsForTenant(scanA, tenantA, []models.Finding{
		{TenantID: tenantA, ResourceID: "bucket-alpha", ResourceType: models.TypeS3, RiskLevel: "HIGH", GeneratedAt: time.Now()},
	})

	// Save findings for Tenant B
	scanB, _ := db.CreateScanForTenant(tenantB, accB)
	db.SaveFindingsForTenant(scanB, tenantB, []models.Finding{
		{TenantID: tenantB, ResourceID: "instance-beta", ResourceType: models.TypeEC2, RiskLevel: "LOW", GeneratedAt: time.Now()},
	})

	// Verify Tenant A only sees Tenant A's findings
	findingsA, err := db.GetLatestFindingsByTenant(tenantA)
	if err != nil || len(findingsA) != 1 {
		t.Fatalf("Expected 1 finding for tenantA, got %d (err: %v)", len(findingsA), err)
	}
	if findingsA[0].ResourceID != "bucket-alpha" {
		t.Errorf("TenantA received wrong finding: %s", findingsA[0].ResourceID)
	}

	// Verify Tenant B only sees Tenant B's findings
	findingsB, err := db.GetLatestFindingsByTenant(tenantB)
	if err != nil || len(findingsB) != 1 {
		t.Fatalf("Expected 1 finding for tenantB, got %d (err: %v)", len(findingsB), err)
	}
	if findingsB[0].ResourceID != "instance-beta" {
		t.Errorf("TenantB received wrong finding: %s", findingsB[0].ResourceID)
	}
}
