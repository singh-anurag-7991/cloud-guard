package scanner

import (
	"context"

	"github.com/singh-anurag-7991/cloud-guard/internal/aws"
	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

type S3Scanner struct {
	Client *aws.Client
}

func NewS3Scanner(client *aws.Client) *S3Scanner {
	return &S3Scanner{Client: client}
}

func (s *S3Scanner) Scan(ctx context.Context) ([]models.Resource, error) {
	buckets, err := s.Client.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	accountID, _ := s.Client.GetAccountID(ctx)

	for _, b := range buckets {
		res := models.Resource{
			ID:        *b.Name,
			AccountID: accountID,
			Type:      models.TypeS3,
			Region:    s.Client.Config.Region, // Buckets are global but accessed via region endpoint usually
			Name:      *b.Name,
			Status:    "active",
			CreatedAt: *b.CreationDate,
			Metadata:  make(map[string]interface{}),
		}

		// Check Public Access Block
		pab, err := s.Client.GetBucketPublicAccessBlock(ctx, *b.Name)
		if err == nil && pab != nil {
			res.Metadata["block_public_acls"] = pab.BlockPublicAcls
			res.Metadata["block_public_policy"] = pab.BlockPublicPolicy
			res.Metadata["ignore_public_acls"] = pab.IgnorePublicAcls
			res.Metadata["restrict_public_buckets"] = pab.RestrictPublicBuckets
		} else {
			// If error or nil, assume no block or error
			res.Metadata["no_pab_config"] = true
		}

		// Check Policy Status
		isPublic, err := s.Client.GetBucketPolicyStatus(ctx, *b.Name)
		if err == nil {
			res.Metadata["is_policy_public"] = isPublic
		}

		resources = append(resources, res)
	}

	return resources, nil
}
