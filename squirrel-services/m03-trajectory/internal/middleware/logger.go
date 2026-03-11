package middleware

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// UnaryInterceptor 一元RPC拦截器
func UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	
	resp, err := handler(ctx, req)
	
	duration := time.Since(start)
	st, _ := status.FromError(err)
	
	log.Printf("[gRPC] method=%s duration=%v code=%s",
		info.FullMethod,
		duration,
		st.Code(),
	)
	
	return resp, err
}
