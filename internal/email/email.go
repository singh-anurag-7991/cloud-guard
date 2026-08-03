// Package email sends the few transactional messages Cloud Guard needs.
//
// Plain net/smtp, no provider SDK. One message type, sent rarely — a dependency
// on a hosted email API would be more code to maintain than the thing it sends.
package email

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
)

// Sender delivers mail, or explains why it cannot.
type Sender struct {
	host     string
	port     string
	username string
	password string
	from     string
}

// NewSender reads SMTP settings from the environment.
//
// Everything optional on purpose: the app must start and serve without mail
// configured. What it must NOT do is pretend to have sent something.
func NewSender() *Sender {
	return &Sender{
		host:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		port:     firstNonEmpty(os.Getenv("SMTP_PORT"), "587"),
		username: strings.TrimSpace(os.Getenv("SMTP_USERNAME")),
		password: os.Getenv("SMTP_PASSWORD"),
		from:     firstNonEmpty(os.Getenv("SMTP_FROM"), "Cloud Guard <no-reply@guardinfra.duckdns.org>"),
	}
}

// Configured reports whether mail can actually be delivered.
//
// Callers use this to decide what to tell the user. A reset page that says
// "check your email" when no mail server exists sends people to wait for
// something that will never arrive.
func (s *Sender) Configured() bool {
	return s.host != "" && s.username != "" && s.password != ""
}

// Send delivers one message.
//
// When SMTP is not configured it logs the body instead and returns an error, so
// a single-operator deployment can still recover an account by reading the
// server log. It does not silently succeed — the caller has to decide what the
// user sees.
func (s *Sender) Send(to, subject, body string) error {
	if !s.Configured() {
		log.Printf(
			"[email] SMTP is not configured, so this was NOT delivered.\n"+
				"        To:      %s\n"+
				"        Subject: %s\n"+
				"        Body:\n%s",
			to, subject, indent(body),
		)
		return fmt.Errorf("smtp is not configured (set SMTP_HOST, SMTP_USERNAME, SMTP_PASSWORD)")
	}

	// CRLF line endings and a blank line between headers and body are required
	// by RFC 5322; plain \n makes some servers treat the headers as body text.
	msg := strings.Join([]string{
		"From: " + s.from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	auth := smtp.PlainAuth("", s.username, s.password, s.host)
	addr := s.host + ":" + s.port

	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("sending mail to %s: %w", to, err)
	}
	return nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if t := strings.TrimSpace(v); t != "" {
			return t
		}
	}
	return ""
}

func indent(s string) string {
	return "          " + strings.ReplaceAll(s, "\n", "\n          ")
}
