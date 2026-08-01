package server

import (
	"log"
	"net/http"
	"time"

	"github.com/singh-anurag-7991/cloud-guard/internal/auth"
)

// LoggingMiddleware logs details of each incoming HTTP request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		tenantID := auth.GetTenantID(r.Context())

		ww := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)

		duration := time.Since(start)
		log.Printf("[HTTP] %s %s | Status: %d | Duration: %v | Tenant: %s",
			r.Method, r.URL.Path, ww.statusCode, duration, tenantID)
	})
}

// RecoveryMiddleware catches panics and returns a 500 error instead of crashing the server.
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[PANIC RECOVERED] %v", err)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
