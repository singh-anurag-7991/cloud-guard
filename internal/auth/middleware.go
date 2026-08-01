package auth

import (
	"context"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type contextKey string

const (
	TenantIDKey contextKey = "tenant_id"
	UserIDKey   contextKey = "user_id"

	// SessionCookie is the cookie holding the opaque session token.
	SessionCookie = "cg_session"

	// SessionTTL is how long a login stays valid.
	SessionTTL = 7 * 24 * time.Hour
)

// SessionStore is the subset of storage.DB the auth layer needs.
type SessionStore interface {
	LookupSession(token string) (tenantID string, userID int64, ok bool)
}

// HashPassword returns a bcrypt hash suitable for storage.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword reports whether plain matches the stored bcrypt hash.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// SetSessionCookie writes the session cookie. secure is derived from the request
// so local http testing still works while production stays Secure-only.
func SetSessionCookie(w http.ResponseWriter, r *http.Request, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(SessionTTL),
	})
}

// ClearSessionCookie expires the session cookie on logout.
func ClearSessionCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// SessionToken reads the raw session token from the request.
func SessionToken(r *http.Request) string {
	if c, err := r.Cookie(SessionCookie); err == nil {
		return c.Value
	}
	return ""
}

// isHTTPS detects TLS directly or via the reverse proxy's X-Forwarded-Proto,
// which is what Caddy sets in front of this app.
func isHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return r.Header.Get("X-Forwarded-Proto") == "https"
}

// RequireAuth rejects unauthenticated requests. Browser requests are redirected
// to /login; API requests get a 401 JSON body.
//
// Previously an empty CLERK_SECRET_KEY silently granted everyone tenant
// "default", leaving the dashboard, /connect, /scan and every /api route open to
// the whole internet with no credentials.
func RequireAuth(store SessionStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID, userID, ok := store.LookupSession(SessionToken(r))
		if !ok {
			if wantsJSON(r) {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func wantsJSON(r *http.Request) bool {
	if r.Header.Get("Accept") == "application/json" {
		return true
	}
	return len(r.URL.Path) >= 5 && r.URL.Path[:5] == "/api/"
}

// GetTenantID retrieves the authenticated tenant ID from request context.
func GetTenantID(ctx context.Context) string {
	if tid, ok := ctx.Value(TenantIDKey).(string); ok && tid != "" {
		return tid
	}
	return ""
}

// GetUserID retrieves the authenticated user ID from request context.
func GetUserID(ctx context.Context) int64 {
	if uid, ok := ctx.Value(UserIDKey).(int64); ok {
		return uid
	}
	return 0
}

// SignupsEnabled lets you close public registration once you have customers,
// by setting DISABLE_SIGNUP=1.
func SignupsEnabled() bool {
	return os.Getenv("DISABLE_SIGNUP") == ""
}
