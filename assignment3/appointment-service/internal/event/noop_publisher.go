package event

import "context"

type NoopPublisher struct{}

func (NoopPublisher) Publish(ctx context.Context, routingKey string, payload any) error {
	return nil
}