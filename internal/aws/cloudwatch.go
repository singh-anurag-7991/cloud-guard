package aws

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

// GetMaxMetric returns the maximum value for a given metric over the specified duration.
func (c *Client) GetMaxMetric(ctx context.Context, namespace, metricName string, dimensions map[string]string, duration time.Duration) (float64, error) {
	cwClient := cloudwatch.NewFromConfig(c.Config)

	endTime := time.Now()
	startTime := endTime.Add(-duration)
	period := int32(3600) // 1 hour data points

	var dims []types.Dimension
	for k, v := range dimensions {
		dims = append(dims, types.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	input := &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String(namespace),
		MetricName: aws.String(metricName),
		Dimensions: dims,
		StartTime:  &startTime,
		EndTime:    &endTime,
		Period:     &period,
		Statistics: []types.Statistic{types.StatisticMaximum},
	}

	output, err := cwClient.GetMetricStatistics(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to get metric statistics: %w", err)
	}

	var maxVal float64
	for _, datapoint := range output.Datapoints {
		if datapoint.Maximum != nil {
			if *datapoint.Maximum > maxVal {
				maxVal = *datapoint.Maximum
			}
		}
	}

	return maxVal, nil
}

// Deprecated: Use GetMaxMetric instead. Keeping for backward compatibility if needed, but we will refactor.
// Actually, let's remove it or wrap it to avoid breaking changes if I wasn't refactoring callers.
// I will update callers.
