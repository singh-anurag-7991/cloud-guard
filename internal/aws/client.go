package aws

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// DefaultExternalID must stay in sync with the ExternalId parameter default in
// deployments/cloudformation.yaml. Override with CLOUDGUARD_EXTERNAL_ID.
const DefaultExternalID = "cloud-guard-saas"

// ExternalID returns the sts:ExternalId value used when assuming customer roles.
func ExternalID() string {
	if v := strings.TrimSpace(os.Getenv("CLOUDGUARD_EXTERNAL_ID")); v != "" {
		return v
	}
	return DefaultExternalID
}

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

	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	stsClient := sts.NewFromConfig(cfg)

	if roleARN != "" {
		if !isValidARN(roleARN) {
			return nil, fmt.Errorf("invalid role ARN format: %s", roleARN)
		}

		// The CloudFormation trust policy enforces an sts:ExternalId condition
		// (guards against the confused-deputy problem). If we don't send a matching
		// ExternalId here, AssumeRole fails with AccessDenied and no scan can run.
		creds := stscreds.NewAssumeRoleProvider(stsClient, roleARN, func(o *stscreds.AssumeRoleOptions) {
			o.ExternalID = aws.String(ExternalID())
		})
		cfg.Credentials = aws.NewCredentialsCache(creds)
	}

	return &Client{
		Config:    cfg,
		STSClient: stsClient,
	}, nil
}

// ValidateRole proves the role can actually be assumed, so onboarding fails fast
// with a clear reason instead of silently succeeding and blowing up at scan time.
// Returns the customer's AWS account ID on success.
func ValidateRole(ctx context.Context, roleARN string) (string, error) {
	if !strings.HasPrefix(roleARN, "arn:aws:iam::") || !strings.Contains(roleARN, ":role/") {
		return "", fmt.Errorf("that is not an IAM role ARN - it should look like arn:aws:iam::123456789012:role/CloudGuardReadOnlyRole-us-east-1")
	}

	// Must match what the app's own IAM policy permits it to assume.
	if !strings.Contains(roleARN, ":role/CloudGuardReadOnlyRole-") {
		return "", fmt.Errorf("the role name must start with \"CloudGuardReadOnlyRole-\" (for example CloudGuardReadOnlyRole-us-east-1). Cloud Guard is only permitted to assume roles with that prefix")
	}

	client, err := NewClient(ctx, roleARN)
	if err != nil {
		return "", err
	}

	// This forces the AssumeRole to actually happen.
	accountID, err := client.GetAccountID(ctx)
	if err != nil {
		return "", explainAssumeRoleError(err)
	}
	return accountID, nil
}

// explainAssumeRoleError turns AWS's opaque errors into something a customer can act on.
func explainAssumeRoleError(err error) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not authorized to perform: sts:AssumeRole"),
		strings.Contains(msg, "is not authorized to perform"):
		return fmt.Errorf("Cloud Guard is not allowed to assume this role. Check the role's Trust relationships tab - it must trust account %s. A CloudFormation service role will not work here", saasAccountHint())
	case strings.Contains(msg, "ExternalId") || strings.Contains(msg, "external id"):
		return fmt.Errorf("the role's ExternalId condition does not match. It must be exactly %q", ExternalID())
	case strings.Contains(msg, "AccessDenied"):
		return fmt.Errorf("AccessDenied when assuming the role. Most often the trust policy names the wrong account, or the ExternalId does not match %q. Full error: %v", ExternalID(), err)
	case strings.Contains(msg, "does not exist") || strings.Contains(msg, "NoSuchEntity"):
		return fmt.Errorf("that role does not exist in the target AWS account - check the ARN")
	default:
		return fmt.Errorf("could not assume the role: %w", err)
	}
}

func saasAccountHint() string {
	if v := strings.TrimSpace(os.Getenv("CLOUDGUARD_SAAS_ACCOUNT_ID")); v != "" {
		return v
	}
	return "143506099819"
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
