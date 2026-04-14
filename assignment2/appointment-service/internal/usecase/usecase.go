package usecase

import (
	"context"
	"log"
	"slices"

	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/repository"
	"github.com/google/uuid"
)

type AppointmentUsecase struct {
	repo         repository.AppointmentRepositoryInterface
	doctorClient DoctorClientInterface
}

func NewAppointmentUsecase(repo repository.AppointmentRepositoryInterface, doctorClient DoctorClientInterface) AppointmentUsecase {
	return AppointmentUsecase{
		repo:         repo,
		doctorClient: doctorClient,
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
	return uc.repo.CreateAppointment(ctx, appointment)
}

func (uc *AppointmentUsecase) GetAppointment(ctx context.Context, id string) (*model.Appointment, error) {
	if id == "" {
		return nil, model.InvalidArgumentGetError
	}
	if _, err := uuid.Parse(id); err != nil {
		return nil, model.InvalidIDFormatError
	}
	return uc.repo.GetAppointment(ctx, id)
}

func (uc *AppointmentUsecase) ListAppointments(ctx context.Context) ([]*model.Appointment, error) {
	return uc.repo.ListAppointments(ctx)
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

	if appointment.Status == "done" && status == "new" {
		return nil, model.DoneStatusTransitionError
	}
	return uc.repo.UpdateAppointmentStatus(ctx, id, status)
}
