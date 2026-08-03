package storage

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
)

// Per-tenant sts:ExternalId values.
//
// WHY THIS EXISTS
//
// Every customer used to share one ExternalId — "cloud-guard-saas" — which was
// also the default printed in the public CloudFormation template. That is not a
// secret, and it makes the following attack work:
//
//	1. Customer A launches the template. Their role's trust policy says
//	   "the Cloud Guard account may assume me, if it presents ExternalId X".
//	2. Customer B signs up, and pastes *customer A's role ARN* into their own
//	   Connect Account form.
//	3. Cloud Guard assumes the role with the same shared X, succeeds, and scans
//	   customer A's AWS account into customer B's dashboard.
//
// Tenant isolation in our database does not help here: the isolation is on our
// side, and the role trusts us, not a particular tenant. The ExternalId is the
// only thing that distinguishes one customer from another at the AWS boundary,
// so it has to be per-customer and unguessable.
//
// With a per-tenant value, step 3 fails: customer B's connect attempt presents
// B's ExternalId, customer A's role requires A's, and STS refuses.

func (db *DB) migrateExternalIDs() error {
	_, err := db.conn.Exec(`CREATE TABLE IF NOT EXISTS tenant_external_ids (
		tenant_id   TEXT PRIMARY KEY,
		external_id TEXT NOT NULL,
		created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

// newExternalID returns an unguessable identifier.
//
// 24 bytes of crypto/rand, URL-safe base64. Not math/rand: an ExternalId that
// can be predicted from a timestamp or a sequence is no better than the shared
// one this replaces.
func newExternalID() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "cg-" + base64.RawURLEncoding.EncodeToString(buf), nil
}

// ExternalIDForTenant returns the tenant's ExternalId, creating one on first
// use.
//
// Called both when showing the onboarding page (so the customer can paste it
// into CloudFormation) and at scan time (so we present the matching value).
// Those two must never disagree, which is why it is stored rather than derived.
func (db *DB) ExternalIDForTenant(tenantID string) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("external id: empty tenant")
	}

	var existing string
	err := db.conn.QueryRow(
		`SELECT external_id FROM tenant_external_ids WHERE tenant_id = ?`, tenantID,
	).Scan(&existing)
	if err == nil {
		return existing, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	id, err := newExternalID()
	if err != nil {
		return "", err
	}

	// INSERT OR IGNORE, then re-read. Two requests for the same new tenant can
	// race here — the onboarding page and a scan, say — and without this one of
	// them would hand out an ExternalId that never got stored.
	if _, err := db.conn.Exec(
		`INSERT OR IGNORE INTO tenant_external_ids (tenant_id, external_id) VALUES (?, ?)`,
		tenantID, id,
	); err != nil {
		return "", err
	}

	var stored string
	if err := db.conn.QueryRow(
		`SELECT external_id FROM tenant_external_ids WHERE tenant_id = ?`, tenantID,
	).Scan(&stored); err != nil {
		return "", err
	}
	return stored, nil
}
