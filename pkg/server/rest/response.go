package rest

import (
	"net/http"
	"sync"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/status"
	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
)

// Response 请求返回
type Response struct {
	*status.Status
	// 请求数据
	Data any `json:"data"`
}

var responsePool = sync.Pool{
	New: func() any {
		return &Response{
			Status: &status.Status{},
			Data:   nil,
		}
	},
}

func newResponse(c *Context, data any, err error) *Response {
	response := responsePool.Get().(*Response)
	if err == nil {
		response.Status = &status.Status{}
	} else {
		if stts, ok := err.(*status.Status); ok {
			response.Status = stts
		} else if stts, ok := gstatus.FromError(err); ok {
			response.Status = &status.Status{
				Code:    grpcCode2HttpStatusCode(stts.Code()),
				Message: stts.Message(),
			}
		} else {
			logger.Error(err.Error())
			response.Status = status.StatusInternalServerError
		}
	}
	if response.Status.Code != 0 && response.Status.Doc == "" {
		response.Status.Doc = c.errPage
	}
	response.Data = data
	c.response = response
	return response
}

// grpc 状态码转换为http状态码
// https://chromium.googlesource.com/external/github.com/grpc/grpc/+/refs/tags/v1.21.4-pre1/doc/statuscodes.md
func grpcCode2HttpStatusCode(code codes.Code) uint32 {
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
