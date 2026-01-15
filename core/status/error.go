package status

import (
	"math"
	"net/http"

	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The status package wraps gRPC status and codes to introduce the concept of
// System Codes and custom Error Codes.
// Error codes <= 200 are reserved for framework internal use.

const (
	// InternalServerErrorStr is the generic message returned for hidden internal failures.
	InternalServerErrorStr = "internal server error"

	// --- Business Error Codes (Format: HTTP_BusinessCode) ---

	// SetCacheFailCode error when writing to cache.
	SetCacheFailCode = 500_20
	// RefreshCacheFailCode error when updating existing cache entries.
	RefreshCacheFailCode = 500_21
	// DeleteCacheFailCode error when removing cache entries.
	DeleteCacheFailCode = 500_22
	// DatabaseNotFoundCode error when a database connection or resource is missing.
	DatabaseNotFoundCode = 500_23
	// GetLockFailCode error when a distributed lock cannot be acquired.
	GetLockFailCode = 500_24

	// UnsupportProtocolCode error when the requested protocol is not handled.
	UnsupportProtocolCode = 404_30
	// MethodNotAllowedCode error for mismatched HTTP methods (e.g., POST instead of GET).
	MethodNotAllowedCode = 400_31
)

var (
	// Standard Errors are defined as functions because they rely on the 'SystemCode'
	// which is only available after configuration is loaded at runtime.

	// InternalServerError generic 500 error.
	InternalServerError = func() error { return Error(codes.Internal, InternalServerErrorStr) }
	// PageNotFoundError generic 404 error.
	PageNotFoundError = func() error { return Error(codes.NotFound, "page not found") }
	// MethodNotAllowedError generic 405 error.
	MethodNotAllowedError = func() error { return Error(MethodNotAllowedCode, "method not allowed") }
	// UnsupportProtocol error for invalid protocol requests.
	UnsupportProtocol = func() error { return Error(UnsupportProtocolCode, "unsupport protocol") }
	// TooManyRequest generic 429 rate limit error.
	TooManyRequest = func() error { return Error(codes.ResourceExhausted, "too may request") }

	// SetCacheFailError specific error for cache write failures.
	SetCacheFailError = func() error { return Error(SetCacheFailCode, InternalServerErrorStr) }
	// RefreCacheFailError specific error for cache refresh failures.
	RefreCacheFailError = func() error { return Error(RefreshCacheFailCode, InternalServerErrorStr) }
	// DeleteCacheFailError specific error for cache deletion failures.
	DeleteCacheFailError = func() error { return Error(DeleteCacheFailCode, InternalServerErrorStr) }
	// DatabaseNotFoundError specific error for database resource missing.
	DatabaseNotFoundError = func() error { return Error(DatabaseNotFoundCode, InternalServerErrorStr) }
)

// Error creates a gRPC status error with an asjard structured code.
// Logic: Generates a code in the format XXXYYYZZ...
// XXX: System Code (3 digits)
// YYY: HTTP Status Code (3 digits)
// ZZ...: Business Error Code (variable length)
func Error(c codes.Code, msg string) error {
	return status.Error(newCode(c), msg)
}

// Errorf creates a formatted gRPC status error with a structured code.
func Errorf(c codes.Code, format string, a ...any) error {
	return status.Errorf(newCode(c), format, a...)
}

// newCode encodes the gRPC code/business code into the asjard unified format.
func newCode(c codes.Code) codes.Code {
	var httpCode, errCode uint32

	if c < 100 {
		// Standard gRPC codes (0-16): infer HTTP mapping.
		httpCode = httpStatusCode(c)
		errCode = uint32(c)
	} else if c < 1000 {
		// Bare HTTP status codes (1xx-5xx).
		httpCode = http.StatusInternalServerError
		errCode = uint32(c)
	} else {
		// Custom composite codes: extract the embedded HTTP code part.
		n := int(utils.Uint32Len(uint32(c)) - 3)
		httpCode = uint32(c) / uint32(math.Pow10(n))
		if http.StatusText(int(httpCode)) == "" {
			httpCode = http.StatusInternalServerError
		}
		errCode = uint32(c) % uint32(math.Pow10(n))
	}

	// Combine: [SystemCode][HTTPCode][BusinessCode]
	errCode = httpCode*uint32(math.Pow10(int(utils.Uint32Len(errCode)))) + errCode
	return codes.Code(runtime.GetAPP().Instance.SystemCode*uint32(math.Pow10(int(utils.Uint32Len(errCode)))) + errCode)
}

// parseCode decomposes a structured code back into its constituent parts.
func parseCode(c codes.Code) (systemCode, httpCode, errCode uint32) {
	if c < 10 {
		httpCode = httpStatusCode(c)
		errCode = uint32(c)
		return
	} else if c < 100_000_0 {
		httpCode = http.StatusInternalServerError
		errCode = uint32(c)
	} else {
		// Complex parsing for 7+ digit codes.
		n := int(utils.Uint32Len(uint32(c)))
		errCode = uint32(c) % uint32(math.Pow10(n-6))
		sysHttpCode := uint32(c) / uint32(math.Pow10(n-6))
		httpCode = sysHttpCode % uint32(math.Pow10(3))
		systemCode = sysHttpCode / uint32(math.Pow10(3))
	}
	return
}

// httpStatusCode maps gRPC codes to standard HTTP status codes based on
// official gRPC-HTTP mapping documentation.
func httpStatusCode(code codes.Code) uint32 {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return 499 // Client Closed Request
	case codes.Unknown, codes.Internal, codes.DataLoss:
		return http.StatusInternalServerError
	case codes.InvalidArgument, codes.FailedPrecondition, codes.OutOfRange:
		return http.StatusBadRequest
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists, codes.Aborted:
		return http.StatusConflict
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
