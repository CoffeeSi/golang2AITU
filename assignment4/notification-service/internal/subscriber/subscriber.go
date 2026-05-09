package subscriber

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const exchangeName = "ap2.events"

type NotificationSubscriber struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

type NotificationRecord struct {
	Time    string         `json:"time"`
	Subject string         `json:"subject"`
	Event   map[string]any `json:"event"`
}

func NewSubscriber(amqpURL string) (*NotificationSubscriber, error) {
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
		if shuttingDown {
			delivery, ok := <-msgs
			if !ok {
				return nil
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
			continue
		}

		select {
		case <-ctx.Done():
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
	return nil
}
