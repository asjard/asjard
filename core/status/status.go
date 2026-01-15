package status

import (
	"net/http"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/protobuf/statuspb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FromError parses a standard Go error into a structured statuspb.Status message.
// This is used by gateways and interceptors to format the final response sent to the client.
func FromError(err error) *statuspb.Status {
	// Initialize a default success state.
	result := &statuspb.Status{
		Success: true,
		Status:  http.StatusOK,
	}

	// If no error occurred, return the success status.
	if err == nil {
		return result
	}

	result.Success = false

	// Check if the error is a gRPC status error.
	if stts, ok := status.FromError(err); ok {
		// Extract the raw numerical code (e.g., 10140430).
		result.Code = uint32(stts.Code())

		// Decompose the code into System, HTTP, and Business segments.
		result.System, result.Status, result.ErrCode = parseCode(stts.Code())

		// The actual error message.
		result.Message = stts.Message()

		// Iterate through gRPC error details to find extended asjard metadata.
		// This allows attaching troubleshooting docs or UI prompts to an error.
		for _, detail := range stts.Details() {
			if st, ok := detail.(*statuspb.Status); ok {
				result.Doc = st.Doc
				result.Prompt = st.Prompt
			}
		}

	} else {
		// If the error is not a gRPC status (primitive Go error),
		// log it as a violation of framework standards and treat it as an Internal Server Error.
		logger.Error("invalid err, must be status.Error", "err", err.Error())

		result.Code = uint32(codes.Internal)
		result.ErrCode = result.Code
		result.System, result.Status, _ = parseCode(codes.Internal)
		result.Message = err.Error()
	}

	return result
}
