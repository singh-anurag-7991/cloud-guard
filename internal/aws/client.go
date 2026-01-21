package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Client holds the AWS config and STS client.
type Client struct {
	Config    aws.Config
	STSClient *sts.Client
}

// NewClient establishes a session. If roleARN is provided, it assumes that role.
// Otherwise it uses default credentials (useful for local dev/hosting).
func NewClient(ctx context.Context, roleARN string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	stsClient := sts.NewFromConfig(cfg)

	if roleARN != "" {
		if !isValidARN(roleARN) {
			return nil, fmt.Errorf("invalid role ARN format: %s", roleARN)
		}

		creds := stscreds.NewAssumeRoleProvider(stsClient, roleARN)
		cfg.Credentials = aws.NewCredentialsCache(creds)
	}

	return &Client{
		Config:    cfg,
		STSClient: stsClient,
	}, nil
}

// GetAccountID returns the AWS Account ID of the authenticated identity.
func (c *Client) GetAccountID(ctx context.Context) (string, error) {
	input := &sts.GetCallerIdentityInput{}
	output, err := c.STSClient.GetCallerIdentity(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get caller identity: %w", err)
	}
	if output.Account == nil {
		return "", fmt.Errorf("account ID is nil")
	}
	return *output.Account, nil
}

func isValidARN(arn string) bool {
	return strings.HasPrefix(arn, "arn:aws:iam::") && strings.Contains(arn, ":role/")
}

// TrustPolicyTemplate returns the JSON policy users need to apply to their role.
// accountID is the ID of YOUR SaaS AWS account (or the one running this code).
func TrustPolicyTemplate(saasAccountID string) string {
	return fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "arn:aws:iam::%s:root"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`, saasAccountID)
}
