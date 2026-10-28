package xhttp

import (
	"net/http"
	"reflect"
	"sync"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/protobuf/statuspb"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type WriteKey int

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

	// If response err not nil and this key setted in ctx will return data in response
	WriteDataInErr WriteKey = 1
)

// Writer defines the function signature for outputting results to the client.
// It takes the request context, the successful data, and any potential error.
type Writer func(ctx *Context, data any, err error)

var (
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
	if err == nil && (data == nil || isNilValue(data)) {
		return
	}

	st := status.FromError(err)
	fillStatusMeta(c, st)
	attachStatusData(c, st, data, err)
	if err := writeJSON(c, int(responseStatusCode(c, st)), st); err != nil {
		logger.Error("write json fail", "err", err)
	}
}

func responseStatusCode(c *Context, st *statuspb.Status) uint32 {
	if c.QueryArgs().Has(QueryParamNeedStatusCode) {
		return st.Status
	}
	return http.StatusOK
}

func fillStatusMeta(c *Context, st *statuspb.Status) {
	if requestId := c.Value(HeaderResponseRequestID); requestId != nil {
		if id, ok := requestId.(string); ok {
			st.RequestId = id
		}
	}
	if requestMethod := c.Value(HeaderResponseRequestMethod); requestMethod != nil {
		if method, ok := requestMethod.(string); ok {
			st.RequestMethod = method
		}
	}
	c.Response.Header().Set(HeaderResponseRequestID, st.RequestId)
	c.Response.Header().Set(HeaderResponseRequestMethod, st.RequestMethod)
}

func attachStatusData(c *Context, st *statuspb.Status, data any, err error) {
	if (err != nil && c.Value(WriteDataInErr) == nil) || data == nil {
		return
	}
	msg, ok := data.(proto.Message)
	if !ok {
		return
	}
	d, err := anypb.New(msg)
	if err != nil {
		logger.Error("can not create anypb.Any", "data", data, "err", err)
		return
	}
	st.Data = d
}

func writeJSON(c *Context, statusCode int, body proto.Message) error {
	c.Response.Header().Set(headerContentType, MIME_JSON)
	c.SetStatusCode(statusCode)

	b, err := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: true,
	}.Marshal(body)
	if err != nil {
		return err
	}
	_, err = c.Write(b)
	return err
}

func isNilValue(data any) bool {
	v := reflect.ValueOf(data)
	switch v.Kind() {
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Func, reflect.Chan:
		return v.IsNil()
	default:
		return false
	}
}
