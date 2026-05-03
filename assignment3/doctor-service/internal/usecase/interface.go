package usecase

import (
	"context"

	"github.com/CoffeeSi/golang2AITU/assignment3/doctor-service/internal/model"
)

type DoctorUsecaseInterface interface {
	CreateDoctor(ctx context.Context, doctor *model.Doctor) error
	GetDoctor(ctx context.Context, id string) (*model.Doctor, error)
	ListDoctors(ctx context.Context) ([]*model.Doctor, error)
}
