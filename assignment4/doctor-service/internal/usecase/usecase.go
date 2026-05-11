package usecase

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment4/doctor-service/internal/event"
	"github.com/CoffeeSi/golang2AITU/assignment4/doctor-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment4/doctor-service/internal/repository"
	"github.com/google/uuid"
)

type DoctorUsecase struct {
	repo      repository.DoctorRepositoryInterface
	publisher event.DoctorEventPublisherInterface
	cache     CacheClientInterface
}

func NewDoctorUsecase(
	repo repository.DoctorRepositoryInterface,
	publisher event.DoctorEventPublisherInterface,
	cache CacheClientInterface,
) DoctorUsecase {
	return DoctorUsecase{repo: repo, publisher: publisher, cache: cache}
}

func (uc *DoctorUsecase) CreateDoctor(ctx context.Context, doctor *model.Doctor) error {
	if doctor.FullName == "" || doctor.Email == "" {
		return model.InvalidCreateArgumentError
	}

	doctor.ID = uuid.New().String()
	doctor.CreatedAt = time.Now()

	err := uc.repo.CreateDoctor(ctx, doctor)
	if err != nil {
		return err
	}

	if err := uc.cache.DeleteList(ctx); err != nil {
		log.Printf("%s", err.Error())
	}
	if err := uc.cache.Set(ctx, doctor); err != nil {
		log.Printf("%s", err.Error())
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

	var doctor *model.Doctor

	doctor, err := uc.cache.Get(ctx, id)
	if err == nil {
		return doctor, nil
	}

	doctor, err = uc.repo.GetDoctorByID(ctx, id)
	if err != nil {
		return nil, err
	}

	err = uc.cache.Set(ctx, doctor)
	if err != nil {
		log.Printf("%s", err.Error())
	}

	return doctor, nil
}

func (uc *DoctorUsecase) ListDoctors(ctx context.Context) ([]*model.Doctor, error) {
	var doctors []*model.Doctor

	doctors, err := uc.cache.GetList(ctx)
	if err == nil {
		return doctors, nil
	}

	if !errors.Is(err, model.RedisCacheMissError) {
		log.Printf("%s", err.Error())
	}

	doctors, err = uc.repo.ListDoctors(ctx)
	if err != nil {
		return nil, err
	}

	err = uc.cache.SetList(ctx, doctors)
	if err != nil {
		log.Printf("%s", err.Error())
	}
	return doctors, nil
}
