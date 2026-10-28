package rest

import (
	"context"
)

// RestContext is the REST/HTTP request context.
// xfasthttp.Context and xhttp.Context both implement this interface so
// interceptors, tracing, metadata, and OpenAPI stay engine-agnostic.
type RestContext interface {
	Method() string
	Path() string
	Header(key string) string
	Headers() map[string][]string
	Query(key string) string
	Queries() map[string][]string
	JSONBody() []byte
	UserValue(key any) any
	SetUserValue(key, value any)
	UserValues() map[string][]string
	ClientIP() string
	RequestSize() int
	ResponseSize() int

	Context() context.Context
	SetContext(ctx context.Context)
	IsI18nEnabled() bool

	Write(p []byte) (int, error)
	SetContentType(contentType string)
	SetStatusCode(statusCode int)
	SendFile(path string)
	Redirect(uri string, statusCode int)

	// Request() *http.Request
	// Response() http.ResponseWriter
}
