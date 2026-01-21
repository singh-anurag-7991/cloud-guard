package rules

import (
	"fmt"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

type EC2StoppedRule struct{}

func (r *EC2StoppedRule) Name() string { return "EC2StoppedRule" }

func (r *EC2StoppedRule) Evaluate(res models.Resource) *models.Finding {
	if res.Type != models.TypeEC2 {
		return nil
	}
	if res.Status == "stopped" {
		return &models.Finding{
			ResourceID:     res.ID,
			ResourceType:   models.TypeEC2,
			RiskLevel:      "LOW",
			Description:    fmt.Sprintf("EC2 instance %s is stopped.", res.Name),
			Recommendation: "Consider terminating if this instance is no longer needed.",
			GeneratedAt:    time.Now(),
		}
	}
	return nil
}

type EC2IdleRule struct{}

func (r *EC2IdleRule) Name() string { return "EC2IdleRule" }

func (r *EC2IdleRule) Evaluate(res models.Resource) *models.Finding {
	if res.Type != models.TypeEC2 || res.Status != "running" {
		return nil
	}

	maxCPU, ok := res.Metadata["max_cpu_7d"].(float64)
	if !ok {
		return nil
	}

	if maxCPU < 5.0 {
		return &models.Finding{
			ResourceID:     res.ID,
			ResourceType:   models.TypeEC2,
			RiskLevel:      "MEDIUM",
			Description:    fmt.Sprintf("EC2 instance %s has very low CPU usage (Max %.2f%% over 7 days).", res.Name, maxCPU),
			Recommendation: "Consider downsizing or stopping this instance.",
			GeneratedAt:    time.Now(),
		}
	}
	return nil
}
