package status

import (
	"math"
	"net/http"

	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 对grpc的status和code外面再包装一层，添加系统码和错误码的概念

var (
	// InternalServerError 系统内部错误
	InternalServerError   = Error(codes.Internal, "internal server error")
	DatabaseNotFoundError = Error(codes.Internal, "database not found")
	InvalidDBError        = Error(codes.Internal, "invalid db")
	PageNotFoundError     = Error(codes.NotFound, "page not found")
	MethodNotAllowedError = Error(codes.Unimplemented, "method not allowed")
	UnsupportProtocol     = Error(codes.Unavailable, "unsupport protocol")
)

// Error 添加系统码
// XXX_YYY_Z
// XXX 为系统码，固定三位
// YYY HTTP状态码,固定三位
// Z 错误码位数不固定
func Error(c codes.Code, msg string) error {
	return status.Error(newWithSystemCode(runtime.GetAPP().Instance.SystemCode, c), msg)
}

func Errorf(c codes.Code, format string, a ...any) error {
	return status.Errorf(newWithSystemCode(runtime.GetAPP().Instance.SystemCode, c), format, a...)
}

// https://datatracker.ietf.org/doc/html/rfc7231#section-6
func newCode(c codes.Code) (httpCode, errCode uint32) {
	// var httpCode, errCode uint32
	// 没有定义http状态码，从codes.Code中推断
	if c < 100 {
		httpCode = httpStatusCode(c)
		errCode = uint32(c)
	} else if c < 1000 {
		// http状态码1xx - 5xx
		httpCode = http.StatusInternalServerError
		errCode = uint32(c)
	} else {
		n := int(utils.Uint32Len(uint32(c)) - 3)
		httpCode = uint32(c) / uint32(math.Pow10(n))
		if http.StatusText(int(httpCode)) == "" {
			httpCode = http.StatusInternalServerError
		}
		errCode = uint32(c) % uint32(math.Pow10(n))
	}
	return
}

func parseCode(c codes.Code) (systemCode, httpCode, errCode uint32) {
	if c < 10 {
		httpCode = httpStatusCode(c)
		errCode = uint32(c)
		return
	} else if c < 100_000_0 {
		httpCode = http.StatusInternalServerError
		errCode = uint32(c)
	} else {
		n := int(utils.Uint32Len(uint32(c)))
		errCode = uint32(c) % uint32(math.Pow10(n-6))
		sysHttpCode := uint32(c) / uint32(math.Pow10(n-6))
		httpCode = sysHttpCode % uint32(math.Pow10(3))
		systemCode = sysHttpCode / uint32(math.Pow10(3))
	}
	return
}

func newWithSystemCode(systemCode uint32, c codes.Code) codes.Code {
	httpCode, errCode := newCode(c)
	errCode = httpCode*uint32(math.Pow10(int(utils.Uint32Len(errCode)))) + errCode
	return codes.Code(systemCode*uint32(math.Pow10(int(utils.Uint32Len(errCode)))) + errCode)
}

// 从code解析到http状态码
// https://chromium.googlesource.com/external/github.com/grpc/grpc/+/refs/tags/v1.21.4-pre1/doc/statuscodes.md
func httpStatusCode(code codes.Code) uint32 {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.Canceled:
		return 499
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
