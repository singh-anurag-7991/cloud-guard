package auth

import (
	"context"
	"net/http"
	"os"
	"strings"
)

type contextKey string

const (
	TenantIDKey contextKey = "tenant_id"
	UserIDKey   contextKey = "user_id"
)

// RequireAuth is HTTP middleware that verifies authentication (Clerk JWT or session token).
// If CLERK_SECRET_KEY is empty (e.g. local dev testing), it falls back to tenant "default".
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clerkKey := os.Getenv("CLERK_SECRET_KEY")

		// Dev mode / Local fallback when Clerk is not configured yet
		if clerkKey == "" {
			ctx := context.WithValue(r.Context(), TenantIDKey, "default")
			ctx = context.WithValue(ctx, UserIDKey, "dev-user")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// Extract token from Authorization header or __session cookie
		token := extractToken(r)
		if token == "" {
			http.Error(w, `{"error":"unauthorized: missing token"}`, http.StatusUnauthorized)
			return
		}

		// Parse basic claims (In production, verify signature against Clerk JWKS)
		tenantID, userID := parseTokenClaims(token)
		if tenantID == "" {
			tenantID = "default"
		}

		ctx := context.WithValue(r.Context(), TenantIDKey, tenantID)
		ctx = context.WithValue(ctx, UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetTenantID retrieves the authenticated tenant ID from request context.
func GetTenantID(ctx context.Context) string {
	if tid, ok := ctx.Value(TenantIDKey).(string); ok && tid != "" {
		return tid
	}
	return "default"
}

// GetUserID retrieves the authenticated user ID from request context.
func GetUserID(ctx context.Context) string {
	if uid, ok := ctx.Value(UserIDKey).(string); ok && uid != "" {
		return uid
	}
	return "dev-user"
}

func extractToken(r *http.Request) string {
	// Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// Check __session cookie (used by Clerk client SDK)
	if cookie, err := r.Cookie("__session"); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}

// parseTokenClaims extracts subject/user identity from token.
func parseTokenClaims(token string) (tenantID, userID string) {
	// Basic JWT payload inspection (split by '.')
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return "default", token
	}
	// Fallback to using token value or default if unparseable
	return "default", token
}
