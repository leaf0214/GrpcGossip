package interceptor

import (
	"context"
	g "google.golang.org/grpc"
)

type Interceptor struct {
	UnaryInterceptors  []g.UnaryServerInterceptor
	StreamInterceptors []g.StreamServerInterceptor
}

func Logger(ctx context.Context, req interface{}, _ *g.UnaryServerInfo, handler g.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}
