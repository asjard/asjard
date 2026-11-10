package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPathValuesFromWildcardRoute(t *testing.T) {
	mux := http.NewServeMux()
	var gotName string
	var gotValues map[string][]string
	mux.Handle("GET /hello/{name}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := NewContext(r, w)
		gotName = c.PathValue("name")
		gotValues = c.PathValues()
		w.WriteHeader(http.StatusOK)
	}))
	rt := &router{
		mux:              mux,
		notFound:         http.NotFoundHandler(),
		methodNotAllowed: http.NotFoundHandler(),
	}

	req := httptest.NewRequest(http.MethodGet, "/hello/world", nil)
	rec := httptest.NewRecorder()
	rt.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "world", gotName)
	require.Equal(t, map[string][]string{"name": {"world"}}, gotValues)
}

func TestPathValuesMultipleAndWildcard(t *testing.T) {
	mux := http.NewServeMux()
	var gotValues map[string][]string
	mux.Handle("DELETE /api/v1/hello/{name}/files/{path...}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := NewContext(r, w)
		gotValues = c.PathValues()
		w.WriteHeader(http.StatusOK)
	}))
	rt := &router{
		mux:              mux,
		notFound:         http.NotFoundHandler(),
		methodNotAllowed: http.NotFoundHandler(),
	}

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/hello/alice/files/a/b.txt", nil)
	rec := httptest.NewRecorder()
	rt.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, map[string][]string{
		"name": {"alice"},
		"path": {"a/b.txt"},
	}, gotValues)
}
