package server

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/auth"
	"github.com/singh-anurag-7991/cloud-guard/internal/storage"
)

// Password reset.
//
// The whole flow is built so that an unauthenticated visitor learns nothing
// about who has an account here. Every response on the request page is
// identical whether the address exists or not.

// maxResetsPerHour stops the form being used to flood somebody's inbox — either
// as harassment or to bury a real security notice under noise.
const maxResetsPerHour = 3

type resetPageData struct {
	Error   string
	Success string
	Token   string
	// SMTPMissing is true when mail cannot be delivered. Shown only to make a
	// misconfigured deployment obvious to its operator rather than leaving
	// users waiting for an email that was never sent.
	SMTPMissing bool
}

func (s *Server) handleForgotPasswordPage(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, "forgot_password.html", resetPageData{
		SMTPMissing: !s.Email.Configured(),
	})
}

// handleForgotPasswordSubmit issues a reset link.
func (s *Server) handleForgotPasswordSubmit(w http.ResponseWriter, r *http.Request) {
	email := storage.NormalizeEmail(r.FormValue("email"))

	// One message for every outcome — unknown address, rate-limited, mail
	// failure. Saying "no account found" would turn this form into a way to
	// test which email addresses are registered.
	done := func() {
		s.renderTemplate(w, "forgot_password.html", resetPageData{
			Success: "If an account exists for that address, a reset link is on its way. " +
				"The link works once and expires in an hour.",
			SMTPMissing: !s.Email.Configured(),
		})
	}

	if email == "" || !strings.Contains(email, "@") {
		s.renderTemplate(w, "forgot_password.html", resetPageData{
			Error: "That does not look like an email address. Check for a typo.",
		})
		return
	}

	user, err := s.DB.GetUserByEmail(email)
	if err != nil {
		// No such user. Same response, same timing shape as the success path.
		done()
		return
	}

	if n, err := s.DB.RecentResetRequestCount(user.ID, time.Hour); err == nil && n >= maxResetsPerHour {
		log.Printf("password reset: rate limited for user %d", user.ID)
		done()
		return
	}

	token, err := s.DB.CreatePasswordReset(user.ID)
	if err != nil {
		log.Printf("password reset: could not create token for user %d: %v", user.ID, err)
		done()
		return
	}

	link := siteOrigin(r) + "/reset-password?token=" + token
	body := "Someone asked to reset the Cloud Guard password for this address.\n\n" +
		"Open this link to choose a new one. It works once and expires in an hour:\n\n" +
		link + "\n\n" +
		"If this was not you, nothing has changed and you can ignore this message.\n" +
		"Your password stays as it is until someone opens that link.\n"

	if err := s.Email.Send(user.Email, "Reset your Cloud Guard password", body); err != nil {
		// Already logged, including the link itself when SMTP is unconfigured.
		log.Printf("password reset: delivery failed for user %d: %v", user.ID, err)
	}

	done()
}

// handleResetPasswordPage shows the new-password form for a valid token.
func (s *Server) handleResetPasswordPage(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")

	if _, ok := s.DB.LookupPasswordReset(token); !ok {
		// Deliberately one message for expired, used and never-existed. Telling
		// them apart confirms to a guesser when they have hit a real token.
		s.renderTemplate(w, "reset_password.html", resetPageData{
			Error: "This link is no longer valid. Reset links work once and expire after an hour — request a new one.",
		})
		return
	}

	s.renderTemplate(w, "reset_password.html", resetPageData{Token: token})
}

// handleResetPasswordSubmit sets the new password.
func (s *Server) handleResetPasswordSubmit(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")
	password := r.FormValue("password")
	confirm := r.FormValue("confirm_password")

	fail := func(msg string) {
		s.renderTemplate(w, "reset_password.html", resetPageData{Error: msg, Token: token})
	}

	if _, ok := s.DB.LookupPasswordReset(token); !ok {
		s.renderTemplate(w, "reset_password.html", resetPageData{
			Error: "This link is no longer valid. Reset links work once and expire after an hour — request a new one.",
		})
		return
	}

	// Length over composition rules, same as signup. Forced symbols push people
	// toward "Password1!" while a long passphrase is stronger and memorable.
	if len(password) < 10 {
		fail("That password is under 10 characters. Add a few more — a short phrase works well.")
		return
	}
	if password != confirm {
		fail("The two passwords do not match.")
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		log.Printf("password reset: hash failed: %v", err)
		fail("Could not set that password. Please try again.")
		return
	}

	// This also burns the token and signs out every existing session.
	if err := s.DB.ConsumePasswordReset(token, hash); err != nil {
		log.Printf("password reset: consume failed: %v", err)
		fail("Could not set that password. The link may have just expired — request a new one.")
		return
	}

	// Straight to login rather than signing them in automatically: proving the
	// new password works right now is worth one extra step, and it confirms to
	// the user that the change actually took effect.
	http.Redirect(w, r, "/login?reset=1", http.StatusSeeOther)
}

// siteOrigin reconstructs the public URL of this deployment.
//
// Caddy terminates TLS and proxies to plain HTTP, so r.TLS is always nil here
// and scheme has to come from the forwarded header. Getting this wrong produces
// an http:// reset link that a browser may refuse to send the cookie over.
func siteOrigin(r *http.Request) string {
	if v := strings.TrimSpace(os.Getenv("CLOUDGUARD_SITE_ORIGIN")); v != "" {
		return strings.TrimSuffix(v, "/")
	}

	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}

	host := r.Header.Get("X-Forwarded-Host")
	if host == "" {
		host = r.Host
	}
	return scheme + "://" + host
}
