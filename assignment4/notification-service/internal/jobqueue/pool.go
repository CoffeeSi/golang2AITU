package jobqueue

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type NotificationJob struct {
	IdempotencyKey string `json:"idempotency_key"`
	AppointmentID  string `json:"appointment_id"`
	DoctorID       string `json:"doctor_id"`
	OccurredAt     string `json:"occurred_at"`
	Channel        string `json:"channel"`
	Recipient      string `json:"recipient"`
	Message        string `json:"message"`

	Attempts int `json:"-"`
}

type WorkerPool struct {
	jobChan    chan NotificationJob
	redis      *redis.Client
	gatewayURL string
}

func NewWorkerPool(poolSize int, redisClient *redis.Client, gatewayURL string) *WorkerPool {
	return &WorkerPool{
		jobChan:    make(chan NotificationJob, poolSize),
		redis:      redisClient,
		gatewayURL: gatewayURL,
	}
}

func (wp *WorkerPool) Start(ctx context.Context, size int) {
	for i := 0; i < size; i++ {
		go wp.worker(ctx)
	}
}

func (wp *WorkerPool) Enqueue(job NotificationJob) {
	if job.Attempts == 0 {
		job.Attempts = 1
	}
	logState("info", job.IdempotencyKey, "enqueued", "", job.Attempts)
	wp.jobChan <- job
}

func (wp *WorkerPool) worker(ctx context.Context) {
	for job := range wp.jobChan {
		wp.process(ctx, job)
	}
}

func (wp *WorkerPool) process(ctx context.Context, job NotificationJob) {
	if job.Attempts == 0 {
		job.Attempts = 1
	}
	logState("info", job.IdempotencyKey, "processing", "", job.Attempts)

	err := wp.sendToGateway(ctx, job)
	if err != nil {
		wp.handleRetry(job, err)
		return
	}

	wp.redis.Set(ctx, "job:"+job.IdempotencyKey, "done", 24*time.Hour)
	logState("info", job.IdempotencyKey, "success", "", job.Attempts)
}

func (p *WorkerPool) sendToGateway(ctx context.Context, job NotificationJob) error {
	jsonData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.gatewayURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusServiceUnavailable:
		return fmt.Errorf("gateway unavailable (503)")
	default:
		return fmt.Errorf("gateway returned unexpected status: %d", resp.StatusCode)
	}
}

func (wp *WorkerPool) handleRetry(job NotificationJob, err error) {
	if job.Attempts >= 3 {
		logState("error", job.IdempotencyKey, "dead_letter", err.Error(), job.Attempts)
		return
	}

	logState("warn", job.IdempotencyKey, "retry", err.Error(), job.Attempts)

	wait := time.Duration(math.Pow(2, float64(job.Attempts-1))) * time.Second

	go func() {
		time.Sleep(wait)
		job.Attempts++
		wp.Enqueue(job)
	}()
}

func logState(level, jobId, status, errMsg string, attempt int) {
	logLine := map[string]any{
		"time":    time.Now().Format(time.RFC3339),
		"level":   level,
		"job_id":  jobId,
		"attempt": attempt,
		"status":  status,
	}
	if errMsg != "" {
		logLine["error"] = errMsg
	}

	data, _ := json.Marshal(logLine)
	if level == "error" {
		fmt.Fprintln(os.Stderr, string(data))
	} else {
		fmt.Println(string(data))
	}
}
