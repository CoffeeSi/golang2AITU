package repository

import (
	"context"

	"github.com/CoffeeSi/golang2AITU/assignment3/appointment-service/internal/model"
)

type AppointmentRepositoryInterface interface {
	CreateAppointment(ctx context.Context, appointment *model.Appointment) error
	GetAppointment(ctx context.Context, id string) (*model.Appointment, error)
	ListAppointments(ctx context.Context) ([]*model.Appointment, error)
	UpdateAppointmentStatus(ctx context.Context, id string, appointment_status string) (*model.Appointment, error)
}
