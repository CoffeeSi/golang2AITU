package repository

import (
	"context"

	"github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/internal/model"
	"gorm.io/gorm"
)

type DoctorRepository struct {
	db *gorm.DB
}

func NewDoctorRepository(db *gorm.DB) DoctorRepository {
	return DoctorRepository{db: db}
}

func (d DoctorRepository) CreateDoctor(ctx context.Context, doctor *model.Doctor) error {
	return d.db.WithContext(ctx).Create(doctor).Error
}

func (d DoctorRepository) GetDoctorByID(ctx context.Context, id string) (*model.Doctor, error) {
	var doctor model.Doctor
	if err := d.db.WithContext(ctx).First(&doctor, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &doctor, nil
}

func (d DoctorRepository) ListDoctors(ctx context.Context) ([]*model.Doctor, error) {
	var doctors []*model.Doctor
	if err := d.db.WithContext(ctx).Find(&doctors).Error; err != nil {
		return nil, err
	}

	return doctors, nil
}

func (d DoctorRepository) GetDoctorByEmail(ctx context.Context, email string) (*model.Doctor, error) {
	var doctor model.Doctor
	if err := d.db.WithContext(ctx).First(&doctor, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &doctor, nil
}
