package cache

import (
	"context"

	"github.com/CoffeeSi/golang2AITU/assignment4/appointment-service/internal/model"
)

type CacheClientInterface interface {
	Set(ctx context.Context, appointment *model.Appointment) error
	SetList(ctx context.Context, appointments []*model.Appointment) error
	Get(ctx context.Context, id string) (*model.Appointment, error)
	GetList(ctx context.Context) ([]*model.Appointment, error)
	Delete(ctx context.Context, id string) error
	DeleteList(ctx context.Context) error
	Update(ctx context.Context, appointment *model.Appointment) error
}
