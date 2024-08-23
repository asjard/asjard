package rest

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/valyala/fasthttp"
)

func NewCorsMiddleware(conf CorsConfig) MiddlewareFunc {
	logger.Debug("new cors middleware", "conf", conf)
	for _, origin := range conf.AllowOrigins {
		if origin == "*" {
			conf.allowAllOrigins = true
		}
	}
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			origin := string(ctx.Request.Header.Peek("Origin"))
			if len(origin) == 0 {
				next(ctx)
				return
			}
			host := string(ctx.Host())
			if origin == "http://"+host || origin == "https://"+host {
				next(ctx)
				return
			}
			if !corsIsOriginValid(conf, origin) {
				ctx.SetStatusCode(http.StatusForbidden)
				return
			}
			ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowOrigin, origin)
			if !conf.allowAllOrigins {
				ctx.Response.Header.Set("Vary", "Origin")
			}
			if string(ctx.Method()) == http.MethodOptions {
				corsPreflight(ctx, conf)
				return
			}
			if conf.AllowCredentials {
				ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowCredentials, "true")
			}
			if len(conf.ExposeHeaders) > 0 {
				ctx.Response.Header.Set(fasthttp.HeaderAccessControlExposeHeaders, strings.Join(conf.ExposeHeaders, ","))
			}
			next(ctx)
		}
	}
}

func corsPreflight(ctx *fasthttp.RequestCtx, conf CorsConfig) {
	if conf.AllowCredentials {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowCredentials, "true")
	}
	if len(conf.AllowMethods) > 0 {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowMethods, strings.Join(conf.AllowMethods, ","))
	}
	if len(conf.AllowHeaders) > 0 {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlAllowHeaders, strings.Join(conf.AllowHeaders, ","))
	}
	if conf.MaxAge.Duration != 0 {
		ctx.Response.Header.Set(fasthttp.HeaderAccessControlMaxAge, strconv.FormatInt(int64(conf.MaxAge.Duration/time.Second), 10))
	}
	if !conf.allowAllOrigins {
		ctx.Response.Header.Add(fasthttp.HeaderVary, fasthttp.HeaderAccessControlRequestMethod)
		ctx.Response.Header.Add(fasthttp.HeaderVary, fasthttp.HeaderAccessControlRequestHeaders)
	}
	ctx.SetStatusCode(http.StatusNoContent)
}

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
