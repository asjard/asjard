package xhttp

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/server/rest"
)

// NewCorsMiddleware creates a new middleware to handle Cross-Origin Resource Sharing.
func NewCorsMiddleware(conf rest.CorsConfig) MiddlewareFunc {
	logger.Debug("new cors middleware", "conf", conf)

	allowAllOrigins := slices.Contains(conf.AllowOrigins, "*")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			host := r.Host
			if origin == "http://"+host || origin == "https://"+host {
				next.ServeHTTP(w, r)
				return
			}

			if !allowAllOrigins && !slices.Contains(conf.AllowOrigins, origin) {
				w.WriteHeader(http.StatusForbidden)
				return
			}

			w.Header().Set(headerAccessControlAllowOrigin, origin)
			if !allowAllOrigins {
				w.Header().Set("Vary", "Origin")
			}

			if r.Method == http.MethodOptions {
				corsPreflight(w, conf, allowAllOrigins)
				return
			}

			if conf.AllowCredentials {
				w.Header().Set(headerAccessControlAllowCredentials, "true")
			}
			if len(conf.ExposeHeaders) > 0 {
				w.Header().Set(headerAccessControlExposeHeaders, strings.Join(conf.ExposeHeaders, ","))
			}
			next.ServeHTTP(w, r)
		})
	}
}

func corsPreflight(w http.ResponseWriter, conf rest.CorsConfig, allowAllOrigins bool) {
	if conf.AllowCredentials {
		w.Header().Set(headerAccessControlAllowCredentials, "true")
	}
	if len(conf.AllowMethods) > 0 {
		w.Header().Set(headerAccessControlAllowMethods, strings.Join(conf.AllowMethods, ","))
	}
	if len(conf.AllowHeaders) > 0 {
		w.Header().Set(headerAccessControlAllowHeaders, strings.Join(conf.AllowHeaders, ","))
	}
	if conf.MaxAge.Duration != 0 {
		w.Header().Set(headerAccessControlMaxAge, strconv.FormatInt(int64(conf.MaxAge.Duration/time.Second), 10))
	}
	if !allowAllOrigins {
		w.Header().Add("Vary", headerAccessControlRequestMethod)
		w.Header().Add("Vary", headerAccessControlRequestHeaders)
	}
	w.WriteHeader(http.StatusNoContent)
}
