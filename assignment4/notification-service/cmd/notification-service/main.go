package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment4/notification-service/internal/subscriber"
)

const maxStartupAttempts = 5

func main() {
	brokerURL, err := brokerURLFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var sub *subscriber.NotificationSubscriber
	for attempt := 1; attempt <= maxStartupAttempts; attempt++ {
		sub, err = subscriber.NewSubscriber(brokerURL)
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

func brokerURLFromEnv() (string, error) {
	if url := os.Getenv("AMQP_URL"); url != "" {
		return url, nil
	}
	return "", fmt.Errorf("AMQP_URL must be set")
}

func startupBackoff(attempt int) time.Duration {
	return time.Duration(1<<uint(attempt-1)) * time.Second
}
