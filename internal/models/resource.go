package models

import "time"

type ResourceType string

const (
	TypeEC2      ResourceType = "EC2"
	TypeS3       ResourceType = "S3"
	TypeRDS      ResourceType = "RDS"
	TypeCost     ResourceType = "Cost"
	TypeEBS      ResourceType = "EBS"
	TypeEIP      ResourceType = "ElasticIP"
	TypeSnapshot ResourceType = "Snapshot"
)

// Confidence tells the customer how much to trust a saving estimate.
//
//	high   - provably unused (an unattached volume bills for nothing else)
//	medium - inferred from metrics or heuristics; worth a human check
//
// A wrong number quoted confidently costs more trust than a missing one.
const (
	ConfidenceHigh   = "high"
	ConfidenceMedium = "medium"
)

type Resource struct {
	ID        string
	TenantID  string
	AccountID string
	Type      ResourceType
	Region    string
	Name      string

	// Common fields
	Status    string    // "running", "stopped", etc.
	CreatedAt time.Time // LaunchTime or CreationDate

	// Dynamic metadata for rules
	Metadata map[string]interface{}
}

type Finding struct {
	ID             string
	TenantID       string
	AccountID      string
	ResourceID     string
	ResourceType   ResourceType
	Region         string
	RiskLevel      string // "HIGH", "MEDIUM", "LOW", "INFO"
	Description    string
	Recommendation string
	GeneratedAt    time.Time

	// MonthlySavingUSD is what the customer stops paying if they act on this,
	// priced from internal/pricing (real AWS list prices, not a guess).
	// Zero means this finding is about risk, not cost.
	MonthlySavingUSD float64

	// Evidence is the specific observation that triggered the rule, e.g.
	// "Unattached since 12 Jun 2026 (50 days), 100 GiB gp3".
	// Without this a customer has no reason to believe the number.
	Evidence string

	// Confidence is ConfidenceHigh or ConfidenceMedium.
	Confidence string

	// RuleID is the stable identifier, e.g. "ebs-unattached-volume".
	RuleID string
}
