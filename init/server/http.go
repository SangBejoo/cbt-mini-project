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
			allowedOrigins := strings.Split(cfg.CORS.AllowOrigin, ",")
			requestOrigin := strings.TrimSpace(r.Header.Get("Origin"))

			w.Header().Set("Vary", "Origin")
			w.Header().Set("Vary", "Access-Control-Request-Method")
			w.Header().Set("Vary", "Access-Control-Request-Headers")

			allowOrigin := ""
			for _, origin := range allowedOrigins {
				allowed := strings.TrimSpace(origin)
				if allowed == "" {
					continue
				}
				if allowed == "*" {
					if requestOrigin != "" {
						allowOrigin = requestOrigin
					} else {
						allowOrigin = "*"
					}
					break
				}
				if requestOrigin == allowed {
					allowOrigin = requestOrigin
					break
				}
			}

			if allowOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-Requested-With, X-Request-Id")
			w.Header().Set("Access-Control-Max-Age", "86400")

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