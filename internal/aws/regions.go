package aws

import (
	"context"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// FallbackRegions is used when DescribeRegions is denied by the customer's role.
// These are the regions the overwhelming majority of accounts actually use, so
// a partial scan is still far more useful than a single-region one.
var FallbackRegions = []string{
	"us-east-1", "us-west-2", "eu-west-1", "ap-south-1", "ap-southeast-1",
}

// ListEnabledRegions returns the regions this account has enabled.
//
// Scanning every AWS region would triple scan time for accounts that only use
// two, and opt-in regions the customer has not enabled return errors rather
// than empty results. Asking AWS which regions are actually live avoids both.
//
// CLOUDGUARD_REGIONS overrides this with a comma-separated list, which keeps
// tests and demos fast and predictable.
func (c *Client) ListEnabledRegions(ctx context.Context) []string {
	if override := strings.TrimSpace(os.Getenv("CLOUDGUARD_REGIONS")); override != "" {
		var regions []string
		for _, r := range strings.Split(override, ",") {
			if r = strings.TrimSpace(r); r != "" {
				regions = append(regions, r)
			}
		}
		if len(regions) > 0 {
			return regions
		}
	}

	client := ec2.NewFromConfig(c.Config)
	out, err := client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		// A read-only role without ec2:DescribeRegions is a plausible customer
		// setup. Degrade to the common regions instead of failing the scan.
		return FallbackRegions
	}

	var regions []string
	for _, r := range out.Regions {
		if r.RegionName == nil {
			continue
		}
		// opt-in-not-required = enabled by default; opted-in = customer turned it on.
		// "not-opted-in" regions reject API calls, so including them only produces noise.
		status := ""
		if r.OptInStatus != nil {
			status = *r.OptInStatus
		}
		if status == "not-opted-in" {
			continue
		}
		regions = append(regions, *r.RegionName)
	}

	if len(regions) == 0 {
		return FallbackRegions
	}
	return regions
}
