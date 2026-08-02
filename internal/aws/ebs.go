package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// ListVolumes returns all EBS volumes in the client's region.
//
// EBS is where the easiest cloud waste lives: volumes routinely outlive the
// instances they were attached to and keep billing indefinitely, and gp2
// volumes cost 25% more than the equivalent gp3 for no benefit.
func (c *Client) ListVolumes(ctx context.Context) ([]types.Volume, error) {
	client := ec2.NewFromConfig(c.Config)
	var volumes []types.Volume
	var nextToken *string

	for {
		out, err := client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{NextToken: nextToken})
		if err != nil {
			return nil, fmt.Errorf("failed to describe volumes: %w", err)
		}
		volumes = append(volumes, out.Volumes...)
		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}
	return volumes, nil
}

// ListAddresses returns Elastic IPs. An EIP not associated with a running
// instance bills hourly for doing nothing.
func (c *Client) ListAddresses(ctx context.Context) ([]types.Address, error) {
	client := ec2.NewFromConfig(c.Config)
	out, err := client.DescribeAddresses(ctx, &ec2.DescribeAddressesInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to describe addresses: %w", err)
	}
	return out.Addresses, nil
}

// ListOwnedSnapshots returns snapshots owned by this account.
//
// Restricted to self-owned snapshots on purpose: without the owner filter AWS
// returns every public snapshot in the region, which is tens of thousands of
// results that have nothing to do with the customer's bill.
func (c *Client) ListOwnedSnapshots(ctx context.Context) ([]types.Snapshot, error) {
	client := ec2.NewFromConfig(c.Config)
	var snapshots []types.Snapshot
	var nextToken *string

	for {
		out, err := client.DescribeSnapshots(ctx, &ec2.DescribeSnapshotsInput{
			OwnerIds:  []string{"self"},
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe snapshots: %w", err)
		}
		snapshots = append(snapshots, out.Snapshots...)
		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}
	return snapshots, nil
}
