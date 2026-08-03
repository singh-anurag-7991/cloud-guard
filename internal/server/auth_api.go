package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/singh-anurag-7991/cloud-guard/internal/auth"
	"github.com/singh-anurag-7991/cloud-guard/internal/storage"
)

// Auth endpoints for the Next.js front end.
//
// These sit alongside the existing HTML handlers rather than replacing them:
// the Go-rendered pages keep working until the Next.js site is verified and
// switched over in Caddy, so a broken deploy is one config line from rollback.
//
// Both paths share the same bcrypt hashes, the same `sessions` table and the
// same per-user tenant isolation. There is one identity store, and it is this
// one. The difference is only in how failure is reported: HTML handlers
// re-render a page, these redirect back to the Next.js form with an error code
// the front end turns into copy.

// safeNext keeps ?next= from becoming an open redirect.
//
// Reflecting an arbitrary destination lets an attacker send a link to our own
// real login page and bounce the victim to a lookalike immediately after they
// authenticate. Only same-site absolute paths are honoured; "//evil.com" is a
// protocol-relative URL and must be rejected along with everything external.
func safeNext(raw, fallback string) string {
	if strings.HasPrefix(raw, "/") && !strings.HasPrefix(raw, "//") {
		return raw
	}
	return fallback
}

// authFail sends the user back to the form with a machine-readable reason.
//
// The code is a token, not a sentence — the wording lives in the Next.js page
// so the copy can change without a Go deploy, and so error text never gets
// duplicated in two languages of the stack.
func authFail(w http.ResponseWriter, r *http.Request, page, code, next string) {
	dest := page + "?error=" + code
	if next != "" && next != "/guard" {
		dest += "&next=" + next
	}
	http.Redirect(w, r, dest, http.StatusSeeOther)
}

// handleAPILogin authenticates and starts a session.
func (s *Server) handleAPILogin(w http.ResponseWriter, r *http.Request) {
	email := storage.NormalizeEmail(r.FormValue("email"))
	password := r.FormValue("password")
	next := safeNext(r.FormValue("next"), "/guard")

	if email == "" || password == "" {
		authFail(w, r, "/guard/login", "invalid", next)
		return
	}

	user, err := s.DB.GetUserByEmail(email)
	// One branch for "no such user" and "wrong password" on purpose. Splitting
	// them tells an attacker which email addresses have accounts here, which is
	// the first step of a credential-stuffing run.
	if err != nil || !auth.CheckPassword(user.PasswordHash, password) {
		authFail(w, r, "/guard/login", "invalid", next)
		return
	}

	token, err := s.DB.CreateSession(user.ID, user.TenantID, auth.SessionTTL)
	if err != nil {
		log.Printf("api login: session create failed: %v", err)
		authFail(w, r, "/guard/login", "invalid", next)
		return
	}

	auth.SetSessionCookie(w, r, token)
	http.Redirect(w, r, next, http.StatusSeeOther)
}

// handleAPISignup creates a user, its tenant, and a session in one step.
func (s *Server) handleAPISignup(w http.ResponseWriter, r *http.Request) {
	next := safeNext(r.FormValue("next"), "/guard")

	if !auth.SignupsEnabled() {
		http.Redirect(w, r, "/guard/login", http.StatusSeeOther)
		return
	}

	email := storage.NormalizeEmail(r.FormValue("email"))
	password := r.FormValue("password")

	if email == "" || !strings.Contains(email, "@") {
		authFail(w, r, "/guard/signup", "email", next)
		return
	}
	// Length over composition rules. Forced symbols push people toward
	// "Password1!" while a long passphrase is both stronger and memorable.
	if len(password) < 10 {
		authFail(w, r, "/guard/signup", "weak", next)
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Printf("api signup: hash failed: %v", err)
		authFail(w, r, "/guard/signup", "weak", next)
		return
	}

	// CreateUser generates the isolated tenant ID. Every later query is scoped
	// by it, so signing up *is* creating the organisation.
	user, err := s.DB.CreateUser(email, hash)
	if err != nil {
		if err == storage.ErrEmailTaken {
			authFail(w, r, "/guard/signup", "taken", next)
			return
		}
		log.Printf("api signup: create user failed: %v", err)
		authFail(w, r, "/guard/signup", "taken", next)
		return
	}

	token, err := s.DB.CreateSession(user.ID, user.TenantID, auth.SessionTTL)
	if err != nil {
		// The account exists now, so send them to log in rather than implying
		// the signup failed and inviting a duplicate attempt.
		log.Printf("api signup: session create failed: %v", err)
		http.Redirect(w, r, "/guard/login", http.StatusSeeOther)
		return
	}

	auth.SetSessionCookie(w, r, token)
	http.Redirect(w, r, next, http.StatusSeeOther)
}

// handleAPILogout ends the session on both sides.
func (s *Server) handleAPILogout(w http.ResponseWriter, r *http.Request) {
	// Delete the server record first. Clearing only the cookie would leave a
	// live token that still works if it was captured.
	s.DB.DeleteSession(auth.SessionToken(r))
	auth.ClearSessionCookie(w, r)
	http.Redirect(w, r, "/guard", http.StatusSeeOther)
}

// handleAPISession reports who is signed in.
//
// The Next.js server components call this to decide whether to render the
// account menu and the dashboard. It returns 200 with `authenticated: false`
// rather than 401 — "nobody is signed in" is a normal answer to this question,
// not an error, and a 401 would clutter logs on every anonymous page view.
func (s *Server) handleAPISession(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	// Note the order: LookupSession returns (tenantID, userID, ok). Getting
	// this backwards compiles fine — both are being destructured — and then
	// silently looks up a user by a tenant string.
	tenantID, userID, ok := s.DB.LookupSession(auth.SessionToken(r))
	if !ok {
		writeJSON(w, http.StatusOK, map[string]any{"authenticated": false})
		return
	}

	user, err := s.DB.GetUserByID(userID)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"authenticated": false})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"authenticated": true,
		"email":         user.Email,
		"tenantId":      tenantID,
	})
}
