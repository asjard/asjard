package metadata

import (
	"context"

	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/spf13/cast"
	"google.golang.org/grpc/metadata"
)

type Val string

func Get(ctx context.Context, key string) Val {
	rtx, ok := ctx.(*rest.Context)
	if ok {
		if vals := rtx.GetHeaderParam(key); len(vals) > 0 {
			return Val(vals[0])
		}
		return ""
	}
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if vals := md.Get(key); len(vals) > 0 {
			return Val(vals[0])
		}
	}
	return ""
}

func (v Val) Int64() int64 {
	return cast.ToInt64(v)
}

func (v Val) Int32() int32 {
	return cast.ToInt32(v)
}

func (v Val) String() string {
	return string(v)
}
