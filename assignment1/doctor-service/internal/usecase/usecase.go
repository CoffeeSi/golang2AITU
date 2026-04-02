package usecase

import (
	"errors"

	"github.com/CoffeeSi/golang2AITU/assignment1/doctor-service/internal/repository"
	"github.com/CoffeeSi/golang2AITU/assignment1/doctor-service/internal/transport/http/dto"
)

type DoctorUsecase struct {
	repo repository.DoctorRepository
}

func NewDoctorUsecase(repo repository.DoctorRepository) DoctorUsecase {
	return DoctorUsecase{repo: repo}
}

func (uc *DoctorUsecase) CreateDoctor(request dto.CreateDoctorRequest) error {
	if request.FullName == "" || request.Specialization == "" || request.Email == "" {
		return errors.New("all fields are required")
	}
	return uc.repo.CreateDoctor(request)
}

func (uc *DoctorUsecase) GetDoctorByID(id string) (*dto.DoctorResponse, error) {
	return uc.repo.GetDoctorByID(id)
}

func (uc *DoctorUsecase) ListDoctors() ([]dto.DoctorResponse, error) {
	return uc.repo.ListDoctors()
}