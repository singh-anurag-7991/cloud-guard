package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/singh-anurag-7991/cloud-guard/internal/auth"
	"github.com/singh-anurag-7991/cloud-guard/internal/storage"
)

type authPageData struct {
	Error          string
	Success        string
	Email          string
	SignupsEnabled bool
}

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	// Already signed in - skip the form.
	if _, _, ok := s.DB.LookupSession(auth.SessionToken(r)); ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	data := authPageData{SignupsEnabled: auth.SignupsEnabled()}
	if r.URL.Query().Get("registered") == "1" {
		data.Success = "Account created. Please sign in."
	}
	s.renderTemplate(w, "login.html", data)
}

func (s *Server) handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	email := storage.NormalizeEmail(r.FormValue("email"))
	password := r.FormValue("password")

	fail := func(msg string) {
		// Deliberately vague: never reveal whether the email exists.
		s.renderTemplate(w, "login.html", authPageData{
			Error: msg, Email: email, SignupsEnabled: auth.SignupsEnabled(),
		})
	}

	if email == "" || password == "" {
		fail("Email and password are required.")
		return
	}

	user, err := s.DB.GetUserByEmail(email)
	if err != nil || !auth.CheckPassword(user.PasswordHash, password) {
		fail("Incorrect email or password.")
		return
	}

	token, err := s.DB.CreateSession(user.ID, user.TenantID, auth.SessionTTL)
	if err != nil {
		log.Printf("session create failed: %v", err)
		fail("Could not start session. Please try again.")
		return
	}

	auth.SetSessionCookie(w, r, token)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (s *Server) handleSignupPage(w http.ResponseWriter, r *http.Request) {
	if !auth.SignupsEnabled() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if _, _, ok := s.DB.LookupSession(auth.SessionToken(r)); ok {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	s.renderTemplate(w, "signup.html", authPageData{SignupsEnabled: true})
}

func (s *Server) handleSignupSubmit(w http.ResponseWriter, r *http.Request) {
	if !auth.SignupsEnabled() {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email := storage.NormalizeEmail(r.FormValue("email"))
	password := r.FormValue("password")
	confirm := r.FormValue("confirm_password")

	fail := func(msg string) {
		s.renderTemplate(w, "signup.html", authPageData{
			Error: msg, Email: email, SignupsEnabled: true,
		})
	}

	if email == "" || !strings.Contains(email, "@") {
		fail("Please enter a valid email address.")
		return
	}
	if len(password) < 10 {
		fail("Password must be at least 10 characters.")
		return
	}
	if password != confirm {
		fail("Passwords do not match.")
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Printf("hash failed: %v", err)
		fail("Could not create account. Please try again.")
		return
	}

	if _, err := s.DB.CreateUser(email, hash); err != nil {
		if err == storage.ErrEmailTaken {
			fail("That email is already registered. Try signing in instead.")
			return
		}
		log.Printf("create user failed: %v", err)
		fail("Could not create account. Please try again.")
		return
	}

	http.Redirect(w, r, "/login?registered=1", http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.DB.DeleteSession(auth.SessionToken(r))
	auth.ClearSessionCookie(w, r)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}
