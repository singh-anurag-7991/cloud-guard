package rules

import (
	"strings"
	"testing"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

// These tests exist to protect the one number the customer will check against
// their own AWS bill. If a saving figure is wrong, the product is worse than
// useless - it is actively misleading.

func TestEBSUnattachedRule(t *testing.T) {
	rule := &EBSUnattachedRule{}

	t.Run("prices an unattached gp3 volume from real list prices", func(t *testing.T) {
		res := models.Resource{
			ID:     "vol-0abc123",
			Type:   models.TypeEBS,
			Region: "us-east-1",
			Status: "available",
			Metadata: map[string]interface{}{
				"size_gib":    int32(500),
				"volume_type": "gp3",
			},
		}

		f := rule.Evaluate(res)
		if f == nil {
			t.Fatal("expected a finding for an unattached volume, got nil")
		}
		// 500 GiB x $0.08/GiB-month = $40.00
		if got, want := f.MonthlySavingUSD, 40.0; got != want {
			t.Errorf("MonthlySavingUSD = %.2f, want %.2f", got, want)
		}
		if f.Confidence != models.ConfidenceHigh {
			t.Errorf("Confidence = %q, want %q", f.Confidence, models.ConfidenceHigh)
		}
		if f.RuleID != "ebs-unattached-volume" {
			t.Errorf("RuleID = %q", f.RuleID)
		}
		// Evidence must name the observation, otherwise the customer has no
		// reason to believe the number.
		if !strings.Contains(f.Evidence, "500 GiB") || !strings.Contains(f.Evidence, "available") {
			t.Errorf("Evidence does not describe the observation: %q", f.Evidence)
		}
	})

	t.Run("ignores attached volumes", func(t *testing.T) {
		res := models.Resource{
			ID:     "vol-inuse",
			Type:   models.TypeEBS,
			Status: "in-use",
			Metadata: map[string]interface{}{
				"size_gib":    int32(100),
				"volume_type": "gp3",
			},
		}
		if f := rule.Evaluate(res); f != nil {
			t.Errorf("attached volume should not be flagged, got %+v", f)
		}
	})

	t.Run("ignores non-EBS resources", func(t *testing.T) {
		res := models.Resource{ID: "i-123", Type: models.TypeEC2, Status: "available"}
		if f := rule.Evaluate(res); f != nil {
			t.Errorf("non-EBS resource should not be flagged, got %+v", f)
		}
	})

	t.Run("skips volumes with no size rather than guessing", func(t *testing.T) {
		res := models.Resource{
			ID:       "vol-nosize",
			Type:     models.TypeEBS,
			Status:   "available",
			Metadata: map[string]interface{}{"volume_type": "gp3"},
		}
		if f := rule.Evaluate(res); f != nil {
			t.Errorf("a volume with unknown size must not produce a priced finding, got %+v", f)
		}
	})
}

func TestElasticIPUnusedRule(t *testing.T) {
	rule := &ElasticIPUnusedRule{}

	t.Run("flags an unassociated address", func(t *testing.T) {
		res := models.Resource{
			ID: "eipalloc-01", Type: models.TypeEIP, Region: "ap-south-1",
			Name: "52.1.2.3", Status: "unassociated",
		}
		f := rule.Evaluate(res)
		if f == nil {
			t.Fatal("expected a finding for an unassociated Elastic IP")
		}
		// $0.005/hour x 730 hours = $3.65/month
		if got, want := f.MonthlySavingUSD, 3.65; got != want {
			t.Errorf("MonthlySavingUSD = %.4f, want %.2f", got, want)
		}
	})

	t.Run("ignores an associated address", func(t *testing.T) {
		res := models.Resource{ID: "eipalloc-02", Type: models.TypeEIP, Status: "associated"}
		if f := rule.Evaluate(res); f != nil {
			t.Errorf("associated EIP is free and must not be flagged, got %+v", f)
		}
	})
}

func TestSnapshotStaleRule(t *testing.T) {
	rule := &SnapshotStaleRule{}

	t.Run("flags a snapshot older than the threshold", func(t *testing.T) {
		res := models.Resource{
			ID: "snap-old", Type: models.TypeSnapshot, Region: "us-east-1",
			CreatedAt: time.Now().Add(-200 * 24 * time.Hour),
			Metadata:  map[string]interface{}{"size_gib": int32(200)},
		}
		f := rule.Evaluate(res)
		if f == nil {
			t.Fatal("expected a finding for a 200-day-old snapshot")
		}
		// 200 GiB x $0.05/GiB-month = $10.00
		if got, want := f.MonthlySavingUSD, 10.0; got != want {
			t.Errorf("MonthlySavingUSD = %.2f, want %.2f", got, want)
		}
		// Age does not prove a snapshot is unwanted - it may back an AMI or a
		// retention policy - so this must never claim high confidence.
		if f.Confidence != models.ConfidenceMedium {
			t.Errorf("Confidence = %q, want %q", f.Confidence, models.ConfidenceMedium)
		}
	})

	t.Run("ignores a recent snapshot", func(t *testing.T) {
		res := models.Resource{
			ID: "snap-new", Type: models.TypeSnapshot,
			CreatedAt: time.Now().Add(-10 * 24 * time.Hour),
			Metadata:  map[string]interface{}{"size_gib": int32(200)},
		}
		if f := rule.Evaluate(res); f != nil {
			t.Errorf("a 10-day-old snapshot is normal backup hygiene, got %+v", f)
		}
	})

	t.Run("ignores a snapshot with no creation time", func(t *testing.T) {
		res := models.Resource{
			ID: "snap-unknown", Type: models.TypeSnapshot,
			Metadata: map[string]interface{}{"size_gib": int32(200)},
		}
		if f := rule.Evaluate(res); f != nil {
			t.Errorf("unknown age must not be treated as old, got %+v", f)
		}
	})
}

func TestEBSGp2ToGp3Rule(t *testing.T) {
	rule := &EBSGp2ToGp3Rule{}

	t.Run("prices the migration on an in-use gp2 volume", func(t *testing.T) {
		res := models.Resource{
			ID: "vol-gp2", Type: models.TypeEBS, Region: "us-east-1", Status: "in-use",
			Metadata: map[string]interface{}{"size_gib": int32(1000), "volume_type": "gp2"},
		}
		f := rule.Evaluate(res)
		if f == nil {
			t.Fatal("expected a gp2->gp3 finding")
		}
		// 1000 GiB x ($0.10 - $0.08) = $20.00
		if got, want := f.MonthlySavingUSD, 20.0; got != want {
			t.Errorf("MonthlySavingUSD = %.2f, want %.2f", got, want)
		}
	})

	t.Run("does not double-count an unattached gp2 volume", func(t *testing.T) {
		// EBSUnattachedRule already recommends deleting this volume entirely.
		// Also recommending a type migration would inflate the headline saving
		// and give the customer two contradictory instructions.
		res := models.Resource{
			ID: "vol-gp2-idle", Type: models.TypeEBS, Status: "available",
			Metadata: map[string]interface{}{"size_gib": int32(1000), "volume_type": "gp2"},
		}
		if f := rule.Evaluate(res); f != nil {
			t.Errorf("unattached gp2 volume is handled by EBSUnattachedRule, got %+v", f)
		}
	})

	t.Run("ignores volumes already on gp3", func(t *testing.T) {
		res := models.Resource{
			ID: "vol-gp3", Type: models.TypeEBS, Status: "in-use",
			Metadata: map[string]interface{}{"size_gib": int32(1000), "volume_type": "gp3"},
		}
		if f := rule.Evaluate(res); f != nil {
			t.Errorf("gp3 volume should not be flagged, got %+v", f)
		}
	})
}

// TestMetaIntAcceptsJSONRoundTrip guards a real failure mode: metadata written
// by the AWS SDK is int32, but anything that survives a JSON round-trip comes
// back as float64. A type assertion that only handles one silently drops every
// finding, which looks exactly like "the scan found nothing".
func TestMetaIntAcceptsJSONRoundTrip(t *testing.T) {
	for name, m := range map[string]map[string]interface{}{
		"int32":   {"size_gib": int32(100)},
		"int":     {"size_gib": 100},
		"int64":   {"size_gib": int64(100)},
		"float64": {"size_gib": float64(100)},
	} {
		got, ok := metaInt(m, "size_gib")
		if !ok || got != 100 {
			t.Errorf("%s: metaInt = (%d, %v), want (100, true)", name, got, ok)
		}
	}

	if _, ok := metaInt(map[string]interface{}{"size_gib": "100"}, "size_gib"); ok {
		t.Error("a string must not be accepted as a size - that would price a guess")
	}
}
