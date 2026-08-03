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
type Product struct {
	Name        string
	Tagline     string
	Description string
	Stack       string
	Status      string // "live" | "soon"
	URL         string
	Icon        string
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
			URL:         "/login",
			Icon:        "🛡️",
		},
		{
			Name:        "Data Guard",
			Tagline:     "Data quality and pipeline monitoring",
			Description: "In design. It will watch data pipelines the way Cloud Guard watches infrastructure cost — schema drift, freshness, row-count anomalies.",
			Stack:       "Planned: Go · Spark · Iceberg",
			// Marked "soon" rather than shipped with a broken UI. A product box
			// that opens onto something non-functional is worse than an honest
			// placeholder - the visitor concludes the working product is fake too.
			Status: "soon",
			URL:    "/products",
			Icon:   "📊",
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
