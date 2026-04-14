package usecase

import (
	"context"

	"github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/internal/repository"
)

type DoctorUsecase struct {
	repo repository.DoctorRepositoryInterface
}

func NewDoctorUsecase(repo repository.DoctorRepositoryInterface) DoctorUsecase {
	return DoctorUsecase{repo: repo}
}

func (uc *DoctorUsecase) CreateDoctor(ctx context.Context, doctor *model.Doctor) error {
	if doctor.FullName == "" || doctor.Email == "" {
		return model.InvalidCreateArgumentError
	}
	if _, err := uc.repo.GetDoctorByEmail(ctx, doctor.Email); err == nil {
		return model.DoctorAlreadyExistsError
	}
	return uc.repo.CreateDoctor(ctx, doctor)
}

func (uc *DoctorUsecase) GetDoctor(ctx context.Context, id string) (*model.Doctor, error) {
	if id == "" {
		return nil, model.InvalidGetArgumentError
	}
	return uc.repo.GetDoctorByID(ctx, id)
}

func (uc *DoctorUsecase) ListDoctors(ctx context.Context) ([]*model.Doctor, error) {
	return uc.repo.ListDoctors(ctx)
}
