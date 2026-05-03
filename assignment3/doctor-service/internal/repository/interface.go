package repository

import (
	"context"

	"github.com/CoffeeSi/golang2AITU/assignment3/doctor-service/internal/model"
)

type DoctorRepositoryInterface interface {
	CreateDoctor(ctx context.Context, doctor *model.Doctor) error
	GetDoctorByID(ctx context.Context, id string) (*model.Doctor, error)
	GetDoctorByEmail(ctx context.Context, email string) (*model.Doctor, error)
	ListDoctors(ctx context.Context) ([]*model.Doctor, error)
}
