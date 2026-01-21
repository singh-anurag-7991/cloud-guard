package rules

import (
	"fmt"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

type S3PublicRule struct{}

func (r *S3PublicRule) Name() string { return "S3PublicRule" }

func (r *S3PublicRule) Evaluate(res models.Resource) *models.Finding {
	if res.Type != models.TypeS3 {
		return nil
	}

	// 1. Check verified public policy
	if isPublic, ok := res.Metadata["is_policy_public"].(bool); ok && isPublic {
		return &models.Finding{
			ResourceID:     res.ID,
			ResourceType:   models.TypeS3,
			RiskLevel:      "HIGH",
			Description:    fmt.Sprintf("S3 Bucket %s is PUBLIC via Bucket Policy.", res.Name),
			Recommendation: "Review bucket policy immediately and remove public access if not intended.",
			GeneratedAt:    time.Now(),
		}
	}

	// 2. Check if Public Access Block is missing or weak
	// If no_pab_config is true, it means we couldn't fetch it or it doesn't exist.
	// Default S3 buckets (old ones) might not have it.
	if _, noPAB := res.Metadata["no_pab_config"]; noPAB {
		// This might be noisy, so maybe LOW risk or just skip for MVP if we want to reduce noise.
		// User specifically asked for "S3 public access = true".
		// So strictly relying on is_policy_public is safer for "True Positives".
		// But let's warn if BlockPublicAccess is not fully enabled?
		// "Risk: HIGH" if public.
		return nil
	}

	// If PAB exists but allows public?
	// For MVP, sticking to "is_policy_public" is the most robust signal of actual publicness.
	// But simply LACK of Block Public Access is a risk too.
	// Let's check if all blocks are false
	/*
		blockPublicAcls := res.Metadata["block_public_acls"].(bool)
		blockPublicPolicy := res.Metadata["block_public_policy"].(bool)
		if !blockPublicAcls && !blockPublicPolicy {
			return &models.Finding{... Risk: LOW ... "Bucket does not block public access"}
		}
	*/

	return nil
}
