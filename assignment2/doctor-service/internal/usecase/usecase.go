package usecase

import (
	"context"

	"github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/internal/repository"
	pb "github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/proto"
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

func (uc *DoctorUsecase) GetDoctor(ctx context.Context, request *pb.GetDoctorRequest) (*model.Doctor, error) {
	if request.Id == "" {
		return nil, model.InvalidGetArgumentError
	}
	return uc.repo.GetDoctorByID(ctx, request.Id)
}

func (uc *DoctorUsecase) ListDoctors(ctx context.Context) ([]*model.Doctor, error) {
	return uc.repo.ListDoctors(ctx)
}
