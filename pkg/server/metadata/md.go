package metadata

import (
	"context"

	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/spf13/cast"
	"google.golang.org/grpc/metadata"
)

// Val represents a metadata value.
// It is a string alias that provides convenient type conversion methods.
type Val string

// Get retrieves a metadata value from the context for a specific key.
// It automatically detects if the context is a REST context or a gRPC context.
func Get(ctx context.Context, key string) Val {
	// 1. Attempt to treat context as a REST (fasthttp) context.
	rtx, ok := ctx.(*rest.Context)
	if ok {
		// In REST, metadata is typically stored in HTTP Headers.
		if vals := rtx.GetHeaderParam(key); len(vals) > 0 {
			return Val(vals[0])
		}
		return ""
	}

	// 2. Fallback to gRPC incoming context metadata.
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		// gRPC metadata stores multiple values per key; we take the first.
		if vals := md.Get(key); len(vals) > 0 {
			return Val(vals[0])
		}
	}

	// Return empty if the key is not found in either protocol.
	return ""
}

// Int64 converts the metadata value to an int64 using robust casting.
func (v Val) Int64() int64 {
	return cast.ToInt64(v)
}

// Int32 converts the metadata value to an int32.
func (v Val) Int32() int32 {
	return cast.ToInt32(v)
}

// String returns the underlying string value.
func (v Val) String() string {
	return string(v)
}
