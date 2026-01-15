/*
Package interceptors contains the middleware logic for the Asjard server.
The quota interceptor is responsible for traffic policing and rate limiting.
*/
package interceptors

const (
	// QuotaInterceptorName is the unique identifier for the quota/rate-limiting interceptor.
	QuotaInterceptorName = "quota"
)

// Quota represents the rate-limiting interceptor component.
// It is designed to protect the server from being overwhelmed by too many requests
// (e.g., implementing Token Bucket or Leaky Bucket algorithms).
type Quota struct{}
