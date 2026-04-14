package repository

import (
	"context"
	"errors"

	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/model"
	"gorm.io/gorm"
)

type AppointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return AppointmentRepository{db: db}
}

func (r AppointmentRepository) CreateAppointment(ctx context.Context, appointment *model.Appointment) error {
	return r.db.WithContext(ctx).Create(&appointment).Error
}

func (r AppointmentRepository) GetAppointment(ctx context.Context, id string) (*model.Appointment, error) {
	var appointment model.Appointment
	if err := r.db.WithContext(ctx).First(&appointment, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, model.AppointmentNotFoundError
		}
		return nil, err
	}
	return &appointment, nil
}

func (r AppointmentRepository) ListAppointments(ctx context.Context) ([]*model.Appointment, error) {
	var appointments []*model.Appointment
	if err := r.db.WithContext(ctx).Find(&appointments).Error; err != nil {
		return nil, err
	}

	return appointments, nil
}

func (r AppointmentRepository) UpdateAppointmentStatus(ctx context.Context, id string, appointment_status string) (*model.Appointment, error) {
	var appointment model.Appointment
	res := r.db.WithContext(ctx).Model(&appointment).Where("id = ?", id).Update("status", appointment_status).First(&appointment)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, model.AppointmentNotFoundError
	}
	return &appointment, nil
}
