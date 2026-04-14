package usecase

import (
	"context"

	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/model"
)

type AppointmentUsecaseInterface interface {
	CreateAppointment(ctx context.Context, appointment *model.Appointment) error
	GetAppointment(ctx context.Context, id string) (*model.Appointment, error)
	ListAppointments(ctx context.Context) ([]*model.Appointment, error)
	UpdateAppointmentStatus(ctx context.Context, id string, status string) (*model.Appointment, error)
}
