package status

import (
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/protobuf/statuspb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FromError 解析错误为code和message
func FromError(err error) *statuspb.Status {
	result := &statuspb.Status{
		Success: true,
	}
	if err == nil {
		return result
	}
	result.Success = false
	if stts, ok := status.FromError(err); ok {
		result.Code = uint32(stts.Code())
		result.System, result.Status, _ = parseCode(stts.Code())
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
		result.System, result.Status, _ = parseCode(codes.Internal)
		result.Message = err.Error()
	}
	return result
}
