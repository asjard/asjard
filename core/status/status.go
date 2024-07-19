package status

import (
	"net/http"

	"github.com/asjard/asjard/core/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Status struct {
	// systemCode, httpCode, errorCode合并后的结果
	Code uint32
	// 系统码
	SystemCode uint32
	// Http状态码
	HttpCode uint32
	// 错误码
	ErrorCode uint32
	// 错误信息
	Message string
}

// FromError 解析错误为code和message
func FromError(err error) *Status {
	result := &Status{}
	if err == nil {
		return result
	}
	if stts, ok := status.FromError(err); ok {
		result.Code = uint32(stts.Code())
		result.SystemCode, result.HttpCode, result.ErrorCode = parseCode(stts.Code())
		result.Message = stts.Message()

	} else {
		logger.Error("invalid err, must be status.Error", "err", err.Error())
		// code = uint32(codes.Internal)
		result.Code = uint32(codes.Internal)
		result.HttpCode = http.StatusInternalServerError
		result.ErrorCode = uint32(codes.Internal)
		result.Message = err.Error()
	}
	return result
}
