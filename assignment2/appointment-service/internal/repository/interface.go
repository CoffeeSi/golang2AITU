package repository

import (
	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/model"
	pb "github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/proto"
)

type AppointmentRepositoryInterface interface {
	CreateAppointment(appointment *model.Appointment) error
	GetAppointment(request *pb.GetAppointmentRequest) (*model.Appointment, error)
	ListAppointments() ([]*model.Appointment, error)
	UpdateAppointmentStatus(request *pb.UpdateStatusRequest) (*model.Appointment, error)
}