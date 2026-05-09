package usecase

import (
	"context"

	"github.com/CoffeeSi/golang2AITU/assignment4/appointment-service/internal/model"
)

type DoctorClientInterface interface {
	DoctorExists(ctx context.Context, doctorID string) error
}

type AppointmentUsecaseInterface interface {
	CreateAppointment(ctx context.Context, appointment *model.Appointment) error
	GetAppointment(ctx context.Context, id string) (*model.Appointment, error)
	ListAppointments(ctx context.Context) ([]*model.Appointment, error)
	UpdateAppointmentStatus(ctx context.Context, id string, status string) (*model.Appointment, error)
}
