package server

import (
	"net/http"
	"strings"
)

type hubData struct {
	Products []ProductDoc
}

// handleHub is where signing in lands. It used to be the Cloud Guard dashboard,
// which meant one product's UI was the front door to a platform holding three —
// and a visitor who wanted to know what Cloud Guard even was got a table of
// findings before an explanation.
func (s *Server) handleHub(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "hub.html", hubData{Products: productDocs()})
}

// handleProductDoc serves /hub/{slug}.
func (s *Server) handleProductDoc(w http.ResponseWriter, r *http.Request) {
	// PathValue rather than trimming the path by hand: the mux has already
	// decoded the segment, so "%2e%2e" cannot sneak through as "..".
	doc, ok := productDoc(strings.ToLower(r.PathValue("slug")))
	if !ok {
		// A wrong slug must 404 rather than fall through to the hub. Silently
		// showing a different page turns a broken link into an invisible bug.
		http.NotFound(w, r)
		return
	}

	s.renderTemplate(w, "product_doc.html", doc)
}
