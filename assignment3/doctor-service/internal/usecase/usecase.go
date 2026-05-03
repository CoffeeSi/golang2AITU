package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment3/doctor-service/internal/event"
	"github.com/CoffeeSi/golang2AITU/assignment3/doctor-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment3/doctor-service/internal/repository"
	"github.com/google/uuid"
)

type DoctorUsecase struct {
	repo      repository.DoctorRepositoryInterface
	publisher event.DoctorEventPublisherInterface
}

func NewDoctorUsecase(repo repository.DoctorRepositoryInterface, publisher event.DoctorEventPublisherInterface) DoctorUsecase {
	return DoctorUsecase{repo: repo, publisher: publisher}
}

func (uc *DoctorUsecase) CreateDoctor(ctx context.Context, doctor *model.Doctor) error {
	if doctor.FullName == "" || doctor.Email == "" {
		return model.InvalidCreateArgumentError
	}

	doctor.ID = uuid.New().String()
	doctor.CreatedAt = time.Now()
	err := uc.repo.CreateDoctor(ctx, doctor)
	if err != nil {
		if errors.Is(err, model.DoctorAlreadyExistsError) {
			return model.DoctorAlreadyExistsError
		}
		return err
	}

	event_payload := event.DoctorCreatedEvent{
		EventType:      event.DoctorCreatedEventType,
		OccurredAt:     time.Now().Format(time.RFC3339),
		ID:             doctor.ID,
		FullName:       doctor.FullName,
		Specialization: doctor.Specialization,
		Email:          doctor.Email,
	}

	err = uc.publisher.Publish(ctx, event.DoctorCreatedEventType, event_payload)
	if err != nil {
		log.Printf("[ERROR] failed to publish doctor created event: %v", err)
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
