package scanner

import (
	"context"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/aws"
	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

type CostScanner struct {
	Client *aws.Client
}

func NewCostScanner(client *aws.Client) *CostScanner {
	return &CostScanner{Client: client}
}

func (s *CostScanner) Scan(ctx context.Context) ([]models.Resource, error) {
	costs, err := s.Client.GetWeeklyCost(ctx)
	if err != nil {
		// If CE API is not enabled, we might get an error.
		// For MVP, we can return error or empty.
		// Let's log and return empty to avoid blocking other scans.
		return nil, err
	}

	total := 0.0
	for _, v := range costs {
		total += v
	}

	accountID, _ := s.Client.GetAccountID(ctx)

	res := models.Resource{
		ID:        "weekly-cost-summary",
		AccountID: accountID,
		Type:      "Cost", // We might need to add this to ResourceType consts
		Region:    "global",
		Name:      "Weekly Cost Summary",
		Status:    "active",
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"total_cost": total,
			"breakdown":  costs,
		},
	}

	return []models.Resource{res}, nil
}
