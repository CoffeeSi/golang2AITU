package repository

import (
	"github.com/CoffeeSi/golang2AITU/assignment1/doctor-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment1/doctor-service/internal/transport/http/dto"
	"gorm.io/gorm"
)

type DoctorRepository struct {
	db *gorm.DB
}

func NewDoctorRepository(db *gorm.DB) DoctorRepository {
	return DoctorRepository{db: db}
}

func (d DoctorRepository) CreateDoctor(request dto.CreateDoctorRequest) error {
	doctor := model.Doctor{
		FullName:       request.FullName,
		Specialization: request.Specialization,
		Email:          request.Email,
	}

	return d.db.Create(&doctor).Error
}

func (d DoctorRepository) GetDoctorByID(id string) (*dto.DoctorResponse, error) {
	var doctor model.Doctor
	if err := d.db.First(&doctor, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &dto.DoctorResponse{
		ID:             doctor.ID,
		FullName:       doctor.FullName,
		Specialization: doctor.Specialization,
		Email:          doctor.Email,
	}, nil
}

func (d DoctorRepository) ListDoctors() ([]dto.DoctorResponse, error) {
	var doctors []model.Doctor
	if err := d.db.Find(&doctors).Error; err != nil {
		return nil, err
	}

	var response []dto.DoctorResponse
	for _, doctor := range doctors {
		response = append(response, dto.DoctorResponse{
			ID:             doctor.ID,
			FullName:       doctor.FullName,
			Specialization: doctor.Specialization,
			Email:          doctor.Email,
		})
	}
	return response, nil
}
