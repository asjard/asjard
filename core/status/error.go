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
// 	其中<=200为框架保留错误码

const (
	// InternalServerErrorStr 系统内部错误统一返回字符串
	InternalServerErrorStr = "internal server error"

	// SetCacheFailCode 设置缓存失败错误码
	SetCacheFailCode = 500_20
	// RefreshCacheFailCode 刷新缓存失败错误码
	RefreshCacheFailCode = 500_21
	// DeleteCacheFailCode 删除缓存失败错误码
	DeleteCacheFailCode = 500_22
	// DatabaseNotFoundCode 数据库不存在错误
	DatabaseNotFoundCode = 500_23
	// GetLockFailCode 获取锁失败错误
	GetLockFailCode = 500_24

	// UnsupportProtocolCode 暂不支持的协议
	UnsupportProtocolCode = 404_30
	// MethodNotAllowedCode 请求方法不匹配错误码
	MethodNotAllowedCode = 400_31
)

var (
	// InternalServerError 系统内部错误
	// 这里的error不能直接用变量的方式定义，因为里面包含了系统码的概念
	// 如果是变量在import的时候就执行了，那个时候配置文件还没有加载
	InternalServerError = func() error { return Error(codes.Internal, InternalServerErrorStr) }
	// PageNotFoundError 页面找不到
	PageNotFoundError = func() error { return Error(codes.NotFound, "page not found") }
	// MethodNotAllowedError 请求方法不匹配
	MethodNotAllowedError = func() error { return Error(MethodNotAllowedCode, "method not allowed") }
	// UnsupportProtocol 暂不支持的协议
	UnsupportProtocol = func() error { return Error(UnsupportProtocolCode, "unsupport protocol") }
	// TooManyRequest 请求过多
	TooManyRequest = func() error { return Error(codes.ResourceExhausted, "too may request") }

	// SetCacheFailError 设置缓存失败错误
	SetCacheFailError = func() error { return Error(SetCacheFailCode, InternalServerErrorStr) }
	// RefreCacheFailError 刷新缓存失败错误
	RefreCacheFailError = func() error { return Error(RefreshCacheFailCode, InternalServerErrorStr) }
	// DeleteCacheFailError 删除缓存失败错误
	DeleteCacheFailError = func() error { return Error(DeleteCacheFailCode, InternalServerErrorStr) }
	// DatabaseNotFoundError 数据库不存在
	DatabaseNotFoundError = func() error { return Error(DatabaseNotFoundCode, InternalServerErrorStr) }
)

// Error 添加系统码
// XXX_YYY_Z
// XXX 为系统码，固定三位
// YYY HTTP状态码,固定三位
// Z 错误码位数不固定
func Error(c codes.Code, msg string) error {
	return status.Error(newCode(c), msg)
}

func Errorf(c codes.Code, format string, a ...any) error {
	return status.Errorf(newCode(c), format, a...)
}

// https://datatracker.ietf.org/doc/html/rfc7231#section-6
func newCode(c codes.Code) codes.Code {
	var httpCode, errCode uint32
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
	errCode = httpCode*uint32(math.Pow10(int(utils.Uint32Len(errCode)))) + errCode
	return codes.Code(runtime.GetAPP().Instance.SystemCode*uint32(math.Pow10(int(utils.Uint32Len(errCode)))) + errCode)
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
		// errCodeLen := utils.Uint32Len(errCode)
		// if errCodeLen == 1 {
		// 	errCodeLen += 1
		// }
		// errCode = systemCode*uint32(math.Pow10(int(errCodeLen))) + errCode
	}
	return
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
