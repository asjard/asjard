package xhttp

import (
	"net/http"
	"strings"
)

// routePattern builds a Go ServeMux pattern from an HTTP method and path.
// Path must use net/http wildcards, e.g. "/users/{id}" or "/files/{path...}".
func routePattern(method, path string) string {
	if method == "" {
		return path
	}
	return method + " " + path
}

// pathParamNames extracts wildcard names from a ServeMux pattern.
func pathParamNames(pattern string) []string {
	var names []string
	for i := 0; i < len(pattern); i++ {
		if pattern[i] != '{' {
			continue
		}
		j := i + 1
		for j < len(pattern) && pattern[j] != '}' {
			j++
		}
		if j >= len(pattern) {
			break
		}
		name := strings.TrimSuffix(pattern[i+1:j], "...")
		if name != "" {
			names = append(names, name)
		}
		i = j
	}
	return names
}

func listenNetwork(addr string) (network, address string) {
	switch {
	case strings.HasPrefix(addr, "unix://"):
		return "unix", strings.TrimPrefix(addr, "unix://")
	case strings.HasPrefix(addr, "unix:"):
		return "unix", strings.TrimPrefix(addr, "unix:")
	case strings.HasPrefix(addr, "unix"):
		return "unix", addr
	default:
		return "tcp", addr
	}
}

func isMethodNotAllowed(pattern, method string) bool {
	routeMethod, _, found := strings.Cut(pattern, " ")
	if !found || strings.HasPrefix(routeMethod, "/") {
		return false
	}
	if routeMethod == method {
		return false
	}
	// ServeMux serves GET handlers for HEAD requests.
	if method == http.MethodHead && routeMethod == http.MethodGet {
		return false
	}
	return true
}
