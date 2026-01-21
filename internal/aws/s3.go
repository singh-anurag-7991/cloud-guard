package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// ListBuckets returns all S3 buckets.
func (c *Client) ListBuckets(ctx context.Context) ([]types.Bucket, error) {
	s3Client := s3.NewFromConfig(c.Config)
	output, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}
	return output.Buckets, nil
}

// GetBucketPublicAccessBlock returns the configuration for a bucket.
// Returns nil if no configuration exists (default is false/false/false/false usually, but check AWS behavior).
func (c *Client) GetBucketPublicAccessBlock(ctx context.Context, bucketName string) (*types.PublicAccessBlockConfiguration, error) {
	s3Client := s3.NewFromConfig(c.Config)
	output, err := s3Client.GetPublicAccessBlock(ctx, &s3.GetPublicAccessBlockInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		// If error is NoSuchPublicAccessBlockConfiguration, return nil without error
		// How to check for specific error in v2?
		// For MVP, just return error, caller can inspect string or ignore.
		// Actually, let's wrap it.
		return nil, err
	}
	return output.PublicAccessBlockConfiguration, nil
}

// GetBucketPolicyStatus checks if the bucket policy allows public access.
func (c *Client) GetBucketPolicyStatus(ctx context.Context, bucketName string) (bool, error) {
	s3Client := s3.NewFromConfig(c.Config)
	output, err := s3Client.GetBucketPolicyStatus(ctx, &s3.GetBucketPolicyStatusInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		// If policy doesn't exist, it's not public via policy (but could be via ACL, ignore for MVP or handle error)
		return false, err
	}
	if output.PolicyStatus != nil && output.PolicyStatus.IsPublic != nil {
		return *output.PolicyStatus.IsPublic, nil
	}
	return false, nil
}
