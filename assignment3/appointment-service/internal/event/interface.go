package event

import "context"

type AppointmentEventPublisherInterface interface {
	Publish(ctx context.Context, routingKey string, payload any) error
}
