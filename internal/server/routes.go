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

	// ── Public pages ───────────────────────────────────────────
	// "/" is the personal portfolio; the Cloud Guard marketing page moved to
	// /cloud-guard and is reached from the portfolio's products box.
	s.Router.HandleFunc("GET /", s.handlePortfolio)
	s.Router.HandleFunc("GET /cloud-guard", s.handleCloudGuardMarketing)
	s.Router.HandleFunc("GET /products", s.handleProducts)
	s.Router.HandleFunc("GET /about", s.handleAbout)

	// Static assets (portrait, any future images).
	// max-age=86400 rather than immutable: the filename is not content-hashed,
	// so a year-long cache would strip the ability to ever replace the photo.
	staticFS := http.FileServer(http.Dir("web/static"))
	s.Router.Handle("GET /static/", http.StripPrefix("/static/",
		cacheControl(staticFS, "public, max-age=86400")))

	// ── Auth (public) ──────────────────────────────────────────
	s.Router.HandleFunc("GET /login", s.handleLoginPage)
	s.Router.HandleFunc("POST /login", s.handleLoginSubmit)
	s.Router.HandleFunc("GET /signup", s.handleSignupPage)
	s.Router.HandleFunc("POST /signup", s.handleSignupSubmit)
	s.Router.HandleFunc("POST /logout", s.handleLogout)

	// Password reset. Public by necessity — someone locked out cannot be asked
	// to sign in first.
	s.Router.HandleFunc("GET /forgot-password", s.handleForgotPasswordPage)
	s.Router.HandleFunc("POST /forgot-password", s.handleForgotPasswordSubmit)
	s.Router.HandleFunc("GET /reset-password", s.handleResetPasswordPage)
	s.Router.HandleFunc("POST /reset-password", s.handleResetPasswordSubmit)

	// Protected HTML routes
	s.Router.Handle("GET /dashboard", protect(s.handleDashboard))
	s.Router.Handle("POST /connect", protect(s.handleConnect))
	s.Router.Handle("POST /disconnect", protect(s.handleDisconnect))
	s.Router.Handle("POST /scan", protect(s.handleScan))
	s.Router.Handle("GET /findings.csv", protect(s.handleExportCSV))

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

// cacheControl sets a Cache-Control header on a handler's responses.
func cacheControl(h http.Handler, value string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", value)
		h.ServeHTTP(w, r)
	})
}
