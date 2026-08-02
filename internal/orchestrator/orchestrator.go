package orchestrator

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"

	"github.com/singh-anurag-7991/cloud-guard/internal/alerting"
	"github.com/singh-anurag-7991/cloud-guard/internal/aws"
	"github.com/singh-anurag-7991/cloud-guard/internal/models"
	"github.com/singh-anurag-7991/cloud-guard/internal/rules"
	"github.com/singh-anurag-7991/cloud-guard/internal/scanner"
	"github.com/singh-anurag-7991/cloud-guard/internal/storage"
)

type Orchestrator struct {
	DB     *storage.DB
	Slack  *alerting.SlackWrapper
	Engine *rules.Engine
}

func New(db *storage.DB, slack *alerting.SlackWrapper) *Orchestrator {
	// Initialize rules
	r := []rules.Rule{
		&rules.EC2StoppedRule{},
		&rules.EC2IdleRule{},
		&rules.S3PublicRule{},
		&rules.RDSOverProvisionedRule{},
		&rules.CostSummaryRule{},
		// Cost-waste rules. These are the ones that put a real dollar figure in
		// front of the customer, which is the whole reason they connected an account.
		&rules.EBSUnattachedRule{},
		&rules.ElasticIPUnusedRule{},
		&rules.SnapshotStaleRule{},
		&rules.EBSGp2ToGp3Rule{},
	}
	engine := rules.NewEngine(r)

	return &Orchestrator{
		DB:     db,
		Slack:  slack,
		Engine: engine,
	}
}

func (o *Orchestrator) RunScan(ctx context.Context, accountID int64, roleARN string) error {
	return o.RunScanForTenant(ctx, "default", accountID, roleARN)
}

func (o *Orchestrator) RunScanForTenant(ctx context.Context, tenantID string, accountID int64, roleARN string) error {
	log.Printf("[Tenant: %s] Starting scan for account %s...", tenantID, roleARN)

	// 1. Create AWS Client
	client, err := aws.NewClient(ctx, roleARN)
	if err != nil {
		return err
	}

	// 2. Collect Resources.
	// Individual scanner failures are tolerated (one bad service shouldn't abort the
	// whole scan) but they are counted. If EVERY scanner fails - which is what
	// happens when AssumeRole is misconfigured - we must report an error rather than
	// silently claiming a successful scan with zero findings.
	var (
		mu             sync.Mutex
		allResources   []models.Resource
		scanErrs       []string
		scannerResults []storage.ScannerResult
		totalScanners  int
		okScanners     int
	)

	collect := func(name string, res []models.Resource, err error) {
		mu.Lock()
		defer mu.Unlock()
		totalScanners++
		if err != nil {
			log.Printf("%s Scan failed: %v", name, err)
			scanErrs = append(scanErrs, fmt.Sprintf("%s: %v", name, err))
			scannerResults = append(scannerResults, storage.ScannerResult{
				Scanner: name, Status: "failed", Message: friendlyScannerError(err),
			})
			return
		}
		for i := range res {
			res[i].TenantID = tenantID
		}
		allResources = append(allResources, res...)
		okScanners++

		// An account with 15 enabled regions produces 47 scanner rows, most of
		// them "0 resources" for regions the customer has never used. That
		// buries the two rows that matter. Only rows with something in them are
		// worth the customer's attention; the region count is reported separately.
		if len(res) > 0 {
			scannerResults = append(scannerResults, storage.ScannerResult{
				Scanner: name, Status: "ok", Resources: len(res),
			})
		}
	}

	// Global services: one call covers every region, so running them per-region
	// would just bill the customer's API quota for duplicate results.
	s3Res, s3Err := scanner.NewS3Scanner(client).Scan(ctx)
	collect("S3", s3Res, s3Err)
	costRes, costErr := scanner.NewCostScanner(client).Scan(ctx)
	collect("Cost", costRes, costErr)

	// Regional services. EC2, EBS and RDS only ever return resources from the
	// region they were called in, so a single-region scan silently misses
	// everything the customer runs elsewhere - which is usually where the
	// forgotten volumes are. Run regions concurrently but bounded, so a customer
	// with 20 enabled regions doesn't wait 20x longer or trip AWS rate limits.
	regions := client.ListEnabledRegions(ctx)
	log.Printf("[Tenant: %s] Scanning %d region(s): %s", tenantID, len(regions), strings.Join(regions, ", "))

	sem := make(chan struct{}, maxConcurrentRegions)
	var wg sync.WaitGroup
	for _, region := range regions {
		wg.Add(1)
		go func(region string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			regional := client.ForRegion(region)
			ec2Res, ec2Err := scanner.NewEC2Scanner(regional).Scan(ctx)
			collect("EC2 ("+region+")", ec2Res, ec2Err)
			stRes, stErr := scanner.NewStorageScanner(regional).Scan(ctx)
			collect("Storage ("+region+")", stRes, stErr)
			rdsRes, rdsErr := scanner.NewRDSScanner(regional).Scan(ctx)
			collect("RDS ("+region+")", rdsRes, rdsErr)
		}(region)
	}
	wg.Wait()

	if totalScanners > 0 && len(scanErrs) == totalScanners {
		return fmt.Errorf("all scanners failed (check the role ARN, its trust policy and the ExternalId): %s",
			strings.Join(scanErrs, "; "))
	}

	// A scan that inspected 18 regions and legitimately found nothing must not
	// render as a blank table - that is indistinguishable from a broken scan.
	// Say what was checked, explicitly.
	if len(scannerResults) == 0 && okScanners > 0 {
		scannerResults = append(scannerResults, storage.ScannerResult{
			Scanner: "All services",
			Status:  "ok",
			Message: fmt.Sprintf("Checked EC2, EBS, RDS and S3 across %d region(s) plus Cost Explorer. No resources found.", len(regions)),
		})
	}
	sortScannerResults(scannerResults)

	// 4. Evaluate Rules
	findings := o.Engine.Evaluate(allResources)
	for i := range findings {
		findings[i].TenantID = tenantID
		findings[i].AccountID = fmt.Sprintf("%d", accountID)
	}
	log.Printf("[Tenant: %s] Scan complete. Found %d issues.", tenantID, len(findings))

	// 5. Persist
	scanID, err := o.DB.CreateScanForTenant(tenantID, accountID)
	if err != nil {
		return err
	}
	if err := o.DB.SaveFindingsForTenant(scanID, tenantID, findings); err != nil {
		log.Printf("Failed to save findings: %v", err)
	}
	if err := o.DB.SaveScannerResults(scanID, tenantID, scannerResults); err != nil {
		log.Printf("Failed to save scanner results: %v", err)
	}

	// 6. Alert
	for _, f := range findings {
		// Only alert on High/Medium or Cost Summary
		if f.RiskLevel == "HIGH" || f.RiskLevel == "MEDIUM" || f.ResourceType == "Cost" {
			if err := o.Slack.SendAlert(f); err != nil {
				log.Printf("Failed to send slack alert: %v", err)
			}
		}
	}

	return nil
}

// maxConcurrentRegions bounds parallel region scans. Unbounded goroutines across
// 20+ regions would trip AWS API rate limits and produce throttling errors that
// look to the customer like a broken product.
const maxConcurrentRegions = 5

// sortScannerResults gives the dashboard a stable order. Without this, the
// concurrent region scans finish in a different order every run and the results
// table appears to shuffle itself between refreshes.
func sortScannerResults(results []storage.ScannerResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Scanner < results[j].Scanner
	})
}

// friendlyScannerError turns AWS errors into something actionable on the dashboard.
func friendlyScannerError(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "not enabled for cost explorer"):
		return "Cost Explorer is not enabled on this AWS account. Enable it in Billing → Cost Explorer (data takes up to 24h to appear)."
	case strings.Contains(msg, "AccessDenied"), strings.Contains(msg, "not authorized"):
		return "Access denied - the role is missing permissions for this service."
	default:
		if len(msg) > 200 {
			msg = msg[:200] + "…"
		}
		return msg
	}
}

func (o *Orchestrator) ScanAll(ctx context.Context) error {
	accounts, err := o.DB.ListAccounts()
	if err != nil {
		return err
	}

	if len(accounts) == 0 {
		return fmt.Errorf("no accounts connected")
	}

	var errs []error
	for _, acc := range accounts {
		if err := o.RunScanForTenant(ctx, acc.TenantID, acc.ID, acc.RoleARN); err != nil {
			log.Printf("Failed to scan account %s for tenant %s: %v", acc.RoleARN, acc.TenantID, err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("encountered %d errors during scan (check logs)", len(errs))
	}
	return nil
}

func (o *Orchestrator) ScanAllForTenant(ctx context.Context, tenantID string) error {
	accounts, err := o.DB.ListAccountsByTenant(tenantID)
	if err != nil {
		return err
	}

	if len(accounts) == 0 {
		return fmt.Errorf("no accounts connected for tenant: %s", tenantID)
	}

	var errs []error
	for _, acc := range accounts {
		if err := o.RunScanForTenant(ctx, tenantID, acc.ID, acc.RoleARN); err != nil {
			log.Printf("Failed to scan account %s for tenant %s: %v", acc.RoleARN, tenantID, err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("encountered %d errors during scan for tenant %s", len(errs), tenantID)
	}
	return nil
}
