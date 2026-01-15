package rest

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/valyala/fasthttp"
)

// NewCorsMiddleware creates a new middleware to handle Cross-Origin Resource Sharing.
func NewCorsMiddleware(conf CorsConfig) MiddlewareFunc {
	logger.Debug("new cors middleware", "conf", conf)

	// Pre-calculate if all origins are allowed to optimize runtime checks.
	for _, origin := range conf.AllowOrigins {
		if origin == "*" {
			conf.allowAllOrigins = true
		}
	}

	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			// 1. Get the Origin header from the request.
			origin := string(append([]byte(nil), ctx.Request.Header.Peek("Origin")...))

			// If no Origin header is present, it's not a cross-origin request.
			if len(origin) == 0 {
				next(ctx)
				return
			}

			// 2. Check if the origin matches the current host (Same-Origin).
			host := string(ctx.Host())
			if origin == "http://"+host || origin == "https://"+host {
				next(ctx)
				return
			}

			// 3. Validate the origin against the allowed list in the configuration.
			if !corsIsOriginValid(conf, origin) {
				ctx.SetStatusCode(http.StatusForbidden)
				return
			}

			// 4. Set standard CORS response headers.
			ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowOrigin, origin)

			// If not allowing all, set 'Vary: Origin' to prevent cache poisoning.
			if !conf.allowAllOrigins {
				ctx.Response.Header.Set("Vary", "Origin")
			}

			// 5. Handle Preflight requests (OPTIONS method).
			if string(ctx.Method()) == http.MethodOptions {
				corsPreflight(ctx, conf)
				return
			}

			// 6. Handle credentials (Cookies/Auth headers).
			if conf.AllowCredentials {
				ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowCredentials, "true")
			}

			// 7. Expose custom headers to the browser client.
			if len(conf.ExposeHeaders) > 0 {
				ctx.Response.Header.Set(fasthttp.HeaderAccessControlExposeHeaders, strings.Join(conf.ExposeHeaders, ","))
			}

			// Continue to the next handler for actual business logic.
			next(ctx)
		}
	}
}

// corsPreflight handles the OPTIONS request sent by browsers before a "non-simple" request.
func corsPreflight(ctx *fasthttp.RequestCtx, conf CorsConfig) {
	if conf.AllowCredentials {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowCredentials, "true")
	}

	// Inform the browser which methods are allowed.
	if len(conf.AllowMethods) > 0 {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowMethods, strings.Join(conf.AllowMethods, ","))
	}

	// Inform the browser which headers can be sent.
	if len(conf.AllowHeaders) > 0 {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowHeaders, strings.Join(conf.AllowHeaders, ","))
	}

	// Cache the preflight response in the browser for a specific duration.
	if conf.MaxAge.Duration != 0 {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlMaxAge, strconv.FormatInt(int64(conf.MaxAge.Duration/time.Second), 10))
	}

	// Tell proxies that the response varies based on these request headers.
	if !conf.allowAllOrigins {
		ctx.Response.Header.Add(fasthttp.HeaderVary, fasthttp.HeaderAccessControlRequestMethod)
		ctx.Response.Header.Add(fasthttp.HeaderVary, fasthttp.HeaderAccessControlRequestHeaders)
	}

	// Preflight requests return 204 No Content.
	ctx.SetStatusCode(http.StatusNoContent)
}

// corsIsOriginValid checks if the provided origin is permitted by the configuration.
func corsIsOriginValid(conf CorsConfig, origin string) bool {
	if conf.allowAllOrigins {
		return true
	}
	for _, value := range conf.AllowOrigins {
		if value == origin {
			return true
		}
	}
	return false
}
