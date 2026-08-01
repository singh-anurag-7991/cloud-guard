package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/singh-anurag-7991/cloud-guard/internal/auth"
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

	cfURL := "https://console.aws.amazon.com/cloudformation/home#/stacks/create/page"

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

	// Basic ARN format validation
	if !strings.HasPrefix(roleARN, "arn:aws:iam:") {
		s.renderTemplate(w, "index.html", dashboardData{Error: "Invalid ARN format. Must start with arn:aws:iam:", TenantID: tenantID})
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

// handleCloudFormationURL generates CloudFormation launch URL for AWS onboarding
func (s *Server) handleCloudFormationURL(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.GetTenantID(r.Context())
	templateURL := "http://localhost:8080/cloudformation.yaml"
	awsLaunchURL := "https://console.aws.amazon.com/cloudformation/home#/stacks/create/page"

	writeJSON(w, http.StatusOK, map[string]string{
		"tenant_id":             tenantID,
		"cloudformation_url":    awsLaunchURL,
		"template_url":          templateURL,
		"suggested_external_id": tenantID,
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
