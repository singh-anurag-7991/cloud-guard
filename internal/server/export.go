package server

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/auth"
)

// handleExportCSV streams the tenant's findings as a CSV download.
//
// A dashboard is where one person looks at the problem; a spreadsheet is how
// they get four other people to act on it. Without an export, every customer
// who wants to share these findings has to retype them, and most simply won't.
func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.GetTenantID(r.Context())

	findings, err := s.DB.GetLatestFindingsByTenant(tenantID)
	if err != nil {
		http.Error(w, "could not load findings", http.StatusInternalServerError)
		return
	}

	filename := fmt.Sprintf("cloud-guard-findings-%s.csv", time.Now().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.Header().Set("Cache-Control", "no-store")

	cw := csv.NewWriter(w)
	defer cw.Flush()

	_ = cw.Write([]string{
		"Severity", "Monthly Saving (USD)", "Annual Saving (USD)", "Resource ID",
		"Resource Type", "Region", "Description", "Evidence", "Recommendation",
		"Fix Command", "Confidence", "Rule", "Detected At",
	})

	for _, f := range findings {
		monthly := ""
		annual := ""
		if f.MonthlySavingUSD > 0 {
			monthly = strconv.FormatFloat(f.MonthlySavingUSD, 'f', 2, 64)
			annual = strconv.FormatFloat(f.MonthlySavingUSD*12, 'f', 2, 64)
		}
		_ = cw.Write([]string{
			f.RiskLevel, monthly, annual, f.ResourceID, string(f.ResourceType), f.Region,
			f.Description, f.Evidence, f.Recommendation, f.FixCommand, f.Confidence,
			f.RuleID, f.GeneratedAt.Format(time.RFC3339),
		})
	}
}
