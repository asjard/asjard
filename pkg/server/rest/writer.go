package rest

import (
	"net/http"
	"reflect"
	"sync"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/valyala/fasthttp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	// QueryParamNeedStatusCode: If this query parameter (e.g., ?nsc=1) is present,
	// the HTTP status code will be mapped from the business status code.
	QueryParamNeedStatusCode = "nsc"
	// HeaderResponseRequestMethod is the header key for returning the request method.
	HeaderResponseRequestMethod = "x-request-method"
	// HeaderResponseRequestID is the header key for returning the unique request ID.
	HeaderResponseRequestID = "x-request-id"
	// DefaultWriterName is the identifier for the standard JSON output handler.
	DefaultWriterName = "default"
)

// Writer defines the function signature for outputting results to the client.
// It takes the request context, the successful data, and any potential error.
type Writer func(ctx *Context, data any, err error)

var (
	// Registry for different output formats.
	writers = map[string]Writer{
		DefaultWriterName: DefaultWriter,
	}
	wm sync.RWMutex
)

// AddWriter registers a new custom output implementation (e.g., an XMLWriter).
func AddWriter(name string, writer Writer) {
	wm.Lock()
	writers[name] = writer
	wm.Unlock()
}

// GetWriter retrieves a writer by name, falling back to the default if not found.
func GetWriter(name string) Writer {
	wm.RLock()
	defer wm.RUnlock()
	if name == "" {
		return writers[DefaultWriterName]
	}
	w, ok := writers[name]
	if ok {
		return w
	}
	return writers[DefaultWriterName]
}

// DefaultWriter handles the standard output logic.
// It wraps responses in a standardized 'status' object containing metadata,
// errors, and the business data payload.
func DefaultWriter(c *Context, data any, err error) {
	// If both data and error are nil, we assume the handler has already
	// manually written the response (e.g., file downloads or streaming).
	if err == nil && (data == nil || reflect.ValueOf(data).IsNil()) {
		return
	}

	// Convert the error into a standardized status object.
	st := status.FromError(err)
	var statusCode uint32 = http.StatusOK

	// Check if the client requested the actual HTTP status code via query params.
	if c.URI().QueryArgs().Has(QueryParamNeedStatusCode) {
		statusCode = st.Status
	}

	// Inject tracing metadata (Request ID and Method) into the response.
	if requestId := c.Value(HeaderResponseRequestID); requestId != nil {
		st.RequestId = requestId.(string)
	}
	if requestMethod := c.Value(HeaderResponseRequestMethod); requestMethod != nil {
		st.RequestMethod = requestMethod.(string)
	}

	c.Response.Header.Set(HeaderResponseRequestID, st.RequestId)
	c.Response.Header.Set(HeaderResponseRequestMethod, st.RequestMethod)

	// If successful, wrap the business data in a Protobuf 'Any' type.
	if err == nil {
		if d, err := anypb.New(data.(proto.Message)); err == nil {
			st.Data = d
		} else {
			logger.Error("can not create anypb.Any", "data", data, "err", err)
		}
	}

	// Finalize by writing the status object as JSON.
	if err := writeJSON(c, int(statusCode), st); err != nil {
		logger.Error("write json fail", "err", err)
	}
}

// writeJSON handles the physical serialization and writing of bytes to the fasthttp response.
func writeJSON(c *Context, statusCode int, body proto.Message) error {
	c.Response.Header.Set(fasthttp.HeaderContentType, MIME_JSON)
	c.Response.SetStatusCode(statusCode)

	// Marshal using protojson to ensure Protobuf field names and empty values
	// are handled according to .proto definitions.
	b, err := protojson.MarshalOptions{
		UseProtoNames:   true, // Use snake_case from .proto files.
		EmitUnpopulated: true, // Include fields with default values (0, "", false).
	}.Marshal(body)
	if err != nil {
		return err
	}

	if _, err := c.Write(b); err != nil {
		return err
	}
	return nil
}
