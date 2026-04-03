package repository

import "github.com/CoffeeSi/golang2AITU/assignment1/appointment-service/internal/transport/http/dto"

type AppointmentRepositoryInterface interface {
	CreateAppointment(request dto.CreateAppointmentRequest) error
	GetAppointmentByID(id string) (*dto.AppointmentResponse, error)
	ListAppointments() ([]dto.AppointmentResponse, error)
	UpdateAppointmentStatus(id string, request dto.UpdateAppointmentStatusRequest) error
}