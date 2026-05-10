package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/CoffeeSi/golang2AITU/assignment4/appointment-service/internal/cache"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func RateLimitInterceptor(c cache.CacheClientInterface, limitRPM int) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		p, ok := peer.FromContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		clientIP := strings.Split(p.Addr.String(), ":")[0]

		allowed, err := c.Allow(ctx, clientIP, limitRPM)
		if err != nil {
			fmt.Printf("[WARN] Rate limiter error: %v\n", err)
			return handler(ctx, req)
		}

		if !allowed {
			return nil, status.Errorf(codes.ResourceExhausted,
				"rate limit exceeded: max %d requests per minute for IP %s", limitRPM, clientIP)
		}

		return handler(ctx, req)
	}
}
