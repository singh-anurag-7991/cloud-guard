package scanner

import (
	"context"

	"github.com/singh-anurag-7991/cloud-guard/internal/aws"
	"github.com/singh-anurag-7991/cloud-guard/internal/models"
)

// StorageScanner collects the resources that generate the most silent AWS
// waste: EBS volumes, Elastic IPs and snapshots.
//
// These three are grouped into one scanner deliberately - they all come from
// the EC2 API, all need the same account ID lookup, and a customer thinks of
// them as one thing ("storage I forgot about"). Splitting them would triple the
// per-scan overhead for no benefit.
type StorageScanner struct {
	Client *aws.Client
}

func NewStorageScanner(client *aws.Client) *StorageScanner {
	return &StorageScanner{Client: client}
}

func (s *StorageScanner) Scan(ctx context.Context) ([]models.Resource, error) {
	accountID, _ := s.Client.GetAccountID(ctx) // best effort; findings still work without it
	region := s.Client.Config.Region

	var resources []models.Resource

	volumes, err := s.Client.ListVolumes(ctx)
	if err != nil {
		return nil, err
	}
	for _, v := range volumes {
		if v.VolumeId == nil {
			continue
		}
		res := models.Resource{
			ID:        *v.VolumeId,
			AccountID: accountID,
			Type:      models.TypeEBS,
			Region:    region,
			Name:      tagValue(v.Tags, "Name"),
			Status:    string(v.State), // "available" means unattached
			Metadata:  map[string]interface{}{},
		}
		if v.CreateTime != nil {
			res.CreatedAt = *v.CreateTime
			res.Metadata["created_at"] = *v.CreateTime
		}
		if v.Size != nil {
			res.Metadata["size_gib"] = *v.Size
		}
		res.Metadata["volume_type"] = string(v.VolumeType)
		res.Metadata["attachment_count"] = len(v.Attachments)
		resources = append(resources, res)
	}

	addresses, err := s.Client.ListAddresses(ctx)
	if err != nil {
		return nil, err
	}
	for _, a := range addresses {
		id := ""
		if a.AllocationId != nil {
			id = *a.AllocationId
		} else if a.PublicIp != nil {
			id = *a.PublicIp
		} else {
			continue
		}

		// An EIP is billed when it is attached to nothing. Both fields must be
		// empty: an address can be bound to a network interface without an
		// instance (e.g. a NAT gateway), and that still counts as in use.
		status := "associated"
		if a.InstanceId == nil && a.NetworkInterfaceId == nil {
			status = "unassociated"
		}

		name := id
		if a.PublicIp != nil {
			name = *a.PublicIp
		}

		resources = append(resources, models.Resource{
			ID:        id,
			AccountID: accountID,
			Type:      models.TypeEIP,
			Region:    region,
			Name:      name,
			Status:    status,
			Metadata: map[string]interface{}{
				"public_ip": name,
			},
		})
	}

	snapshots, err := s.Client.ListOwnedSnapshots(ctx)
	if err != nil {
		return nil, err
	}
	for _, sn := range snapshots {
		if sn.SnapshotId == nil {
			continue
		}
		res := models.Resource{
			ID:        *sn.SnapshotId,
			AccountID: accountID,
			Type:      models.TypeSnapshot,
			Region:    region,
			Name:      tagValue(sn.Tags, "Name"),
			Status:    string(sn.State),
			Metadata:  map[string]interface{}{},
		}
		if sn.StartTime != nil {
			res.CreatedAt = *sn.StartTime
		}
		// VolumeSize is the size of the source volume. AWS bills snapshots on
		// changed blocks only, so this is an upper bound - we say so in the
		// finding rather than pretending it is exact.
		if sn.VolumeSize != nil {
			res.Metadata["size_gib"] = *sn.VolumeSize
		}
		if sn.Description != nil {
			res.Metadata["description"] = *sn.Description
		}
		resources = append(resources, res)
	}

	return resources, nil
}
