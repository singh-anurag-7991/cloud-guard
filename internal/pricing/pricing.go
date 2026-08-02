// Package pricing holds AWS list prices used to estimate savings.
//
// Every number here is a real published AWS on-demand list price, not an
// invention. The dashboard previously showed a "$620/mo saved" figure computed
// as (highCount*150 + mediumCount*100) - a number with no relationship to the
// customer's actual bill. Quoting a confident wrong number destroys trust
// faster than quoting nothing, so savings now come from this table.
//
// Prices are us-east-1 on-demand, USD, captured 2026-08-01 from
// https://aws.amazon.com/ebs/pricing/ and https://aws.amazon.com/vpc/pricing/
//
// Limitations we are honest about:
//   - Region differences are ignored; other regions cost more, so estimates
//     are conservative (we under-promise).
//   - Volume discounts, Savings Plans and RIs are not modelled.
//   - This is a static snapshot. Wire up the AWS Price List API before
//     charging customers on these numbers at scale.
package pricing

import "time"

// SourceDate is when these prices were captured. Surface it in reports so a
// customer can see how fresh the estimate is.
var SourceDate = time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

// Region these prices are quoted for.
const Region = "us-east-1"

// Per-GB-month storage prices (USD).
const (
	EBSGp2PerGBMonth = 0.10  // General Purpose SSD (gp2)
	EBSGp3PerGBMonth = 0.08  // General Purpose SSD (gp3)
	EBSIo1PerGBMonth = 0.125 // Provisioned IOPS SSD (io1)
	EBSIo2PerGBMonth = 0.125 // Provisioned IOPS SSD (io2)
	EBSSt1PerGBMonth = 0.045 // Throughput Optimized HDD (st1)
	EBSSc1PerGBMonth = 0.015 // Cold HDD (sc1)
	EBSStandardPerGB = 0.05  // Magnetic (standard, previous generation)

	SnapshotPerGBMonth = 0.05 // EBS snapshots to S3
)

// Hourly prices (USD).
const (
	// An Elastic IP not associated with a running instance is billed hourly.
	IdleElasticIPPerHour = 0.005
	NATGatewayPerHour    = 0.045
)

// HoursPerMonth is the 730-hour convention AWS itself uses for monthly estimates.
const HoursPerMonth = 730

// EBSMonthlyCost returns the monthly cost of a volume of the given type and size.
// Unknown volume types fall back to gp2 pricing, which is the most common and
// avoids over-stating savings for exotic types.
func EBSMonthlyCost(volumeType string, sizeGiB int32) float64 {
	perGB := EBSGp2PerGBMonth
	switch volumeType {
	case "gp3":
		perGB = EBSGp3PerGBMonth
	case "gp2":
		perGB = EBSGp2PerGBMonth
	case "io1":
		perGB = EBSIo1PerGBMonth
	case "io2":
		perGB = EBSIo2PerGBMonth
	case "st1":
		perGB = EBSSt1PerGBMonth
	case "sc1":
		perGB = EBSSc1PerGBMonth
	case "standard":
		perGB = EBSStandardPerGB
	}
	return perGB * float64(sizeGiB)
}

// Gp2ToGp3MonthlySaving is what a customer saves by migrating a gp2 volume to
// gp3 at the same size. gp3 also gives 3000 IOPS baseline free, so for most
// workloads this is a strict improvement.
func Gp2ToGp3MonthlySaving(sizeGiB int32) float64 {
	return (EBSGp2PerGBMonth - EBSGp3PerGBMonth) * float64(sizeGiB)
}

// SnapshotMonthlyCost prices a snapshot by its size.
func SnapshotMonthlyCost(sizeGiB int32) float64 {
	return SnapshotPerGBMonth * float64(sizeGiB)
}

// IdleElasticIPMonthlyCost is the cost of leaving one Elastic IP unassociated.
func IdleElasticIPMonthlyCost() float64 {
	return IdleElasticIPPerHour * HoursPerMonth
}
