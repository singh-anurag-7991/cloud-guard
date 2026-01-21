package rules

import (
	"fmt"
	"sort"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

type CostSummaryRule struct{}

func (r *CostSummaryRule) Name() string { return "CostSummaryRule" }

func (r *CostSummaryRule) Evaluate(res models.Resource) *models.Finding {
	if res.Type != "Cost" {
		return nil
	}

	total, _ := res.Metadata["total_cost"].(float64)
	breakdown, _ := res.Metadata["breakdown"].(map[string]float64)

	// Format breakdown for description
	// Sort by cost desc
	type kv struct {
		Key   string
		Value float64
	}
	var ss []kv
	for k, v := range breakdown {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})

	desc := fmt.Sprintf("Total Weekly Cost: $%.2f\nTop Services:\n", total)
	for i, kv := range ss {
		if i >= 5 {
			break
		}
		if kv.Value > 0 {
			desc += fmt.Sprintf("- %s: $%.2f\n", kv.Key, kv.Value)
		}
	}

	return &models.Finding{
		ResourceID:     res.ID,
		ResourceType:   "Cost",
		RiskLevel:      "INFO",
		Description:    desc,
		Recommendation: "Monitor these costs weekly.",
		GeneratedAt:    time.Now(),
	}
}
