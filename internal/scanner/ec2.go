package scanner

import (
	"context"
	"log"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/aws"
	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

type EC2Scanner struct {
	Client *aws.Client
}

func NewEC2Scanner(client *aws.Client) *EC2Scanner {
	return &EC2Scanner{Client: client}
}

func (s *EC2Scanner) Scan(ctx context.Context) ([]models.Resource, error) {
	instances, err := s.Client.ListInstances(ctx)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	accountID, _ := s.Client.GetAccountID(ctx) // Best effort

	for _, inst := range instances {
		// Name tag
		name := ""
		for _, tag := range inst.Tags {
			if *tag.Key == "Name" {
				name = *tag.Value
				break
			}
		}

		res := models.Resource{
			ID:        *inst.InstanceId,
			AccountID: accountID,
			Type:      models.TypeEC2,
			Region:    s.Client.Config.Region,
			Name:      name,
			Status:    string(inst.State.Name),
			CreatedAt: *inst.LaunchTime,
			Metadata:  make(map[string]interface{}),
		}

		// Get 7-day Max CPU
		// Note: This is a sequential network call, might be slow for many instances.
		// For MVP, this is acceptable. Ideally use goroutines.
		maxCPU, err := s.Client.GetMaxMetric(ctx, "AWS/EC2", "CPUUtilization", map[string]string{"InstanceId": *inst.InstanceId}, 7*24*time.Hour)
		if err != nil {
			log.Printf("Failed to get metrics for %s: %v", *inst.InstanceId, err)
			// Continue without metrics (or set -1 to indicate missing)
			res.Metadata["max_cpu_7d"] = -1.0
		} else {
			res.Metadata["max_cpu_7d"] = maxCPU
		}

		res.Metadata["instance_type"] = string(inst.InstanceType)

		resources = append(resources, res)
	}

	return resources, nil
}
