package repository

import (
	"github.com/CoffeeSi/golang2AITU/assignment1/appointment-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment1/appointment-service/internal/transport/http/dto"
	"gorm.io/gorm"
)

type AppointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return AppointmentRepository{db: db}
}

func (r AppointmentRepository) CreateAppointment(request dto.CreateAppointmentRequest) error {
	appointment := model.Appointment{
		Title:       request.Title,
		Description: request.Description,
		DoctorID:    request.DoctorID,
		Status:      "new",
	}

	return r.db.Create(&appointment).Error
}

func (r AppointmentRepository) GetAppointmentByID(id string) (*dto.AppointmentResponse, error) {
	var appointment model.Appointment
	if err := r.db.First(&appointment, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &dto.AppointmentResponse{
		ID:             appointment.ID,
		Title:          appointment.Title,
		Description:    appointment.Description,
		DoctorID:       appointment.DoctorID,
		Status:         appointment.Status,
		CreatedAt:      appointment.CreatedAt,
		UpdatedAt:      appointment.UpdatedAt,
	}, nil
}

func (r AppointmentRepository) ListAppointments() ([]dto.AppointmentResponse, error) {
	var appointments []model.Appointment
	if err := r.db.Find(&appointments).Error; err != nil {
		return nil, err
	}

	var response []dto.AppointmentResponse
	for _, appointment := range appointments {
		response = append(response, dto.AppointmentResponse{
			ID:             appointment.ID,
			Title:          appointment.Title,
			Description:    appointment.Description,
			DoctorID:       appointment.DoctorID,
			Status:         appointment.Status,
			CreatedAt:      appointment.CreatedAt,
			UpdatedAt:      appointment.UpdatedAt,
		})
	}
	return response, nil
}

func (r AppointmentRepository) UpdateAppointmentStatus(id string, request dto.UpdateAppointmentStatusRequest) error {
	return r.db.Model(&model.Appointment{}).Where("id = ?", id).Update("status", request.Status).Error
}