package subscriber

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment4/notification-service/internal/jobqueue"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

const exchangeName = "ap2.events"

type NotificationSubscriber struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	redis   *redis.Client
	wp      *jobqueue.WorkerPool
}

type NotificationRecord struct {
	Time    string         `json:"time"`
	Subject string         `json:"subject"`
	Event   map[string]any `json:"event"`
}

func NewSubscriber(amqpURL string, redisClient *redis.Client, workerPool *jobqueue.WorkerPool) (*NotificationSubscriber, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	return &NotificationSubscriber{
		conn:    conn,
		channel: channel,
		redis:   redisClient,
		wp:      workerPool,
	}, nil
}

func (s *NotificationSubscriber) Close() error {
	var errs []error

	if s.channel != nil {
		if err := s.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	return nil
}

func (s *NotificationSubscriber) Listen(ctx context.Context) error {
	err := s.channel.ExchangeDeclare(
		exchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	q, err := s.channel.QueueDeclare(
		"",
		false,
		true,
		true,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = s.channel.QueueBind(
		q.Name,
		"",
		exchangeName,
		false,
		nil,
	)

	if err != nil {
		return err
	}

	notifyClose := s.channel.NotifyClose(make(chan *amqp.Error, 1))
	go func() {
		if closeErr := <-notifyClose; closeErr != nil {
			fmt.Fprintf(os.Stderr, "rabbitmq channel closed: code=%d reason=%s server=%t recover=%t\n", closeErr.Code, closeErr.Reason, closeErr.Server, closeErr.Recover)
		}
	}()

	consumerTag := "notification-service"
	msgs, err := s.channel.Consume(
		q.Name,
		consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	shuttingDown := false
	for {
		select {
		case <-ctx.Done():
			if shuttingDown {
				continue
			}
			shuttingDown = true
			if err := s.channel.Cancel(consumerTag, false); err != nil {
				return err
			}
		case delivery, ok := <-msgs:
			if !ok {
				if shuttingDown || ctx.Err() != nil {
					return nil
				}
				return errors.New("notification consumer channel closed unexpectedly")
			}
			if err := s.handleDelivery(delivery); err != nil {
				fmt.Fprintf(os.Stderr, "failed to process notification: subject=%s err=%v\n", delivery.RoutingKey, err)
				if nackErr := delivery.Nack(false, false); nackErr != nil {
					return nackErr
				}
				continue
			}
			if err := delivery.Ack(false); err != nil {
				return err
			}
		}
	}
}

func (s *NotificationSubscriber) handleDelivery(delivery amqp.Delivery) error {
	var event map[string]any
	if err := json.Unmarshal(delivery.Body, &event); err != nil {
		return err
	}

	record := NotificationRecord{
		Time:    time.Now().UTC().Format(time.RFC3339),
		Subject: delivery.RoutingKey,
		Event:   event,
	}

	payload, err := json.Marshal(record)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, string(payload))

	s.HandleAppointmentStatusUpdated(context.Background(), delivery.Body)
	return nil
}

func (s *NotificationSubscriber) HandleAppointmentStatusUpdated(ctx context.Context, payload []byte) {
	var event struct {
		EventType  string `json:"event_type"`
		ID         string `json:"id"`
		DoctorID   string `json:"doctor_id"`
		OccurredAt string `json:"occurred_at"`
		NewStatus  string `json:"new_status"`
	}
	_ = json.Unmarshal(payload, &event)

	if event.NewStatus != "done" {
		return
	}

	raw := event.EventType + event.ID + event.OccurredAt
	hash := sha256.Sum256([]byte(raw))
	idKey := hex.EncodeToString(hash[:])

	exists, err := s.redis.Exists(ctx, "job:"+idKey).Result()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to check job existence: %v\n", err)
		return
	}
	if exists > 0 {
		return
	}

	job := jobqueue.NotificationJob{
		IdempotencyKey: idKey,
		AppointmentID:  event.ID,
		DoctorID:       event.DoctorID,
		OccurredAt:     event.OccurredAt,
		Channel:        "email",
		Recipient:      "patient@clinic.kz",
		Message:        fmt.Sprintf("Your appointment %s with doctor %s is complete.", event.ID, event.DoctorID),
	}

	s.wp.Enqueue(job)
}
