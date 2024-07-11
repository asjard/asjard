package rest

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/valyala/fasthttp"
)

type CorsMiddleware struct {
	conf CorsConfig
}

func NewCorsMiddleware(conf CorsConfig) *CorsMiddleware {
	for _, origin := range conf.AllowOrigins {
		if origin == "*" {
			conf.allowAllOrigins = true
		}
	}
	return &CorsMiddleware{
		conf: conf,
	}
}

func (c CorsMiddleware) Handler(next fasthttp.RequestHandler) func(ctx *fasthttp.RequestCtx) {
	return func(ctx *fasthttp.RequestCtx) {
		origin := string(ctx.Request.Header.Peek("Origin"))
		if len(origin) == 0 {
			next(ctx)
			return
		}
		host := string(ctx.Host())
		logger.Debug("cors", "origin", origin, "host", host)
		if origin == "http://"+host || origin == "https://"+host {
			next(ctx)
			return
		}
		if !isOriginValid(c.conf, origin) {
			ctx.SetStatusCode(http.StatusForbidden)
			return
		}
		ctx.Response.Header.Set("Access-Control-Allow-Origin", origin)
		if !c.conf.allowAllOrigins {
			ctx.Response.Header.Set("Vary", "Origin")
		}
		if string(ctx.Method()) == http.MethodOptions {
			logger.Debug("cors options request")
			if c.conf.AllowCredentials {
				ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
			}
			if len(c.conf.AllowMethods) > 0 {
				ctx.Response.Header.Set("Access-Control-Allow-Methods", strings.Join(c.conf.AllowMethods, ","))
			}
			if len(c.conf.AllowHeaders) > 0 {
				ctx.Response.Header.Set("Access-Control-Allow-Headers", strings.Join(c.conf.AllowHeaders, ","))
			}
			if c.conf.MaxAge.Duration != 0 {
				ctx.Response.Header.Set("Access-Control-Max-Age", strconv.FormatInt(int64(c.conf.MaxAge.Duration/time.Second), 10))
			}
			if !c.conf.allowAllOrigins {
				ctx.Response.Header.Add("Vary", "Access-Control-Request-Method")
				ctx.Response.Header.Add("Vary", "Access-Control-Request-Headers")
			}
			ctx.SetStatusCode(http.StatusNoContent)
			return
		} else {
			logger.Debug("cors not options request")
			if c.conf.AllowCredentials {
				ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")
			}
			if len(c.conf.ExposeHeaders) > 0 {
				ctx.Response.Header.Set("Access-Control-Expose-Headers", strings.Join(c.conf.ExposeHeaders, ","))
			}
		}
		next(ctx)
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
