package orchestrator

import (
	"context"
	"fmt"
	"log"

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
	log.Printf("Starting scan for account %s...", roleARN)

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

	// 3. Collect Resources
	var allResources []models.Resource

	// EC2
	if res, err := ec2Scan.Scan(ctx); err == nil {
		allResources = append(allResources, res...)
	} else {
		log.Printf("EC2 Scan failed: %v", err)
	}

	// S3
	if res, err := s3Scan.Scan(ctx); err == nil {
		allResources = append(allResources, res...)
	} else {
		log.Printf("S3 Scan failed: %v", err)
	}

	// RDS
	if res, err := rdsScan.Scan(ctx); err == nil {
		allResources = append(allResources, res...)
	} else {
		log.Printf("RDS Scan failed: %v", err)
	}

	// Cost
	if res, err := costScan.Scan(ctx); err == nil {
		allResources = append(allResources, res...)
	} else {
		log.Printf("Cost Scan failed: %v", err)
	}

	// 4. Evaluate Rules
	findings := o.Engine.Evaluate(allResources)
	log.Printf("Scan complete. Found %d issues.", len(findings))

	// 5. Persist
	scanID, err := o.DB.CreateScan(accountID)
	if err != nil {
		return err
	}
	if err := o.DB.SaveFindings(scanID, findings); err != nil {
		log.Printf("Failed to save findings: %v", err)
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
		if err := o.RunScan(ctx, acc.ID, acc.RoleARN); err != nil {
			log.Printf("Failed to scan account %s: %v", acc.RoleARN, err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("encountered %d errors during scan (check logs)", len(errs))
	}
	return nil
}
