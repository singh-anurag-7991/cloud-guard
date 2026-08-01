package server

import (
	"net/http"

	"github.com/singh-anurag-7991/cloud-guard/internal/auth"
)

// registerRoutes sets up all HTTP routes on the server's mux.
func (s *Server) registerRoutes() {
	// Health check — load balancer / monitoring ke liye
	s.Router.HandleFunc("GET /healthz", s.handleHealthz)

	// Public Pages — Marketing & Portfolio
	s.Router.HandleFunc("GET /", s.handleLandingPage)
	s.Router.HandleFunc("GET /products", s.handleProducts)
	s.Router.HandleFunc("GET /about", s.handleAbout)

	// Protected HTML routes
	s.Router.Handle("GET /dashboard", auth.RequireAuth(http.HandlerFunc(s.handleDashboard)))
	s.Router.Handle("POST /connect", auth.RequireAuth(http.HandlerFunc(s.handleConnect)))
	s.Router.Handle("POST /scan", auth.RequireAuth(http.HandlerFunc(s.handleScan)))

	// ── JSON APIs (Protected) ──────────────────────────────────
	s.Router.Handle("GET /api/findings", auth.RequireAuth(http.HandlerFunc(s.handleAPIFindings)))
	s.Router.Handle("GET /api/accounts", auth.RequireAuth(http.HandlerFunc(s.handleAPIAccounts)))
	s.Router.Handle("POST /api/scan", auth.RequireAuth(http.HandlerFunc(s.handleAPIScan)))
	s.Router.Handle("GET /api/accounts/cloudformation-url", auth.RequireAuth(http.HandlerFunc(s.handleCloudFormationURL)))
	s.Router.HandleFunc("GET /cloudformation.yaml", s.handleServeCloudFormationYAML)

	// ── Billing APIs ───────────────────────────────────────────
	s.Router.Handle("POST /api/billing/checkout", auth.RequireAuth(http.HandlerFunc(s.Billing.HandleCheckout)))
	s.Router.Handle("POST /api/billing/portal", auth.RequireAuth(http.HandlerFunc(s.Billing.HandlePortal)))
	s.Router.HandleFunc("POST /api/webhooks/stripe", s.Billing.HandleWebhook)
}
