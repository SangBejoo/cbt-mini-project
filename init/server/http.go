package server

import (
	"net/http"
	"strings"

	"cbt-test-mini-project/init/config"
)

// corsMiddleware adds CORS headers to all responses and handles OPTIONS preflight requests.
func corsMiddleware(cfg *config.Main) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Parse allowed origins from comma-separated list
			allowedOrigins := strings.Split(cfg.CORS.AllowOrigin, ",")
			
			// Get the origin from the request
			requestOrigin := r.Header.Get("Origin")
			
			// Check if request origin is in allowed list
			allowOrigin := ""
			for _, origin := range allowedOrigins {
				origin = strings.TrimSpace(origin)
				if requestOrigin == origin || cfg.CORS.AllowOrigin == "*" {
					allowOrigin = origin
					break
				}
			}
			
			// If origin is allowed (or wildcard), set the header
			if allowOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			} else if cfg.CORS.AllowOrigin == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept")
			w.Header().Set("Access-Control-Max-Age", "86400") // Cache preflight response for 24 hours

			// Handle preflight OPTIONS request globally
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// StartHTTPServer starts the HTTP server with the given handler and address.
func StartHTTPServer(addr string, handler http.Handler, cfg *config.Main) error {
	return http.ListenAndServe(addr, corsMiddleware(cfg)(handler))
}