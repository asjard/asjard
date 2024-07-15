package status

import (
	"github.com/asjard/asjard/core/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FromError 解析错误为code和message
func FromError(err error) (code uint32, message string) {
	if err == nil {
		return 0, ""
	}
	if stts, ok := status.FromError(err); ok {
		code = uint32(stts.Code())
		message = stts.Message()

	} else {
		logger.Error("invalid err, must be status.Error", "err", err.Error())
		code = uint32(codes.Internal)
		message = err.Error()
	}
	return
}
