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

// LegacyExternalID is the single shared value every customer used before
// per-tenant IDs existed.
//
// It is kept only so accounts connected under the old scheme keep working —
// their AWS trust policy still contains this string, and we cannot change what
// is in someone else's account. New connections get an unguessable per-tenant
// value from storage.ExternalIDForTenant.
//
// Do not use this for anything new. It appears in the public CloudFormation
// template, so it is not a secret, and a shared ExternalId lets one customer
// connect another customer's role ARN and read their account through us.
const LegacyExternalID = "cloud-guard-saas"

// resolveExternalID picks the value to present to STS.
func resolveExternalID(externalID string) string {
	if v := strings.TrimSpace(externalID); v != "" {
		return v
	}
	// An env override still wins for local testing against a hand-made role.
	if v := strings.TrimSpace(os.Getenv("CLOUDGUARD_EXTERNAL_ID")); v != "" {
		return v
	}
	return LegacyExternalID
}

// Client holds the AWS config and STS client.
type Client struct {
	Config    aws.Config
	STSClient *sts.Client
}

// NewClient establishes a session. If roleARN is provided, it assumes that role.
// Otherwise it uses default credentials (useful for local dev/hosting).
func NewClient(ctx context.Context, roleARN, externalID string) (*Client, error) {
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

		// The customer's trust policy requires a matching sts:ExternalId. That
		// condition is what stops one customer pasting another customer's role
		// ARN into their own account and reading it through us — so the value
		// must be the one *that account* was connected with, never a global.
		creds := stscreds.NewAssumeRoleProvider(stsClient, roleARN, func(o *stscreds.AssumeRoleOptions) {
			o.ExternalID = aws.String(resolveExternalID(externalID))
		})
		cfg.Credentials = aws.NewCredentialsCache(creds)
	}

	return &Client{
		Config:    cfg,
		STSClient: stsClient,
	}, nil
}

// ForRegion returns a copy of the client pointed at a different region.
//
// The credentials cache is shared, so all regions reuse the single AssumeRole
// call rather than re-assuming the customer's role once per region. Regional
// services (EC2, EBS) only ever return resources from their own region, so
// scanning one region means missing everything a customer runs elsewhere -
// which for most accounts is where the forgotten volumes actually are.
func (c *Client) ForRegion(region string) *Client {
	cfg := c.Config.Copy()
	cfg.Region = region
	return &Client{
		Config:    cfg,
		STSClient: c.STSClient,
	}
}

// ValidateRole proves the role can actually be assumed, so onboarding fails fast
// with a clear reason instead of silently succeeding and blowing up at scan time.
// Returns the customer's AWS account ID on success.
func ValidateRole(ctx context.Context, roleARN, externalID string) (string, error) {
	if !strings.HasPrefix(roleARN, "arn:aws:iam::") || !strings.Contains(roleARN, ":role/") {
		return "", fmt.Errorf("that is not an IAM role ARN - it should look like arn:aws:iam::123456789012:role/CloudGuardReadOnlyRole-us-east-1")
	}

	// Must match what the app's own IAM policy permits it to assume.
	if !strings.Contains(roleARN, ":role/CloudGuardReadOnlyRole-") {
		return "", fmt.Errorf("the role name must start with \"CloudGuardReadOnlyRole-\" (for example CloudGuardReadOnlyRole-us-east-1). Cloud Guard is only permitted to assume roles with that prefix")
	}

	client, err := NewClient(ctx, roleARN, externalID)
	if err != nil {
		return "", err
	}

	// This forces the AssumeRole to actually happen.
	accountID, err := client.GetAccountID(ctx)
	if err != nil {
		return "", explainAssumeRoleError(err, resolveExternalID(externalID))
	}
	return accountID, nil
}

// explainAssumeRoleError turns AWS's opaque errors into something a customer can act on.
func explainAssumeRoleError(err error, externalID string) error {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not authorized to perform: sts:AssumeRole"),
		strings.Contains(msg, "is not authorized to perform"):
		return fmt.Errorf("Cloud Guard is not allowed to assume this role. Check the role's Trust relationships tab - it must trust account %s. A CloudFormation service role will not work here", saasAccountHint())
	case strings.Contains(msg, "ExternalId") || strings.Contains(msg, "external id"):
		return fmt.Errorf("the role's ExternalId condition does not match. It must be exactly %q", externalID)
	case strings.Contains(msg, "AccessDenied"):
		return fmt.Errorf("AccessDenied when assuming the role. Most often the trust policy names the wrong account, or the ExternalId does not match %q. Full error: %v", externalID, err)
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
