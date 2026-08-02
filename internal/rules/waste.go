package rules

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/models"
	"github.com/singh-anurag-7991/cloud-guard/internal/pricing"
)

// The rules in this file exist because the older rules ("this instance is
// stopped") tell a customer something they already know and attach no money to
// it. These four fire on almost every real AWS account, each maps to a specific
// line on the bill, and each one names the exact resource and the exact dollar
// figure so the customer can verify it themselves.

// metaInt reads a numeric metadata value regardless of whether it arrived as an
// int32 from the AWS SDK or a float64 from a JSON round-trip.
func metaInt(m map[string]interface{}, key string) (int32, bool) {
	switch v := m[key].(type) {
	case int32:
		return v, true
	case int:
		return int32(v), true
	case int64:
		return int32(v), true
	case float64:
		return int32(v), true
	default:
		return 0, false
	}
}

func metaString(m map[string]interface{}, key string) string {
	s, _ := m[key].(string)
	return s
}

func metaTime(m map[string]interface{}, key string) (time.Time, bool) {
	t, ok := m[key].(time.Time)
	return t, ok
}

// humanAge renders a duration the way a person would say it.
func humanAge(since time.Time) string {
	days := int(time.Since(since).Hours() / 24)
	switch {
	case days >= 365:
		return fmt.Sprintf("%d year(s)", days/365)
	case days >= 30:
		return fmt.Sprintf("%d month(s)", days/30)
	default:
		return fmt.Sprintf("%d day(s)", days)
	}
}

// ---------------------------------------------------------------------------
// EBSUnattachedRule - the single most reliable source of cloud waste.
// A volume in state "available" is attached to nothing. It stores data no
// running machine can read, and AWS bills for every GiB of it every month.
// ---------------------------------------------------------------------------

type EBSUnattachedRule struct{}

func (r *EBSUnattachedRule) Name() string { return "EBSUnattachedRule" }

func (r *EBSUnattachedRule) Evaluate(res models.Resource) *models.Finding {
	if res.Type != models.TypeEBS || res.Status != "available" {
		return nil
	}

	size, ok := metaInt(res.Metadata, "size_gib")
	if !ok {
		return nil
	}
	volType := metaString(res.Metadata, "volume_type")
	saving := pricing.EBSMonthlyCost(volType, size)

	evidence := fmt.Sprintf("%d GiB %s volume, state=available (not attached to any instance)", size, volType)
	if created, ok := metaTime(res.Metadata, "created_at"); ok {
		evidence += fmt.Sprintf(", created %s ago", humanAge(created))
	}

	return &models.Finding{
		ResourceID:   res.ID,
		ResourceType: models.TypeEBS,
		Region:       res.Region,
		RiskLevel:    "HIGH",
		RuleID:       "ebs-unattached-volume",
		Description: fmt.Sprintf("EBS volume %s (%d GiB %s) is not attached to any instance and is still being billed.",
			res.ID, size, volType),
		Recommendation: "Snapshot it if the data still matters, then delete the volume. " +
			"A snapshot costs $0.05/GiB-month versus $0.08-0.125/GiB-month for a live volume.",
		MonthlySavingUSD: saving,
		Evidence:         evidence,
		// High confidence: an unattached volume is provably serving no instance.
		Confidence: models.ConfidenceHigh,
		// Snapshot first, delete second. Handing over a bare delete-volume for a
		// volume we cannot see inside would eventually destroy someone's data.
		FixCommand: fmt.Sprintf(
			"aws ec2 create-snapshot --volume-id %s --region %s --description 'pre-delete backup' && aws ec2 delete-volume --volume-id %s --region %s",
			res.ID, res.Region, res.ID, res.Region),
		GeneratedAt: time.Now(),
	}
}

// ---------------------------------------------------------------------------
// ElasticIPUnusedRule - AWS charges for an Elastic IP precisely when you are
// NOT using it. Small per-IP, but they accumulate silently and it is a
// zero-risk deletion, which makes it an easy first win for a customer.
// ---------------------------------------------------------------------------

type ElasticIPUnusedRule struct{}

func (r *ElasticIPUnusedRule) Name() string { return "ElasticIPUnusedRule" }

func (r *ElasticIPUnusedRule) Evaluate(res models.Resource) *models.Finding {
	if res.Type != models.TypeEIP || res.Status != "unassociated" {
		return nil
	}

	return &models.Finding{
		ResourceID:   res.ID,
		ResourceType: models.TypeEIP,
		Region:       res.Region,
		RiskLevel:    "MEDIUM",
		RuleID:       "eip-unassociated",
		Description: fmt.Sprintf("Elastic IP %s is not associated with any instance or network interface.",
			res.Name),
		Recommendation: "Release the Elastic IP. AWS bills unassociated addresses at " +
			"$0.005/hour; associated ones are free.",
		MonthlySavingUSD: pricing.IdleElasticIPMonthlyCost(),
		Evidence: fmt.Sprintf("Address %s has no association (no instance ID, no network interface), billed at $%.3f/hour",
			res.Name, pricing.IdleElasticIPPerHour),
		Confidence: models.ConfidenceHigh,
		FixCommand: fmt.Sprintf("aws ec2 release-address --allocation-id %s --region %s", res.ID, res.Region),
		GeneratedAt: time.Now(),
	}
}

// ---------------------------------------------------------------------------
// SnapshotStaleRule - snapshots are created by backup scripts and almost never
// cleaned up. 90 days is deliberately conservative: most retention policies are
// 7-30 days, so anything past 90 is very unlikely to be deliberate.
// ---------------------------------------------------------------------------

const defaultStaleSnapshotDays = 90

// staleSnapshotDays reads the threshold from CLOUDGUARD_SNAPSHOT_STALE_DAYS.
//
// This is configurable for two real reasons, not just for testing: retention
// policies genuinely differ between customers (7 days for a dev account, 365
// for a regulated one), and AWS gives no way to create a back-dated snapshot,
// so the only way to exercise this rule end-to-end is to lower the threshold.
func staleSnapshotDays() int {
	if v := strings.TrimSpace(os.Getenv("CLOUDGUARD_SNAPSHOT_STALE_DAYS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return n
		}
	}
	return defaultStaleSnapshotDays
}

type SnapshotStaleRule struct{}

func (r *SnapshotStaleRule) Name() string { return "SnapshotStaleRule" }

func (r *SnapshotStaleRule) Evaluate(res models.Resource) *models.Finding {
	if res.Type != models.TypeSnapshot {
		return nil
	}

	threshold := time.Duration(staleSnapshotDays()) * 24 * time.Hour
	if res.CreatedAt.IsZero() || time.Since(res.CreatedAt) < threshold {
		return nil
	}

	size, ok := metaInt(res.Metadata, "size_gib")
	if !ok {
		return nil
	}

	desc := metaString(res.Metadata, "description")
	evidence := fmt.Sprintf("%d GiB snapshot created %s ago (%s)",
		size, humanAge(res.CreatedAt), res.CreatedAt.Format("2 Jan 2006"))
	if desc != "" {
		evidence += fmt.Sprintf(", description: %q", desc)
	}

	return &models.Finding{
		ResourceID:   res.ID,
		ResourceType: models.TypeSnapshot,
		Region:       res.Region,
		RiskLevel:    "MEDIUM",
		RuleID:       "snapshot-stale",
		Description: fmt.Sprintf("Snapshot %s (%d GiB) is %s old and still being billed.",
			res.ID, size, humanAge(res.CreatedAt)),
		Recommendation: "Confirm it is not part of a retention policy or an AMI, then delete it. " +
			"Consider a lifecycle policy so this does not recur.",
		MonthlySavingUSD: pricing.SnapshotMonthlyCost(size),
		Evidence:         evidence,
		// Medium: age alone does not prove it is unwanted. It may back an AMI or
		// satisfy a compliance retention rule, so a human should confirm.
		Confidence: models.ConfidenceMedium,
		// Deliberately paired with the AMI check. Deleting a snapshot that backs
		// a registered AMI breaks every future launch from that AMI, and the
		// failure surfaces weeks later when someone tries to scale out.
		FixCommand: fmt.Sprintf(
			"aws ec2 describe-images --owners self --filters Name=block-device-mapping.snapshot-id,Values=%s --region %s  # if this returns nothing, then: aws ec2 delete-snapshot --snapshot-id %s --region %s",
			res.ID, res.Region, res.ID, res.Region),
		GeneratedAt: time.Now(),
	}
}

// ---------------------------------------------------------------------------
// EBSGp2ToGp3Rule - gp3 is 20% cheaper per GiB than gp2 and gives 3000 IOPS /
// 125 MiB/s baseline for free. For the overwhelming majority of gp2 volumes the
// migration is a single API call with no downtime and no performance loss.
// ---------------------------------------------------------------------------

type EBSGp2ToGp3Rule struct{}

func (r *EBSGp2ToGp3Rule) Name() string { return "EBSGp2ToGp3Rule" }

func (r *EBSGp2ToGp3Rule) Evaluate(res models.Resource) *models.Finding {
	if res.Type != models.TypeEBS {
		return nil
	}
	if metaString(res.Metadata, "volume_type") != "gp2" {
		return nil
	}
	// Unattached gp2 volumes are already covered by EBSUnattachedRule, which
	// recommends deleting them outright. Suggesting a migration as well would
	// double-count the saving and give contradictory advice.
	if res.Status == "available" {
		return nil
	}

	size, ok := metaInt(res.Metadata, "size_gib")
	if !ok {
		return nil
	}

	return &models.Finding{
		ResourceID:   res.ID,
		ResourceType: models.TypeEBS,
		Region:       res.Region,
		RiskLevel:    "LOW",
		RuleID:       "ebs-gp2-to-gp3",
		Description: fmt.Sprintf("EBS volume %s (%d GiB) uses gp2. Migrating to gp3 costs 20%% less for the same capacity.",
			res.ID, size),
		Recommendation: "Run modify-volume to change the type to gp3. This is an online change with " +
			"no downtime, and gp3 includes 3000 IOPS / 125 MiB/s baseline at no extra charge.",
		MonthlySavingUSD: pricing.Gp2ToGp3MonthlySaving(size),
		Evidence: fmt.Sprintf("%d GiB gp2 at $%.2f/GiB-month = $%.2f/mo; same size on gp3 at $%.2f/GiB-month = $%.2f/mo",
			size, pricing.EBSGp2PerGBMonth, pricing.EBSMonthlyCost("gp2", size),
			pricing.EBSGp3PerGBMonth, pricing.EBSMonthlyCost("gp3", size)),
		// High: this is arithmetic on published list prices, not an inference.
		Confidence: models.ConfidenceHigh,
		FixCommand: fmt.Sprintf("aws ec2 modify-volume --volume-id %s --volume-type gp3 --region %s", res.ID, res.Region),
		GeneratedAt: time.Now(),
	}
}
