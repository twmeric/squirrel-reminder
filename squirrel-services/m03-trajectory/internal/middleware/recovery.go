package middleware

import (
	"context"
	"log"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryInterceptor panic恢复拦截器
func RecoveryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[PANIC] method=%s panic=%v\n%s", 
				info.FullMethod, 
				r,
				string(debug.Stack()),
			)
			err = status.Errorf(codes.Internal, "internal error: %v", r)
		}
	}()
	
	return handler(ctx, req)
}
