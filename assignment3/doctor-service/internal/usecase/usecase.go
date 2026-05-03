package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment3/doctor-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment3/doctor-service/internal/repository"
	"github.com/google/uuid"
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
	// if _, err := uc.repo.GetDoctorByEmail(ctx, doctor.Email); err == nil {
	// 	return model.DoctorAlreadyExistsError
	// }
	doctor.ID = uuid.New().String()
	doctor.CreatedAt = time.Now()
	err := uc.repo.CreateDoctor(ctx, doctor)
	if err != nil {
		if errors.Is(err, model.DoctorAlreadyExistsError) {
			return model.DoctorAlreadyExistsError
		}
		return err
	}
	return nil
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
