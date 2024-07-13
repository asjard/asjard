package rest

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

func NewCorsMiddleware(conf CorsConfig) MiddlewareFunc {
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
			if !isOriginValid(conf, origin) {
				ctx.SetStatusCode(http.StatusForbidden)
				return
			}
			ctx.Response.Header.Set("Access-Control-Allow-Origin", origin)
			if !conf.allowAllOrigins {
				ctx.Response.Header.Set("Vary", "Origin")
			}
			if string(ctx.Method()) == http.MethodOptions {
				if conf.AllowCredentials {
					ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
				}
				if len(conf.AllowMethods) > 0 {
					ctx.Response.Header.Set("Access-Control-Allow-Methods", strings.Join(conf.AllowMethods, ","))
				}
				if len(conf.AllowHeaders) > 0 {
					ctx.Response.Header.Set("Access-Control-Allow-Headers", strings.Join(conf.AllowHeaders, ","))
				}
				if conf.MaxAge.Duration != 0 {
					ctx.Response.Header.Set("Access-Control-Max-Age", strconv.FormatInt(int64(conf.MaxAge.Duration/time.Second), 10))
				}
				if !conf.allowAllOrigins {
					ctx.Response.Header.Add("Vary", "Access-Control-Request-Method")
					ctx.Response.Header.Add("Vary", "Access-Control-Request-Headers")
				}
				ctx.SetStatusCode(http.StatusNoContent)
				return
			} else {
				if conf.AllowCredentials {
					ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
				}
				if len(conf.ExposeHeaders) > 0 {
					ctx.Response.Header.Set("Access-Control-Expose-Headers", strings.Join(conf.ExposeHeaders, ","))
				}
			}
			next(ctx)
		}
	}
}

func isOriginValid(conf CorsConfig, origin string) bool {
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
