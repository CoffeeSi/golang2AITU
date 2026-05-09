package usecase

import (
	"context"
	"log"
	"slices"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment4/appointment-service/internal/cache"
	"github.com/CoffeeSi/golang2AITU/assignment4/appointment-service/internal/event"
	"github.com/CoffeeSi/golang2AITU/assignment4/appointment-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment4/appointment-service/internal/repository"
	"github.com/google/uuid"
)

type AppointmentUsecase struct {
	repo         repository.AppointmentRepositoryInterface
	doctorClient DoctorClientInterface
	publisher    event.AppointmentEventPublisherInterface
	cache        cache.CacheClientInterface
}

func NewAppointmentUsecase(
	repo repository.AppointmentRepositoryInterface,
	doctorClient DoctorClientInterface,
	publisher event.AppointmentEventPublisherInterface,
	cache cache.CacheClientInterface,
) AppointmentUsecase {
	return AppointmentUsecase{
		repo:         repo,
		doctorClient: doctorClient,
		publisher:    publisher,
		cache:        cache,
	}
}

func (uc *AppointmentUsecase) CreateAppointment(ctx context.Context, appointment *model.Appointment) error {
	if appointment.Title == "" || appointment.DoctorID == "" {
		return model.InvalidArgumentCreateError
	}
	if err := uc.doctorClient.DoctorExists(ctx, appointment.DoctorID); err != nil {
		if err == model.ServiceUnavailableError {
			log.Printf("[ERROR] doctor service dependency failure: doctor_id=%s err=%v", appointment.DoctorID, err)
		}
		return err
	}
	appointment.ID = uuid.New().String()
	appointment.Status = model.StatusNew
	appointment.CreatedAt = time.Now()
	appointment.UpdatedAt = time.Now()
	err := uc.repo.CreateAppointment(ctx, appointment)
	if err != nil {
		return err
	}

	if err := uc.cache.DeleteList(ctx); err != nil {
		log.Printf("%s", err.Error())
	}
	if err := uc.cache.Set(ctx, appointment); err != nil {
		log.Printf("%s", err.Error())
	}

	event_payload := event.AppointmentCreatedEvent{
		EventType:  event.AppointmentCreatedEventType,
		OccurredAt: time.Now().Format(time.RFC3339),
		ID:         appointment.ID,
		Title:      appointment.Title,
		DoctorID:   appointment.DoctorID,
		Status:     string(appointment.Status),
	}

	err = uc.publisher.Publish(ctx, event.AppointmentCreatedEventType, event_payload)
	if err != nil {
		log.Printf("[ERROR] failed to publish appointment created event: %v", err)
	}
	return nil
}

func (uc *AppointmentUsecase) GetAppointment(ctx context.Context, id string) (*model.Appointment, error) {
	if id == "" {
		return nil, model.InvalidArgumentGetError
	}
	if _, err := uuid.Parse(id); err != nil {
		return nil, model.InvalidIDFormatError
	}

	var appointment *model.Appointment

	appointment, err := uc.cache.Get(ctx, id)
	if err == nil {
		return appointment, nil
	}

	appointment, err = uc.repo.GetAppointment(ctx, id)
	if err != nil {
		return nil, err
	}

	err = uc.cache.Set(ctx, appointment)
	if err != nil {
		log.Printf("%s", err.Error())
	}

	return appointment, nil
}

func (uc *AppointmentUsecase) ListAppointments(ctx context.Context) ([]*model.Appointment, error) {
	var appointments []*model.Appointment
	appointments, err := uc.cache.GetList(ctx)
	if err == nil {
		return appointments, nil
	}

	appointments, err = uc.repo.ListAppointments(ctx)
	if err != nil {
		return nil, err
	}

	err = uc.cache.SetList(ctx, appointments)
	if err != nil {
		log.Printf("%s", err.Error())
	}

	return appointments, nil
}

func (uc *AppointmentUsecase) UpdateAppointmentStatus(ctx context.Context, id string, status string) (*model.Appointment, error) {
	if id == "" || status == "" {
		return nil, model.InvalidArgumentUpdateStatusError
	}
	if _, err := uuid.Parse(id); err != nil {
		return nil, model.InvalidIDFormatError
	}

	appointment, err := uc.GetAppointment(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := uc.doctorClient.DoctorExists(ctx, appointment.DoctorID); err != nil {
		if err == model.ServiceUnavailableError {
			log.Printf("[ERROR] doctor service dependency failure: doctor_id=%s err=%v", appointment.DoctorID, err)
		}
		return nil, err
	}

	if !slices.Contains(model.Statuses, model.Status(status)) {
		return nil, model.InvalidStatusError
	}

	if appointment.Status == model.StatusDone && status == string(model.StatusDone) {
		return nil, model.DoneStatusTransitionError
	}

	updated_appointment, err := uc.repo.UpdateAppointmentStatus(ctx, id, status)
	if err != nil {
		return nil, err
	}

	err = uc.cache.Update(ctx, updated_appointment)
	if err != nil {
		log.Printf("%s", err.Error())
	}

	event_payload := event.AppointmentStatusUpdatedEvent{
		EventType:  event.AppointmentStatusUpdatedEventType,
		OccurredAt: time.Now().Format(time.RFC3339),
		ID:         id,
		OldStatus:  string(appointment.Status),
		NewStatus:  status,
	}
	err = uc.publisher.Publish(ctx, event.AppointmentStatusUpdatedEventType, event_payload)
	if err != nil {
		log.Printf("[ERROR] failed to publish appointment status updated event: %v", err)
	}
	return updated_appointment, nil
}
