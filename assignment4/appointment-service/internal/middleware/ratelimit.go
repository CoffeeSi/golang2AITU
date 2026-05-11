package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment4/appointment-service/internal/usecase"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func RateLimitInterceptor(c usecase.RateLimiterInterface, limitRPM int) grpc.UnaryServerInterceptor {
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
			retryAfter := 60
			if d, retryErr := c.RetryAfter(ctx, clientIP); retryErr == nil {
				retryAfter = int((d + time.Second - 1) / time.Second)
				if retryAfter < 1 {
					retryAfter = 1
				}
			}
			return nil, status.Errorf(codes.ResourceExhausted,
				"rate limit exceeded: max %d requests per minute for IP %s; retry after %ds", limitRPM, clientIP, retryAfter)
		}

		return handler(ctx, req)
	}
}
