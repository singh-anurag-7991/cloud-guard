package server

import (
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/singh-anurag-7991/cloud-guard/internal/storage"
)

// Product is one entry in the portfolio's products box.
//
// Badge and CTA are carried per product rather than derived from Status,
// because the three products are reachable in genuinely different ways: one is
// a deployed service you can sign into, two are repositories you can read.
// Labelling a repository "Live" would imply a hosted app that does not exist.
type Product struct {
	Name        string
	Tagline     string
	Description string
	Stack       string
	Status      string // "live" | "oss" — drives the card's accent only
	Badge       string // text inside the pill
	PillClass   string
	CTA         string
	URL         string
	Icon        string
	// External sends the link to a new tab. A recruiter who clicks through to
	// GitHub should still have the portfolio open behind them.
	External bool
}

// products is the single source of truth for what is shown on the portfolio,
// on /products, and in the signed-in hub. Three copies of this list would drift
// within a week and end up advertising something that does not exist.
func products() []Product {
	return []Product{
		{
			Name:        "Cloud Guard",
			Tagline:     "Find the AWS spend you forgot about",
			Description: "Connects to an AWS account through a read-only IAM role and reports what is costing money for no reason — unattached EBS volumes, idle Elastic IPs, stale snapshots, gp2 volumes that should be gp3. Every finding is priced from AWS list prices and comes with the exact CLI command to fix it.",
			Stack:       "Go · SQLite · AWS SDK v2 · Docker · Caddy",
			Status:      "live",
			Badge:       "● Live",
			PillClass:   "pill-live",
			CTA:         "Open Cloud Guard →",
			// The product page, not the login form. Someone arriving from a CV
			// wants to know what this is before being asked for a password.
			URL:  "/cloud-guard",
			Icon: "🛡️",
		},
		{
			Name:    "Data Guard",
			Tagline: "A firewall for bad data",
			// This card said "in design" for weeks while the repository was
			// finished. Description now matches what is actually in it.
			Description: "Validates API payloads and SQL rows against declarative rules, pushing the checks down into SQL instead of pulling every row back. Alerts only when a check flips state, so the channel stays readable. Catches silent data failures before downstream logic acts on them.",
			Stack:       "Go · PostgreSQL · Next.js",
			Status:      "oss",
			Badge:       "Open source",
			PillClass:   "pill-oss",
			CTA:         "View on GitHub →",
			URL:         "https://github.com/singh-anurag-7991/data-guard",
			Icon:        "📊",
			External:    true,
		},
		{
			Name:        "Shield",
			Tagline:     "Rate limiting that holds under load",
			Description: "A Go rate limiter for APIs, with sliding-window and token-bucket strategies over either Redis or an in-memory store. Built to make the enforcement decision cheap enough to sit in the request path.",
			Stack:       "Go · Redis",
			Status:      "oss",
			Badge:       "Open source",
			PillClass:   "pill-oss",
			CTA:         "View on GitHub →",
			URL:         "https://github.com/singh-anurag-7991/shield",
			Icon:        "🧱",
			External:    true,
		},
	}
}

type portfolioData struct {
	Products []Product
	Visitors storage.PageViewStats
	PhotoURL string
}

// visitorSalt keeps hashed visitor IDs from being reversible via a rainbow
// table of common IP + user-agent pairs. Regenerated on restart when unset,
// which only means a returning visitor may be counted twice after a deploy.
var visitorSalt = func() string {
	if s := strings.TrimSpace(os.Getenv("CLOUDGUARD_VISITOR_SALT")); s != "" {
		return s
	}
	return "cg-visitor-salt-v1"
}()

// clientIP extracts the caller's address, honouring the proxy header.
//
// Caddy terminates TLS and forwards to the app, so r.RemoteAddr is always
// 127.0.0.1 here. Without reading X-Forwarded-For every visitor would hash to
// the same value and the counter would permanently read 1.
func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Left-most entry is the original client; the rest are proxies.
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}
	return r.RemoteAddr
}

// countVisit records the visit and returns the current stats.
// Counting must never break the page, so errors are logged and swallowed.
func (s *Server) countVisit(r *http.Request, path string) storage.PageViewStats {
	hash := storage.VisitorHash(clientIP(r), r.UserAgent(), visitorSalt)
	if err := s.DB.RecordPageView(path, hash); err != nil {
		log.Printf("page view record failed for %s: %v", path, err)
	}
	stats, err := s.DB.GetPageViewStats("")
	if err != nil {
		log.Printf("page view stats failed: %v", err)
	}
	return stats
}

// handlePortfolio serves the personal homepage at "/".
func (s *Server) handlePortfolio(w http.ResponseWriter, r *http.Request) {
	// http.ServeMux matches "GET /" as a catch-all, so an unknown URL would
	// otherwise render the homepage with a 200 and look like a working page.
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := portfolioData{
		Products: products(),
		Visitors: s.countVisit(r, "/"),
		PhotoURL: "/static/anurag-520.jpg",
	}

	s.renderTemplate(w, "portfolio.html", data)
}

// handleCloudGuardMarketing serves the old landing page, now reachable from the
// portfolio's product box rather than being the site's front door.
func (s *Server) handleCloudGuardMarketing(w http.ResponseWriter, r *http.Request) {
	s.countVisit(r, "/cloud-guard")
	s.renderTemplate(w, "landing.html", nil)
}
