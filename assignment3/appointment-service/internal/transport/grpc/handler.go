package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment3/appointment-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment3/appointment-service/internal/usecase"
	pb "github.com/CoffeeSi/golang2AITU/assignment3/appointment-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AppointmentHandler struct {
	pb.UnimplementedAppointmentServiceServer
	uc usecase.AppointmentUsecaseInterface
}

func RegisterAppointmentHandlers(s *grpc.Server, uc usecase.AppointmentUsecaseInterface) *AppointmentHandler {
	handler := &AppointmentHandler{uc: uc}
	pb.RegisterAppointmentServiceServer(s, handler)
	return handler
}

func (h *AppointmentHandler) CreateAppointment(ctx context.Context, request *pb.CreateAppointmentRequest) (*pb.AppointmentResponse, error) {
	newAppointment := &model.Appointment{
		Title:       request.Title,
		Description: request.Description,
		DoctorID:    request.DoctorId,
	}

	err := h.uc.CreateAppointment(ctx, newAppointment)
	if err != nil {
		if errors.Is(err, model.ServiceUnavailableError) {
			return nil, status.Error(codes.Unavailable, err.Error())
		}
		if errors.Is(err, model.InvalidArgumentCreateError) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, model.DoctorDoesNotExistError) {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.AppointmentResponse{
		Id:          newAppointment.ID,
		Title:       newAppointment.Title,
		Description: newAppointment.Description,
		DoctorId:    newAppointment.DoctorID,
		Status:      string(newAppointment.Status),
		CreatedAt:   newAppointment.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   newAppointment.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (h *AppointmentHandler) GetAppointment(ctx context.Context, request *pb.GetAppointmentRequest) (*pb.AppointmentResponse, error) {
	appointment_id := request.Id
	appointment, err := h.uc.GetAppointment(ctx, appointment_id)
	if err != nil {
		if errors.Is(err, model.AppointmentNotFoundError) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, model.InvalidArgumentGetError) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, model.InvalidIDFormatError) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.AppointmentResponse{
		Id:          appointment.ID,
		Title:       appointment.Title,
		Description: appointment.Description,
		DoctorId:    appointment.DoctorID,
		Status:      string(appointment.Status),
		CreatedAt:   appointment.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   appointment.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (h *AppointmentHandler) ListAppointments(ctx context.Context, request *pb.ListAppointmentsRequest) (*pb.ListAppointmentsResponse, error) {
	appointments, err := h.uc.ListAppointments(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	var pbAppointments []*pb.AppointmentResponse
	for _, appointment := range appointments {
		pbAppointments = append(pbAppointments, &pb.AppointmentResponse{
			Id:          appointment.ID,
			Title:       appointment.Title,
			Description: appointment.Description,
			DoctorId:    appointment.DoctorID,
			Status:      string(appointment.Status),
			CreatedAt:   appointment.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   appointment.UpdatedAt.Format(time.RFC3339),
		})
	}
	return &pb.ListAppointmentsResponse{Appointments: pbAppointments}, nil
}

func (h *AppointmentHandler) UpdateAppointmentStatus(ctx context.Context, request *pb.UpdateStatusRequest) (*pb.AppointmentResponse, error) {
	appointment_id, appointment_status := request.Id, request.Status
	appointment, err := h.uc.UpdateAppointmentStatus(ctx, appointment_id, appointment_status)
	if err != nil {
		if errors.Is(err, model.InvalidArgumentUpdateStatusError) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, model.InvalidIDFormatError) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, model.AppointmentNotFoundError) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, model.InvalidStatusError) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, model.DoneStatusTransitionError) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.AppointmentResponse{
		Id:          appointment.ID,
		Title:       appointment.Title,
		Description: appointment.Description,
		DoctorId:    appointment.DoctorID,
		Status:      string(appointment.Status),
		CreatedAt:   appointment.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   appointment.UpdatedAt.Format(time.RFC3339),
	}, nil
}
