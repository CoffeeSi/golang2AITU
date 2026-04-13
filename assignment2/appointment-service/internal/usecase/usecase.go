package usecase

import (
	"context"
	"errors"
	"slices"

	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/repository"
	pb "github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/proto"
	pb_doctor "github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ServiceUnavailableError = errors.New("doctor service is temporarily unavailable")

type AppointmentUsecase struct {
	repo             repository.AppointmentRepositoryInterface
	doctorServiceURL string
	doctorClient     pb_doctor.DoctorServiceClient
}

func NewAppointmentUsecase(repo repository.AppointmentRepositoryInterface, doctorServiceURL string, conn *grpc.ClientConn) AppointmentUsecase {
	return AppointmentUsecase{
		repo:             repo,
		doctorServiceURL: doctorServiceURL,
		doctorClient:     pb_doctor.NewDoctorServiceClient(conn),
	}
}

func (uc *AppointmentUsecase) CreateAppointment(ctx context.Context, appointment *model.Appointment) error {
	if err := uc.doctorExists(ctx, appointment.DoctorID); err != nil {
		return err
	}
	return uc.repo.CreateAppointment(appointment)
}

func (uc *AppointmentUsecase) GetAppointment(request *pb.GetAppointmentRequest) (*model.Appointment, error) {
	return uc.repo.GetAppointment(request)
}

func (uc *AppointmentUsecase) ListAppointments() ([]*model.Appointment, error) {
	return uc.repo.ListAppointments()
}

func (uc *AppointmentUsecase) UpdateAppointmentStatus(request *pb.UpdateStatusRequest) (*model.Appointment, error) {
	if request.Id == "" || request.Status == "" {
		return nil, status.Error(codes.InvalidArgument, "ID and status are required")
	}

	appointment, err := uc.repo.GetAppointment(&pb.GetAppointmentRequest{Id: request.Id})
	if err != nil {
		return nil, err
	}
	if appointment == nil {
		return nil, status.Error(codes.NotFound, "appointment not found")
	}

	if !slices.Contains(model.Statuses, model.Status(request.Status)) {
		return nil, status.Error(codes.InvalidArgument, "invalid status")
	}

	if appointment.Status == "done" {
		return nil, status.Error(codes.InvalidArgument, "transitioning a status from done back to new is not allowed")
	}
	return uc.repo.UpdateAppointmentStatus(request)
}

func (uc *AppointmentUsecase) doctorExists(ctx context.Context, doctorID string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	_, err := uc.doctorClient.GetDoctor(ctx, &pb_doctor.GetDoctorRequest{Id: doctorID})
	if err == nil {
		return nil
	}

	grpcStatus, ok := status.FromError(err)
	if !ok {
		return ServiceUnavailableError
	}

	if grpcStatus.Code() == codes.NotFound {
		return status.Error(codes.NotFound, "doctor not found")
	}

	return ServiceUnavailableError
}
