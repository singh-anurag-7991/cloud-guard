package storage

import (
	"path/filepath"
	"strings"
	"testing"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := InitDB(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// TestExternalIDIsStableForATenant guards the property the whole scheme rests
// on: the value shown during onboarding must be the value presented at scan
// time. If these ever diverge, every scan fails with AccessDenied and the cause
// is invisible from the outside.
func TestExternalIDIsStableForATenant(t *testing.T) {
	db := newTestDB(t)

	first, err := db.ExternalIDForTenant("t_alpha")
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if first == "" {
		t.Fatal("got an empty external id")
	}

	second, err := db.ExternalIDForTenant("t_alpha")
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if first != second {
		t.Errorf("external id changed between calls: %q then %q", first, second)
	}
}

// TestExternalIDDiffersBetweenTenants is the security property.
//
// With one shared value, customer B could paste customer A's role ARN into
// their own account and Cloud Guard would assume it successfully — reading A's
// AWS account into B's dashboard. Distinct per-tenant values make STS refuse.
func TestExternalIDDiffersBetweenTenants(t *testing.T) {
	db := newTestDB(t)

	a, err := db.ExternalIDForTenant("t_alpha")
	if err != nil {
		t.Fatalf("tenant a: %v", err)
	}
	b, err := db.ExternalIDForTenant("t_beta")
	if err != nil {
		t.Fatalf("tenant b: %v", err)
	}

	if a == b {
		t.Fatalf("two tenants share an external id (%q) — one customer could connect another's role", a)
	}
}

// TestExternalIDIsNotGuessable checks shape, not entropy: a value that is short,
// predictable, or derived from the tenant name would defeat the point even
// though it differs per tenant.
func TestExternalIDIsNotGuessable(t *testing.T) {
	db := newTestDB(t)

	id, err := db.ExternalIDForTenant("t_alpha")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	if !strings.HasPrefix(id, "cg-") {
		t.Errorf("external id %q does not carry the cg- prefix the CloudFormation template validates", id)
	}
	// 24 random bytes base64url = 32 characters, plus the prefix.
	if len(id) < 30 {
		t.Errorf("external id %q is too short to resist guessing", id)
	}
	if strings.Contains(id, "t_alpha") {
		t.Errorf("external id %q embeds the tenant id — anyone who knows the tenant could derive it", id)
	}
}

// TestExternalIDRejectsEmptyTenant stops a bug upstream from quietly minting a
// shared identity for every caller that forgot to pass a tenant.
func TestExternalIDRejectsEmptyTenant(t *testing.T) {
	db := newTestDB(t)
	if _, err := db.ExternalIDForTenant(""); err == nil {
		t.Error("expected an error for an empty tenant id, got none")
	}
}

// TestAccountRemembersItsExternalID covers the migration path: accounts
// connected before per-tenant IDs existed have an empty value and must keep
// working, because their AWS trust policy still names the old shared string.
func TestAccountRemembersItsExternalID(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.AddAccountForTenant("t_alpha", "arn:aws:iam::111111111111:role/CloudGuardReadOnlyRole-us-east-1", "cg-abc123xyz"); err != nil {
		t.Fatalf("add new-style account: %v", err)
	}
	if _, err := db.AddAccountForTenant("t_alpha", "arn:aws:iam::222222222222:role/CloudGuardReadOnlyRole-us-east-1", ""); err != nil {
		t.Fatalf("add legacy account: %v", err)
	}

	accounts, err := db.ListAccountsByTenant("t_alpha")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(accounts))
	}

	byARN := map[string]string{}
	for _, a := range accounts {
		byARN[a.RoleARN] = a.ExternalID
	}

	if got := byARN["arn:aws:iam::111111111111:role/CloudGuardReadOnlyRole-us-east-1"]; got != "cg-abc123xyz" {
		t.Errorf("new account external id = %q, want cg-abc123xyz", got)
	}
	if got := byARN["arn:aws:iam::222222222222:role/CloudGuardReadOnlyRole-us-east-1"]; got != "" {
		t.Errorf("legacy account external id = %q, want empty so it falls back to the shared value", got)
	}
}
