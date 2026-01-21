package models

import "time"

type ResourceType string

const (
	TypeEC2  ResourceType = "EC2"
	TypeS3   ResourceType = "S3"
	TypeRDS  ResourceType = "RDS"
	TypeCost ResourceType = "Cost"
)

type Resource struct {
	ID        string
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
	ResourceID     string
	ResourceType   ResourceType
	RiskLevel      string // "HIGH", "MEDIUM", "LOW", "INFO"
	Description    string
	Recommendation string
	GeneratedAt    time.Time
}
