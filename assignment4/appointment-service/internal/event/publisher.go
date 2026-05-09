package event

import (
	"context"
	"encoding/json"

	amqp "github.com/rabbitmq/amqp091-go"
)

type EventPublisher struct {
	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewPublisher(amqpURL string) (*EventPublisher, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}
	err = channel.ExchangeDeclare(
		"ap2.events",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}
	return &EventPublisher{
		conn:    conn,
		channel: channel,
	}, nil
}

func (p *EventPublisher) Publish(ctx context.Context, routingKey string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	return p.channel.PublishWithContext(ctx,
		"ap2.events",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}

func (p *EventPublisher) Close() error {
	if err := p.channel.Close(); err != nil {
		return err
	}
	return p.conn.Close()
}
