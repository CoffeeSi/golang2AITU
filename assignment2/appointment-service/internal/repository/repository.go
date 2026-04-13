package repository

import (
	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/model"
	pb "github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type AppointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return AppointmentRepository{db: db}
}

func (r AppointmentRepository) CreateAppointment(appointment *model.Appointment) error {
	return r.db.Create(&appointment).Error
}

func (r AppointmentRepository) GetAppointment(request *pb.GetAppointmentRequest) (*model.Appointment, error) {
	var appointment model.Appointment
	err := r.db.First(&appointment, "id = ?", request.Id).Error
	if err == gorm.ErrRecordNotFound {
		return nil, status.Error(codes.NotFound, "appointment not found")
	} else if err != nil {
		return nil, err
	}
	return &appointment, nil
}

func (r AppointmentRepository) ListAppointments() ([]*model.Appointment, error) {
	var appointments []*model.Appointment
	if err := r.db.Find(&appointments).Error; err != nil {
		return nil, err
	}

	return appointments, nil
}

func (r AppointmentRepository) UpdateAppointmentStatus(request *pb.UpdateStatusRequest) (*model.Appointment, error) {
	var appointment model.Appointment
	err := r.db.Model(&appointment).Where("id = ?", request.Id).Update("status", request.Status).Error
	if err != nil {
		return nil, err
	}
	if appointment.ID == "" {
		return nil, status.Error(codes.NotFound, "appointment not found")
	}
	return &appointment, nil
}
