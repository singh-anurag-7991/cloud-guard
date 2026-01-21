package rules

import (
	"testing"

	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

func TestEC2StoppedRule(t *testing.T) {
	rule := &EC2StoppedRule{}

	tests := []struct {
		name     string
		resource models.Resource
		wantRisk string // Empty if no finding expected
	}{
		{
			name: "Stopped Instance",
			resource: models.Resource{
				Type:   models.TypeEC2,
				Status: "stopped",
				Name:   "StoppedVM",
			},
			wantRisk: "LOW",
		},
		{
			name: "Running Instance",
			resource: models.Resource{
				Type:   models.TypeEC2,
				Status: "running",
				Name:   "RunningVM",
			},
			wantRisk: "",
		},
		{
			name: "Not EC2",
			resource: models.Resource{
				Type:   models.TypeS3,
				Status: "stopped",
			},
			wantRisk: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule.Evaluate(tt.resource)
			if tt.wantRisk == "" {
				if got != nil {
					t.Errorf("Evaluate() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("Evaluate() = nil, want risk %s", tt.wantRisk)
				} else if got.RiskLevel != tt.wantRisk {
					t.Errorf("RiskLevel = %s, want %s", got.RiskLevel, tt.wantRisk)
				}
			}
		})
	}
}

func TestEC2IdleRule(t *testing.T) {
	rule := &EC2IdleRule{}

	tests := []struct {
		name     string
		resource models.Resource
		wantRisk string
	}{
		{
			name: "Idle Instance (Low CPU)",
			resource: models.Resource{
				Type:   models.TypeEC2,
				Status: "running",
				Metadata: map[string]interface{}{
					"max_cpu_7d": 2.5,
				},
			},
			wantRisk: "MEDIUM",
		},
		{
			name: "Active Instance (High CPU)",
			resource: models.Resource{
				Type:   models.TypeEC2,
				Status: "running",
				Metadata: map[string]interface{}{
					"max_cpu_7d": 45.0,
				},
			},
			wantRisk: "",
		},
		{
			name: "Missing Metrics",
			resource: models.Resource{
				Type:     models.TypeEC2,
				Status:   "running",
				Metadata: map[string]interface{}{},
			},
			wantRisk: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule.Evaluate(tt.resource)
			if tt.wantRisk == "" {
				if got != nil {
					t.Errorf("Evaluate() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("Evaluate() = nil, want risk %s", tt.wantRisk)
				} else if got.RiskLevel != tt.wantRisk {
					t.Errorf("RiskLevel = %s, want %s", got.RiskLevel, tt.wantRisk)
				}
			}
		})
	}
}

func TestS3PublicRule(t *testing.T) {
	rule := &S3PublicRule{}

	tests := []struct {
		name     string
		resource models.Resource
		wantRisk string
	}{
		{
			name: "Public Policy",
			resource: models.Resource{
				Type: models.TypeS3,
				Metadata: map[string]interface{}{
					"is_policy_public": true,
				},
			},
			wantRisk: "HIGH",
		},
		{
			name: "Private",
			resource: models.Resource{
				Type: models.TypeS3,
				Metadata: map[string]interface{}{
					"is_policy_public": false,
				},
			},
			wantRisk: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule.Evaluate(tt.resource)
			if tt.wantRisk == "" {
				if got != nil {
					t.Errorf("Evaluate() = %v, want nil", got)
				}
			} else {
				if got == nil {
					t.Errorf("Evaluate() = nil, want risk %s", tt.wantRisk)
				} else if got.RiskLevel != tt.wantRisk {
					t.Errorf("RiskLevel = %s, want %s", got.RiskLevel, tt.wantRisk)
				}
			}
		})
	}
}

func BenchmarkRuleEngine(b *testing.B) {
	// Setup rules
	r := []Rule{
		&EC2StoppedRule{},
		&EC2IdleRule{},
		&S3PublicRule{},
		&RDSOverProvisionedRule{},
	}
	engine := NewEngine(r)

	// Setup large number of resources
	resources := make([]models.Resource, 1000)
	for i := 0; i < 1000; i++ {
		if i%4 == 0 {
			resources[i] = models.Resource{Type: models.TypeEC2, Status: "stopped", ID: "i-stopped"}
		} else if i%4 == 1 {
			resources[i] = models.Resource{
				Type:     models.TypeEC2,
				Status:   "running",
				ID:       "i-running",
				Metadata: map[string]interface{}{"max_cpu_7d": 1.0},
			}
		} else if i%4 == 2 {
			resources[i] = models.Resource{
				Type:     models.TypeS3,
				ID:       "s3-bucket",
				Metadata: map[string]interface{}{"is_policy_public": true},
			}
		} else {
			resources[i] = models.Resource{
				Type:     models.TypeRDS,
				Status:   "available",
				ID:       "db-prod",
				Metadata: map[string]interface{}{"instance_class": "db.m5.large", "max_cpu_7d": 2.0},
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.Evaluate(resources)
	}
}
