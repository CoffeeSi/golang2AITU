package cache

import (
	"context"

	"github.com/CoffeeSi/golang2AITU/assignment4/doctor-service/internal/model"
)

type CacheClientInterface interface {
	Set(ctx context.Context, doctor *model.Doctor) error
	SetList(ctx context.Context, doctors []*model.Doctor) error
	Get(ctx context.Context, id string) (*model.Doctor, error)
	GetList(ctx context.Context) ([]*model.Doctor, error)
	Delete(ctx context.Context, id string) error
	DeleteList(ctx context.Context) error
	Allow(ctx context.Context, clientIP string, limit int) (bool, error)
}
