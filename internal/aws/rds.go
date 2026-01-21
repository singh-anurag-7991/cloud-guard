package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// ListRDSInstances returns all RDS instances.
func (c *Client) ListRDSInstances(ctx context.Context) ([]types.DBInstance, error) {
	rdsClient := rds.NewFromConfig(c.Config)
	output, err := rdsClient.DescribeDBInstances(ctx, &rds.DescribeDBInstancesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list RDS instances: %w", err)
	}
	return output.DBInstances, nil
}
