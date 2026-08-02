package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/singh-anurag-7991/cloud-guard/internal/auth"
	cloudguardaws "github.com/singh-anurag-7991/cloud-guard/internal/aws"
	"github.com/singh-anurag-7991/cloud-guard/internal/models"
	"github.com/singh-anurag-7991/cloud-guard/internal/pricing"
	"github.com/singh-anurag-7991/cloud-guard/internal/storage"
	"github.com/singh-anurag-7991/cloud-guard/internal/version"
)

// ──────────────────────────────────────────────────────────────
// Health Check
// ──────────────────────────────────────────────────────────────

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	// commit/built let you confirm which build is actually live from a browser,
	// instead of guessing whether a deploy finished.
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "cloud-guard",
		"commit":  version.Commit,
		"built":   version.BuildTime,
	})
}

// ──────────────────────────────────────────────────────────────
// Landing Page (HTML) — Public marketing & portfolio showcase
// ──────────────────────────────────────────────────────────────

func (s *Server) handleLandingPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	s.renderTemplate(w, "landing.html", nil)
}

// ──────────────────────────────────────────────────────────────
// Products Page (HTML) — Deep-dive into Guard Platform products
// ──────────────────────────────────────────────────────────────

func (s *Server) handleProducts(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "products.html", nil)
}

// ──────────────────────────────────────────────────────────────
// About Page (HTML) — Engineer portfolio & contact
// ──────────────────────────────────────────────────────────────

func (s *Server) handleAbout(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "about.html", nil)
}

// ──────────────────────────────────────────────────────────────
// Dashboard (HTML) — serves the main page with findings
// ──────────────────────────────────────────────────────────────

// dashboardData is the data passed to the HTML template.
type dashboardData struct {
	Error              string
	Success            string
	TenantID           string
	TotalFindings      int
	HighCount          int
	MediumCount        int
	LowCount           int
	InfoCount          int
	AccountsCount      int
	EstimatedSavings   string
	Findings           []models.Finding
	Accounts           []storage.AccountRecord
	CloudFormationURL string

	// SavingsFindings is the subset of findings that carry a real dollar figure,
	// already sorted by size. These are what a customer came here for.
	SavingsFindings []models.Finding

	// HasPricedFindings distinguishes "we found nothing to save" from
	// "we have not priced anything yet" - the template must not show $0 as
	// though it were a confident result.
	HasPricedFindings bool

	// AnnualSavings is the monthly figure x12, shown because a $38/mo line item
	// reads as noise while $456/yr reads as a decision.
	AnnualSavings string

	// PricingNote states where the numbers come from and what they exclude.
	// Every savings figure on screen needs this next to it.
	PricingNote string

	// Real scan telemetry, so a clean scan with zero risks still shows evidence
	// that it actually inspected the account.
	ResourcesScanned int
	LastScanAt       string
	ScannerResults   []storage.ScannerResult
}

// buildDashboardData assembles the full dashboard state for a tenant.
//
// This exists because the error paths used to construct dashboardData by hand
// with only Error/TenantID/Findings set, which rendered accounts and every
// counter as 0 - making a failed scan look like the data had been deleted.
func (s *Server) buildDashboardData(tenantID string) dashboardData {
	findings, err := s.DB.GetLatestFindingsByTenant(tenantID)
	if err != nil {
		log.Printf("Error fetching findings for tenant %s: %v", tenantID, err)
	}
	accounts, _ := s.DB.ListAccountsByTenant(tenantID)

	var high, med, low, info int
	var monthlyTotal float64
	var priced []models.Finding
	for _, f := range findings {
		switch f.RiskLevel {
		case "HIGH":
			high++
		case "MEDIUM":
			med++
		case "LOW":
			low++
		default:
			info++
		}
		if f.MonthlySavingUSD > 0 {
			monthlyTotal += f.MonthlySavingUSD
			priced = append(priced, f)
		}
	}

	// The savings total is now the sum of individually priced findings, each of
	// which names a resource the customer can go and look at. The previous
	// formula - (highCount*150 + mediumCount*100) - produced a confident number
	// with no relationship to the customer's bill. A customer who checks that
	// number once and finds it invented never trusts the product again.
	savings := "$0"
	annual := "$0"
	if monthlyTotal > 0 {
		savings = fmt.Sprintf("$%.2f/mo", monthlyTotal)
		annual = fmt.Sprintf("$%.0f/yr", monthlyTotal*12)
	}

	d := dashboardData{
		TenantID:          tenantID,
		Findings:          findings,
		Accounts:          accounts,
		TotalFindings:     len(findings),
		HighCount:         high,
		MediumCount:       med,
		LowCount:          low,
		InfoCount:         info,
		AccountsCount:     len(accounts),
		EstimatedSavings:  savings,
		AnnualSavings:     annual,
		SavingsFindings:   priced,
		HasPricedFindings: len(priced) > 0,
		PricingNote: fmt.Sprintf(
			"Estimates use AWS %s on-demand list prices (captured %s). Savings Plans, Reserved Instances and volume discounts are not applied, so your actual saving may differ.",
			pricing.Region, pricing.SourceDate.Format("2 Jan 2006")),
		CloudFormationURL: cloudFormationLaunchURL(),
	}

	if sum, err := s.DB.GetLatestScanSummary(tenantID); err == nil && sum != nil {
		d.ScannerResults = sum.Results
		d.ResourcesScanned = sum.TotalResourcesScanned()
		d.LastScanAt = sum.RunAt.Format("2 Jan 2006, 15:04 MST")
	}
	return d
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Only serve exact "/dashboard" path
	if r.URL.Path != "/dashboard" {
		http.NotFound(w, r)
		return
	}

	tenantID := auth.GetTenantID(r.Context())
	data := s.buildDashboardData(tenantID)

	switch r.URL.Query().Get("success") {
	case "account_connected":
		data.Success = "AWS account connected and verified."
	case "scan_complete":
		data.Success = "Scan completed."
	case "account_removed":
		data.Success = "Account removed."
	}

	s.renderTemplate(w, "index.html", data)
}

// ──────────────────────────────────────────────────────────────
// Connect AWS Account (HTML form POST)
// ──────────────────────────────────────────────────────────────

func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.GetTenantID(r.Context())
	roleARN := strings.TrimSpace(r.FormValue("role_arn"))

	// fail renders the *full* dashboard plus the error, so the page never looks
	// like the user's connected accounts vanished.
	fail := func(msg string) {
		d := s.buildDashboardData(tenantID)
		d.Error = msg
		s.renderTemplate(w, "index.html", d)
	}

	if roleARN == "" {
		fail("Role ARN is required.")
		return
	}

	if strings.HasPrefix(roleARN, "arn:aws:cloudformation:") {
		fail("That's the CloudFormation stack ARN, not the role ARN. Open your stack in AWS, go to the Outputs tab, and copy the RoleARN value.")
		return
	}

	// Actually assume the role before saving it. Previously any well-formed ARN was
	// accepted and the failure only surfaced later as an opaque scan error.
	if _, err := cloudguardaws.ValidateRole(r.Context(), roleARN); err != nil {
		fail(err.Error())
		return
	}

	if _, err := s.DB.AddAccountForTenant(tenantID, roleARN); err != nil {
		log.Printf("Error adding account for tenant %s: %v", tenantID, err)
		fail("Failed to save the account: " + err.Error())
		return
	}

	http.Redirect(w, r, "/dashboard?success=account_connected", http.StatusSeeOther)
}

// handleDisconnect removes a connected AWS account and its scan history.
func (s *Server) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.GetTenantID(r.Context())

	id, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("account_id")), 10, 64)
	if err != nil {
		d := s.buildDashboardData(tenantID)
		d.Error = "Invalid account id."
		s.renderTemplate(w, "index.html", d)
		return
	}

	if err := s.DB.DeleteAccountForTenant(tenantID, id); err != nil {
		d := s.buildDashboardData(tenantID)
		d.Error = "Could not remove account: " + err.Error()
		s.renderTemplate(w, "index.html", d)
		return
	}

	http.Redirect(w, r, "/dashboard?success=account_removed", http.StatusSeeOther)
}

// ──────────────────────────────────────────────────────────────
// Manual Scan (HTML form POST)
// ──────────────────────────────────────────────────────────────

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.GetTenantID(r.Context())
	err := s.Orchestrator.ScanAllForTenant(r.Context(), tenantID)
	if err != nil {
		log.Printf("Scan failed for tenant %s: %v", tenantID, err)
		d := s.buildDashboardData(tenantID)
		d.Error = "Scan failed: " + err.Error()
		s.renderTemplate(w, "index.html", d)
		return
	}

	http.Redirect(w, r, "/dashboard?success=scan_complete", http.StatusSeeOther)
}

// ──────────────────────────────────────────────────────────────
// JSON APIs
// ──────────────────────────────────────────────────────────────

func (s *Server) handleAPIFindings(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.GetTenantID(r.Context())
	findings, err := s.DB.GetLatestFindingsByTenant(tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tenant_id": tenantID,
		"findings":  findings,
		"total":     len(findings),
	})
}

func (s *Server) handleAPIAccounts(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.GetTenantID(r.Context())
	accounts, err := s.DB.ListAccountsByTenant(tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tenant_id": tenantID,
		"accounts":  accounts,
		"total":     len(accounts),
	})
}

func (s *Server) handleAPIScan(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.GetTenantID(r.Context())
	err := s.Orchestrator.ScanAllForTenant(r.Context(), tenantID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"tenant_id": tenantID,
			"status":    "failed",
			"error":     err.Error(),
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"tenant_id": tenantID,
		"status":    "completed",
	})
}

// cloudFormationLaunchURL builds the AWS console quick-create link.
//
// CloudFormation only accepts an S3-hosted templateURL, so a true 1-click link
// requires the template to be uploaded to a public S3 bucket and that URL set in
// CF_TEMPLATE_S3_URL. Without it we fall back to the plain create-stack page,
// where the customer uploads the downloaded YAML themselves (2-click).
func cloudFormationLaunchURL() string {
	s3URL := strings.TrimSpace(os.Getenv("CF_TEMPLATE_S3_URL"))
	if s3URL == "" {
		return "https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/create"
	}
	return "https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/create/review" +
		"?templateURL=" + url.QueryEscape(s3URL) +
		"&stackName=CloudGuardReadOnlyRole" +
		"&param_ExternalId=" + url.QueryEscape(cloudguardaws.ExternalID()) +
		"&param_SaaSAccountID=" + url.QueryEscape(saasAccountID())
}

// saasAccountID is the AWS account that assumes customer roles. Must match the
// SaaSAccountID default in deployments/cloudformation.yaml.
func saasAccountID() string {
	if v := strings.TrimSpace(os.Getenv("CLOUDGUARD_SAAS_ACCOUNT_ID")); v != "" {
		return v
	}
	return "143506099819"
}

// handleCloudFormationURL generates CloudFormation launch URL for AWS onboarding
func (s *Server) handleCloudFormationURL(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.GetTenantID(r.Context())

	writeJSON(w, http.StatusOK, map[string]string{
		"tenant_id":          tenantID,
		"cloudformation_url": cloudFormationLaunchURL(),
		"template_url":       "/cloudformation.yaml",
		"external_id":        cloudguardaws.ExternalID(),
		"saas_account_id":    saasAccountID(),
	})
}

// handleServeCloudFormationYAML serves the CloudFormation template file
func (s *Server) handleServeCloudFormationYAML(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/yaml")
	http.ServeFile(w, r, "deployments/cloudformation.yaml")
}

// ──────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────

// renderTemplate renders an HTML template with the given data.
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := s.Templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON: %v", err)
	}
}
