package aws

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
)

// GetWeeklyCost returns the cost for the last 7 days grouped by Service.
func (c *Client) GetWeeklyCost(ctx context.Context) (map[string]float64, error) {
	ceClient := costexplorer.NewFromConfig(c.Config)

	end := time.Now().Format("2006-01-02")
	start := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

	input := &costexplorer.GetCostAndUsageInput{
		TimePeriod: &types.DateInterval{
			Start: aws.String(start),
			End:   aws.String(end),
		},
		Granularity: types.GranularityDaily,
		Metrics:     []string{"UnblendedCost"},
		GroupBy: []types.GroupDefinition{
			{
				Type: types.GroupDefinitionTypeDimension,
				Key:  aws.String("SERVICE"),
			},
		},
	}

	output, err := ceClient.GetCostAndUsage(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get cost and usage: %w", err)
	}

	costs := make(map[string]float64)
	for _, result := range output.ResultsByTime {
		for _, group := range result.Groups {
			for _, metric := range group.Metrics {
				amountStr := *metric.Amount
				amount, err := strconv.ParseFloat(amountStr, 64)
				if err != nil {
					continue
				}
				key := group.Keys[0] // Service name
				costs[key] += amount
			}
		}
	}

	return costs, nil
}
