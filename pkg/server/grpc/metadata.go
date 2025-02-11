package grpc

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// MetadataGet 获取元数据
func MetadataGet(ctx context.Context, key string) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if vals := md.Get(key); len(vals) > 0 {
			return vals[0]
		}
	}
	return ""
}
