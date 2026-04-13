package usecase

import (
	"context"

	"github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/internal/repository"
	pb "github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DoctorUsecase struct {
	repo repository.DoctorRepository
}

func NewDoctorUsecase(repo repository.DoctorRepository) DoctorUsecase {
	return DoctorUsecase{repo: repo}
}

func (uc *DoctorUsecase) CreateDoctor(ctx context.Context, doctor *model.Doctor) error {
	if doctor.FullName == "" || doctor.Specialization == "" || doctor.Email == "" {
		return status.Error(codes.InvalidArgument, "Full name, specialization and email are required")
	}
	if _, err := uc.repo.GetDoctorByEmail(ctx, doctor.Email); err == nil {
		return status.Error(codes.AlreadyExists, "Doctor with this email already exists")
	}
	return uc.repo.CreateDoctor(ctx, doctor)
}

func (uc *DoctorUsecase) GetDoctor(ctx context.Context, request *pb.GetDoctorRequest) (*model.Doctor, error) {
	if request.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "ID is required")
	}
	return uc.repo.GetDoctorByID(ctx, request.Id)
}

func (uc *DoctorUsecase) ListDoctors(ctx context.Context) ([]*model.Doctor, error) {
	return uc.repo.ListDoctors(ctx)
}
