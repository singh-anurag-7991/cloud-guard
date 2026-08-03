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
	s.Router.HandleFunc("GET /work", s.handleWork)
	s.Router.HandleFunc("GET /journey", s.handleJourney)
	s.Router.HandleFunc("GET /contact", s.handleContact)
	// GuardInfra: the platform overview, then one documentation page per
	// product. Each product page is the route into that product's dashboard.
	s.Router.HandleFunc("GET /products", s.handleProducts)
	s.Router.HandleFunc("GET /products/cloud-guard", s.handleDocCloudGuard)
	s.Router.HandleFunc("GET /products/shield", s.handleDocShield)
	s.Router.HandleFunc("GET /products/data-guard", s.handleDocDataGuard)
	s.Router.HandleFunc("GET /cloud-guard", s.handleCloudGuardMarketing)
	// /about predates the portfolio and its content now lives on /journey.
	// Redirecting rather than deleting keeps any link already shared working.
	s.Router.HandleFunc("GET /about", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/journey", http.StatusMovedPermanently)
	})

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

	// Protected HTML routes
	s.Router.Handle("GET /dashboard", protect(s.handleDashboard))
	s.Router.Handle("POST /connect", protect(s.handleConnect))
	s.Router.Handle("POST /disconnect", protect(s.handleDisconnect))
	s.Router.Handle("POST /scan", protect(s.handleScan))
	s.Router.Handle("GET /findings.csv", protect(s.handleExportCSV))

	// ── Auth API for the Next.js front end ─────────────────────
	// Public by design: these are how a session begins. They share the same
	// users, sessions and tenant isolation as the HTML handlers above.
	s.Router.HandleFunc("POST /api/auth/login", s.handleAPILogin)
	s.Router.HandleFunc("POST /api/auth/signup", s.handleAPISignup)
	s.Router.HandleFunc("POST /api/auth/logout", s.handleAPILogout)
	s.Router.HandleFunc("GET /api/auth/session", s.handleAPISession)

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
