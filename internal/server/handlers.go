package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/singh-anurag-7991/cloud-guard/internal/auth"
	cloudguardaws "github.com/singh-anurag-7991/cloud-guard/internal/aws"
	"github.com/singh-anurag-7991/cloud-guard/internal/models"
	"github.com/singh-anurag-7991/cloud-guard/internal/storage"
)

// ──────────────────────────────────────────────────────────────
// Health Check
// ──────────────────────────────────────────────────────────────

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"service": "cloud-guard",
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
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Only serve exact "/dashboard" path
	if r.URL.Path != "/dashboard" {
		http.NotFound(w, r)
		return
	}

	tenantID := auth.GetTenantID(r.Context())
	findings, err := s.DB.GetLatestFindingsByTenant(tenantID)
	accounts, _ := s.DB.ListAccountsByTenant(tenantID)

	var high, med, low, info int
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
	}

	cfURL := cloudFormationLaunchURL()

	// Calculate estimated savings ($150 per stopped/idle EC2, $200 per oversized RDS)
	savingsEst := (high * 150) + (med * 100)
	savingsStr := "$0"
	if savingsEst > 0 {
		savingsStr = fmt.Sprintf("$%d/mo", savingsEst)
	}

	data := dashboardData{
		TenantID:           tenantID,
		Findings:           findings,
		Accounts:           accounts,
		TotalFindings:      len(findings),
		HighCount:          high,
		MediumCount:        med,
		LowCount:           low,
		InfoCount:          info,
		AccountsCount:      len(accounts),
		EstimatedSavings:   savingsStr,
		CloudFormationURL: cfURL,
	}

	if err != nil {
		log.Printf("Error fetching findings for tenant %s: %v", tenantID, err)
		data.Error = "Could not load findings."
	}

	// Handle success query param
	if r.URL.Query().Get("success") == "account_connected" {
		data.Success = "AWS Account connected successfully!"
	} else if r.URL.Query().Get("success") == "scan_complete" {
		data.Success = "Scan completed successfully!"
	}

	s.renderTemplate(w, "index.html", data)
}

// ──────────────────────────────────────────────────────────────
// Connect AWS Account (HTML form POST)
// ──────────────────────────────────────────────────────────────

func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.GetTenantID(r.Context())
	roleARN := strings.TrimSpace(r.FormValue("role_arn"))
	if roleARN == "" {
		s.renderTemplate(w, "index.html", dashboardData{Error: "Role ARN is required.", TenantID: tenantID})
		return
	}

	// Validate it is an IAM *role* ARN. The commonest mistake is pasting the
	// CloudFormation stack ARN (arn:aws:cloudformation:...:stack/...) instead of the
	// RoleARN from the stack's Outputs tab, so call that out explicitly.
	if !strings.HasPrefix(roleARN, "arn:aws:iam::") || !strings.Contains(roleARN, ":role/") {
		msg := "That doesn't look like an IAM role ARN. It should look like " +
			"arn:aws:iam::123456789012:role/CloudGuardReadOnlyRole-us-east-1"
		if strings.HasPrefix(roleARN, "arn:aws:cloudformation:") {
			msg = "That's the CloudFormation stack ARN, not the role ARN. " +
				"Open your stack in AWS, go to the Outputs tab, and copy the RoleARN value."
		}
		s.renderTemplate(w, "index.html", dashboardData{Error: msg, TenantID: tenantID})
		return
	}

	_, err := s.DB.AddAccountForTenant(tenantID, roleARN)
	if err != nil {
		log.Printf("Error adding account for tenant %s: %v", tenantID, err)
		s.renderTemplate(w, "index.html", dashboardData{Error: "Failed to add account: " + err.Error(), TenantID: tenantID})
		return
	}

	// Redirect back to dashboard with success
	http.Redirect(w, r, "/dashboard?success=account_connected", http.StatusSeeOther)
}

// ──────────────────────────────────────────────────────────────
// Manual Scan (HTML form POST)
// ──────────────────────────────────────────────────────────────

func (s *Server) handleScan(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.GetTenantID(r.Context())
	err := s.Orchestrator.ScanAllForTenant(r.Context(), tenantID)
	if err != nil {
		log.Printf("Scan failed for tenant %s: %v", tenantID, err)
		findings, _ := s.DB.GetLatestFindingsByTenant(tenantID)
		s.renderTemplate(w, "index.html", dashboardData{
			Error:    "Scan failed: " + err.Error(),
			TenantID: tenantID,
			Findings: findings,
		})
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
