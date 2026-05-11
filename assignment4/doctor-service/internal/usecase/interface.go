package usecase

import (
	"context"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment4/doctor-service/internal/model"
)

type CacheClientInterface interface {
	Set(ctx context.Context, doctor *model.Doctor) error
	SetList(ctx context.Context, doctors []*model.Doctor) error
	Get(ctx context.Context, id string) (*model.Doctor, error)
	GetList(ctx context.Context) ([]*model.Doctor, error)
	Delete(ctx context.Context, id string) error
	DeleteList(ctx context.Context) error
}

type RateLimiterInterface interface {
	Allow(ctx context.Context, clientIP string, limit int) (bool, error)
	RetryAfter(ctx context.Context, clientIP string) (time.Duration, error)
}

type DoctorUsecaseInterface interface {
	CreateDoctor(ctx context.Context, doctor *model.Doctor) error
	GetDoctor(ctx context.Context, id string) (*model.Doctor, error)
	ListDoctors(ctx context.Context) ([]*model.Doctor, error)
}
