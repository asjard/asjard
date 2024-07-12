package rest

import (
	"net/http"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/protobuf/responsepb"
	"google.golang.org/grpc/codes"
	gstatus "google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// Response 请求返回
// type Response struct {
// 	*status.Status
// 	// 请求数据
// 	Data any `json:"data"`
// }

// var responsePool = sync.Pool{
// 	New: func() any {
// 		return &Response{
// 			Status: &status.Status{},
// 			Data:   nil,
// 		}
// 	},
// }

// func newResponse(c *Context, data any, err error) *Response {
// 	response := responsePool.Get().(*Response)
// 	if err == nil {
// 		response.Status = &status.Status{}
// 	} else {
// 		if stts, ok := err.(*status.Status); ok {
// 			response.Status = stts
// 		} else if stts, ok := gstatus.FromError(err); ok {
// 			response.Status = &status.Status{
// 				Code:    grpcCode2HttpStatusCode(stts.Code()),
// 				Message: stts.Message(),
// 			}
// 		} else {
// 			logger.Error(err.Error())
// 			response.Status = status.StatusInternalServerError
// 		}
// 	}
// 	if response.Status.Code != 0 && response.Status.Doc == "" {
// 		response.Status.Doc = c.errPage
// 	}
// 	response.Data = data
// 	c.response = response
// 	return response
// }

func newResponse(c *Context, data any, err error) *responsepb.Response {
	response := &responsepb.Response{
		Data: &anypb.Any{},
	}
	if err != nil {
		if stts, ok := gstatus.FromError(err); ok {
			response.Code = uint32(stts.Code())
			response.Message = stts.Message()

		} else {
			logger.Error("invalid err, must be status.Error", "err", err.Error())
			response.Code = uint32(codes.Internal)
			response.Message = err.Error()
		}
	} else {

		d, err := anypb.New(data.(proto.Message))
		if err != nil {
			logger.Error("can not create anypb.Any", "data", data)
			response.Code = uint32(codes.Internal)
			response.Message = "internal server error"
			return response
		}
		response.Data = d
	}
	if response.Code != 0 && response.Doc == "" {
		response.Doc = c.errPage
	}
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
