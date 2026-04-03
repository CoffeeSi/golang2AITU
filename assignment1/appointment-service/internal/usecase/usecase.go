package usecase

import (
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment1/appointment-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment1/appointment-service/internal/repository"
	"github.com/CoffeeSi/golang2AITU/assignment1/appointment-service/internal/transport/http/dto"
)

var ServiceUnavailableError = errors.New("doctor service is temporarily unavailable")

type AppointmentUsecase struct {
	repo             repository.AppointmentRepositoryInterface
	doctorServiceURL string
	httpClient       *http.Client
}

func NewAppointmentUsecase(repo repository.AppointmentRepositoryInterface, doctorServiceURL string) AppointmentUsecase {
	return AppointmentUsecase{
		repo:             repo,
		doctorServiceURL: doctorServiceURL,
		httpClient:       &http.Client{Timeout: 5 * time.Second},
	}
}

func (uc *AppointmentUsecase) CreateAppointment(request dto.CreateAppointmentRequest) error {
	if err := uc.doctorExists(request.DoctorID); err != nil {
		return err
	}
	return uc.repo.CreateAppointment(request)
}

func (uc *AppointmentUsecase) GetAppointmentByID(id string) (*dto.AppointmentResponse, error) {
	return uc.repo.GetAppointmentByID(id)
}

func (uc *AppointmentUsecase) ListAppointments() ([]dto.AppointmentResponse, error) {
	return uc.repo.ListAppointments()
}

func (uc *AppointmentUsecase) UpdateAppointmentStatus(id string, request dto.UpdateAppointmentStatusRequest) error {
	appointment, err := uc.repo.GetAppointmentByID(id)
	if err != nil {
		return err
	}
	if appointment == nil {
		return fmt.Errorf("appointment not found")
	}

	if err := uc.doctorExists(appointment.DoctorID); err != nil {
		return err
	}

	if !slices.Contains(model.Statuses, request.Status) {
		return fmt.Errorf("invalid status: %s", request.Status)
	}

	if appointment.Status == "done" {
		return fmt.Errorf("transitioning a status from done back to new is not allowed")
	}
	return uc.repo.UpdateAppointmentStatus(id, request)
}

func (uc *AppointmentUsecase) doctorExists(doctorID string) error {
	resp, err := uc.httpClient.Get(fmt.Sprintf("%s/doctors/%s", uc.doctorServiceURL, doctorID))
	if err != nil {
		return ServiceUnavailableError
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("doctor with id %s not found", doctorID)
	}

	return ServiceUnavailableError
}
