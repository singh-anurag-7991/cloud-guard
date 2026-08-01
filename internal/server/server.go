package server

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/singh-anurag-7991/cloud-guard/internal/alerting"
	"github.com/singh-anurag-7991/cloud-guard/internal/billing"
	"github.com/singh-anurag-7991/cloud-guard/internal/orchestrator"
	"github.com/singh-anurag-7991/cloud-guard/internal/storage"
)

// Server holds all dependencies and the HTTP router.
type Server struct {
	DB           *storage.DB
	Orchestrator *orchestrator.Orchestrator
	Billing      *billing.Service
	Slack        *alerting.SlackWrapper
	Templates    *template.Template
	Router       *http.ServeMux
}

// New creates a Server with all dependencies wired up.
func New(db *storage.DB, slack *alerting.SlackWrapper) *Server {
	orch := orchestrator.New(db, slack)
	billSvc := billing.NewService(db)

	// Parse templates from web/ directory
	tmplPath := filepath.Join("web", "*.html")
	tmpl, err := template.ParseGlob(tmplPath)
	if err != nil {
		log.Printf("Warning: could not parse templates from %s: %v", tmplPath, err)
		tmpl = template.New("empty")
	}

	s := &Server{
		DB:           db,
		Orchestrator: orch,
		Billing:      billSvc,
		Slack:        slack,
		Templates:    tmpl,
		Router:       http.NewServeMux(),
	}

	s.registerRoutes()
	return s
}

// ServeHTTP makes Server implement http.Handler with logging & recovery middleware.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := RecoveryMiddleware(LoggingMiddleware(s.Router))
	handler.ServeHTTP(w, r)
}
