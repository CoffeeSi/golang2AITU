package event

import "context"

type DoctorEventPublisherInterface interface {
	Publish(ctx context.Context, routingKey string, payload any) error
}
