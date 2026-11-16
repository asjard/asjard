package rest

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/asjard/asjard/core/logger"
)

// NewCorsMiddleware creates a new middleware to handle Cross-Origin Resource Sharing.
func NewCorsMiddleware(conf CorsConfig) MiddlewareFunc {
	logger.Debug("new cors middleware", "conf", conf)

	if conf.Disabled {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// Pre-calculate if all origins are allowed to optimize runtime checks.
	for _, origin := range conf.AllowOrigins {
		if origin == "*" {
			conf.allowAllOrigins = true
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the Origin header from the request.
			origin := r.Header.Get("Origin")

			// If no Origin header is present, it's not a cross-origin request.
			if len(origin) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Check if the origin matches the current host (Same-Origin).
			host := string(r.Host)
			if origin == "http://"+host || origin == "https://"+host {
				next.ServeHTTP(w, r)
				return
			}

			// Validate the origin against the allowed list in the configuration.
			if !corsIsOriginValid(conf, origin) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			// Set standard CORS response headers.
			w.Header().Set("Access-Control-Allow-Origin", origin)

			// If not allowing all, set 'Vary: Origin' to prevent cache poisoning.
			if !conf.allowAllOrigins {
				w.Header().Set("Vary", "Origin")
			}

			// Handle Preflight requests (OPTIONS method).
			if string(r.Method) == http.MethodOptions {
				corsPreflight(w, conf)
				return
			}

			// Handle credentials (Cookies/Auth headers).
			if conf.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Expose custom headers to the browser client.
			if len(conf.ExposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(conf.ExposeHeaders, ","))
			}

			// Continue to the next handler for actual business logic.
			next.ServeHTTP(w, r)
		})
	}
}

// corsPreflight handles the OPTIONS request sent by browsers before a "non-simple" request.
func corsPreflight(w http.ResponseWriter, conf CorsConfig) {
	if conf.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// Inform the browser which methods are allowed.
	if len(conf.AllowMethods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(conf.AllowMethods, ","))
	}

	// Inform the browser which headers can be sent.
	if len(conf.AllowHeaders) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(conf.AllowHeaders, ","))
	}

	// Cache the preflight response in the browser for a specific duration.
	if conf.MaxAge.Duration != 0 {
		w.Header().Set("Access-Control-Max-Age", strconv.FormatInt(int64(conf.MaxAge.Duration/time.Second), 10))
	}

	// Tell proxies that the response varies based on these request headers.
	if !conf.allowAllOrigins {
		w.Header().Add("Vary", "Access-Control-Request-Method")
		w.Header().Add("Vary", "Access-Control-Request-Headers")
	}

	// Preflight requests return 204 No Content.
	w.WriteHeader(http.StatusNoContent)
}

// corsIsOriginValid checks if the provided origin is permitted by the configuration.
func corsIsOriginValid(conf CorsConfig, origin string) bool {
	if conf.allowAllOrigins {
		return true
	}
	return slices.Contains(conf.AllowOrigins, origin)
}
