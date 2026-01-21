package rules

import (
	"fmt"
	"strings"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

type RDSOverProvisionedRule struct{}

func (r *RDSOverProvisionedRule) Name() string { return "RDSOverProvisionedRule" }

func (r *RDSOverProvisionedRule) Evaluate(res models.Resource) *models.Finding {
	if res.Type != models.TypeRDS || res.Status != "available" {
		return nil
	}

	maxCPU, ok := res.Metadata["max_cpu_7d"].(float64)
	if !ok || maxCPU < 0 {
		return nil
	}

	class, _ := res.Metadata["instance_class"].(string)

	// Rule: "large" instance + low CPU -> over-provisioned
	// Simple heuristic: if class contains "large", "xlarge" etc. and CPU < 5%
	if (strings.Contains(class, "large") || strings.Contains(class, "xlarge")) && maxCPU < 5.0 {
		return &models.Finding{
			ResourceID:     res.ID,
			ResourceType:   models.TypeRDS,
			RiskLevel:      "MEDIUM",
			Description:    fmt.Sprintf("RDS Instance %s (%s) is underutilized (Max CPU %.2f%%).", res.Name, class, maxCPU),
			Recommendation: "Consider downsizing to a smaller instance class.",
			GeneratedAt:    time.Now(),
		}
	}

	return nil
}
