package orchestrator

import (
	"context"
	"fmt"
	"log"
	"strings"

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

	// 2. Initialize Scanners
	ec2Scan := scanner.NewEC2Scanner(client)
	s3Scan := scanner.NewS3Scanner(client)
	rdsScan := scanner.NewRDSScanner(client)
	costScan := scanner.NewCostScanner(client)

	// 3. Collect Resources.
	// Individual scanner failures are tolerated (one bad service shouldn't abort the
	// whole scan) but they are counted. If EVERY scanner fails - which is what
	// happens when AssumeRole is misconfigured - we must report an error rather than
	// silently claiming a successful scan with zero findings.
	var allResources []models.Resource
	var scanErrs []string
	var scannerResults []storage.ScannerResult

	collect := func(name string, res []models.Resource, err error) {
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
		scannerResults = append(scannerResults, storage.ScannerResult{
			Scanner: name, Status: "ok", Resources: len(res),
		})
	}

	r1, e1 := ec2Scan.Scan(ctx)
	collect("EC2", r1, e1)
	r2, e2 := s3Scan.Scan(ctx)
	collect("S3", r2, e2)
	r3, e3 := rdsScan.Scan(ctx)
	collect("RDS", r3, e3)
	r4, e4 := costScan.Scan(ctx)
	collect("Cost", r4, e4)

	const totalScanners = 4
	if len(scanErrs) == totalScanners {
		return fmt.Errorf("all scanners failed (check the role ARN, its trust policy and the ExternalId): %s",
			strings.Join(scanErrs, "; "))
	}

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
