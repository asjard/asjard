package rest

import (
	"net/http"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/protobuf/responsepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

func newResponse(c *Context, data any, err error) *responsepb.Response {
	response := &responsepb.Response{
		Success: true,
		Status:  http.StatusOK,
		Data:    &anypb.Any{},
	}
	if err != nil {
		st := status.FromError(err)
		response.Code = st.Code
		response.Status = st.HttpCode
		response.System = st.SystemCode
		response.Success = false
		response.Message = st.Message
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
