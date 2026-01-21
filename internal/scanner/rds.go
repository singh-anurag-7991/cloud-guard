package scanner

import (
	"context"
	"log"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/aws"
	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

type RDSScanner struct {
	Client *aws.Client
}

func NewRDSScanner(client *aws.Client) *RDSScanner {
	return &RDSScanner{Client: client}
}

func (s *RDSScanner) Scan(ctx context.Context) ([]models.Resource, error) {
	instances, err := s.Client.ListRDSInstances(ctx)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	accountID, _ := s.Client.GetAccountID(ctx)

	for _, db := range instances {
		res := models.Resource{
			ID:        *db.DBInstanceIdentifier,
			AccountID: accountID,
			Type:      models.TypeRDS,
			Region:    s.Client.Config.Region,
			Name:      *db.DBInstanceIdentifier,
			Status:    *db.DBInstanceStatus,
			CreatedAt: *db.InstanceCreateTime,
			Metadata:  make(map[string]interface{}),
		}

		res.Metadata["instance_class"] = *db.DBInstanceClass

		// If DB is available, get metrics
		if res.Status == "available" {
			maxCPU, err := s.Client.GetMaxMetric(ctx, "AWS/RDS", "CPUUtilization", map[string]string{"DBInstanceIdentifier": *db.DBInstanceIdentifier}, 7*24*time.Hour)
			if err != nil {
				log.Printf("Failed to get metrics for RDS %s: %v", *db.DBInstanceIdentifier, err)
				res.Metadata["max_cpu_7d"] = -1.0
			} else {
				res.Metadata["max_cpu_7d"] = maxCPU
			}
		}

		resources = append(resources, res)
	}

	return resources, nil
}
