/*
Package status 对grpc错误的一层包装，添加了系统码，http状态码，错误码的概念
以及一些框架定义的错误
*/
package status

import (
	"net/http"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/protobuf/statuspb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FromError 解析错误为code和message
func FromError(err error) *statuspb.Status {
	result := &statuspb.Status{
		Success: true,
		Status:  http.StatusOK,
	}
	if err == nil {
		return result
	}
	result.Success = false
	if stts, ok := status.FromError(err); ok {
		result.Code = uint32(stts.Code())
		result.System, result.Status, result.ErrCode = parseCode(stts.Code())
		result.Message = stts.Message()
		for _, detail := range stts.Details() {
			if st, ok := detail.(*statuspb.Status); ok {
				result.Doc = st.Doc
				result.Prompt = st.Prompt
			}
		}

	} else {
		logger.Error("invalid err, must be status.Error", "err", err.Error())
		result.Code = uint32(codes.Internal)
		result.ErrCode = result.Code
		result.System, result.Status, _ = parseCode(codes.Internal)
		result.Message = err.Error()
	}
	return result
}
