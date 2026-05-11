package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment4/notification-service/internal/jobqueue"
	"github.com/CoffeeSi/golang2AITU/assignment4/notification-service/internal/subscriber"
	"github.com/redis/go-redis/v9"
)

const maxStartupAttempts = 5

func main() {
	brokerURL := brokerURLFromEnv()
	redisURL := redisURLFromEnv()

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse REDIS_URL: %v\n", err)
		os.Exit(1)
	}

	redisClient := redis.NewClient(opt)
	defer redisClient.Close()

	// Check Redis availability with a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("[WARNING] Redis is unavailable at startup: %v. Notification service will continue without idempotency checks.", err)
	}
	cancel()

	gatewayURL := "http://localhost:8080/notify"
	if url := os.Getenv("GATEWAY_URL"); url != "" {
		gatewayURL = url + "/notify"
	}

	workerPoolSize := 3
	if sizeStr := os.Getenv("WORKER_POOL_SIZE"); sizeStr != "" {
		if size, err := strconv.Atoi(sizeStr); err == nil && size > 0 {
			workerPoolSize = size
		}
	}

	// Use a larger buffer to prevent blocking the RabbitMQ subscriber
	// Calculate buffer size as 10x the worker pool size, minimum 100
	bufferSize := workerPoolSize * 10
	if bufferSize < 100 {
		bufferSize = 100
	}

	wp := jobqueue.NewWorkerPool(workerPoolSize, bufferSize, redisClient, gatewayURL)
	wp.Start(context.Background(), workerPoolSize)

	var sub *subscriber.NotificationSubscriber
	for attempt := 1; attempt <= maxStartupAttempts; attempt++ {
		sub, err = subscriber.NewSubscriber(brokerURL, redisClient, wp)
		if err == nil {
			break
		}

		if attempt == maxStartupAttempts {
			fmt.Fprintf(os.Stderr, "failed to connect to message broker after %d attempts: %v\n", attempt, err)
			os.Exit(1)
		}

		backoff := startupBackoff(attempt)
		fmt.Fprintf(os.Stderr, "message broker unavailable, retrying in %s: attempt=%d err=%v\n", backoff, attempt, err)
		time.Sleep(backoff)
	}
	defer func() {
		if closeErr := sub.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "failed to close notification subscriber: %v\n", closeErr)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := sub.Listen(ctx); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "notification service stopped with error: %v\n", err)
		os.Exit(1)
	}
}

func brokerURLFromEnv() string {
	if url := os.Getenv("AMQP_URL"); url != "" {
		return url
	}
	return "amqp://guest:guest@localhost/"
}

func redisURLFromEnv() string {
	if redisAddr := os.Getenv("REDIS_URL"); redisAddr != "" {
		return redisAddr
	}
	return "redis://localhost:6379"
}

func startupBackoff(attempt int) time.Duration {
	return time.Duration(1<<uint(attempt-1)) * time.Second
}
