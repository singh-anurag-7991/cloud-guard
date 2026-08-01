package server

import (
	"net/http"

	"github.com/singh-anurag-7991/cloud-guard/internal/auth"
)

// registerRoutes sets up all HTTP routes on the server's mux.
func (s *Server) registerRoutes() {
	// protect wraps a handler so only authenticated sessions reach it.
	protect := func(h http.HandlerFunc) http.Handler {
		return auth.RequireAuth(s.DB, h)
	}

	// Health check — load balancer / monitoring ke liye
	s.Router.HandleFunc("GET /healthz", s.handleHealthz)

	// Public Pages — Marketing & Portfolio
	s.Router.HandleFunc("GET /", s.handleLandingPage)
	s.Router.HandleFunc("GET /products", s.handleProducts)
	s.Router.HandleFunc("GET /about", s.handleAbout)

	// ── Auth (public) ──────────────────────────────────────────
	s.Router.HandleFunc("GET /login", s.handleLoginPage)
	s.Router.HandleFunc("POST /login", s.handleLoginSubmit)
	s.Router.HandleFunc("GET /signup", s.handleSignupPage)
	s.Router.HandleFunc("POST /signup", s.handleSignupSubmit)
	s.Router.HandleFunc("POST /logout", s.handleLogout)

	// Protected HTML routes
	s.Router.Handle("GET /dashboard", protect(s.handleDashboard))
	s.Router.Handle("POST /connect", protect(s.handleConnect))
	s.Router.Handle("POST /disconnect", protect(s.handleDisconnect))
	s.Router.Handle("POST /scan", protect(s.handleScan))

	// ── JSON APIs (Protected) ──────────────────────────────────
	s.Router.Handle("GET /api/findings", protect(s.handleAPIFindings))
	s.Router.Handle("GET /api/accounts", protect(s.handleAPIAccounts))
	s.Router.Handle("POST /api/scan", protect(s.handleAPIScan))
	s.Router.Handle("GET /api/accounts/cloudformation-url", protect(s.handleCloudFormationURL))
	s.Router.HandleFunc("GET /cloudformation.yaml", s.handleServeCloudFormationYAML)

	// ── Billing APIs ───────────────────────────────────────────
	s.Router.Handle("POST /api/billing/checkout", protect(s.Billing.HandleCheckout))
	s.Router.Handle("POST /api/billing/portal", protect(s.Billing.HandlePortal))
	// Stripe webhooks are server-to-server and authenticated by signature, not session.
	s.Router.HandleFunc("POST /api/webhooks/stripe", s.Billing.HandleWebhook)
}
