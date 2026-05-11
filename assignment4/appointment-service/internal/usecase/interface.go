package usecase

import (
	"context"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment4/appointment-service/internal/model"
)

type DoctorClientInterface interface {
	DoctorExists(ctx context.Context, doctorID string) error
}

type CacheClientInterface interface {
	Set(ctx context.Context, appointment *model.Appointment) error
	SetList(ctx context.Context, appointments []*model.Appointment) error
	Get(ctx context.Context, id string) (*model.Appointment, error)
	GetList(ctx context.Context) ([]*model.Appointment, error)
	Delete(ctx context.Context, id string) error
	DeleteList(ctx context.Context) error
	Update(ctx context.Context, appointment *model.Appointment) error
}

type RateLimiterInterface interface {
	Allow(ctx context.Context, clientIP string, limit int) (bool, error)
	RetryAfter(ctx context.Context, clientIP string) (time.Duration, error)
}

type AppointmentUsecaseInterface interface {
	CreateAppointment(ctx context.Context, appointment *model.Appointment) error
	GetAppointment(ctx context.Context, id string) (*model.Appointment, error)
	ListAppointments(ctx context.Context) ([]*model.Appointment, error)
	UpdateAppointmentStatus(ctx context.Context, id string, status string) (*model.Appointment, error)
}
