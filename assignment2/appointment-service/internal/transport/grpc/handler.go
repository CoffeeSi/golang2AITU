package grpc

import (
	"context"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/usecase"
	pb "github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AppointmentHandler struct {
	pb.UnimplementedAppointmentServiceServer
	uc *usecase.AppointmentUsecase
}

func RegisterAppointmentHandlers(s *grpc.Server, uc *usecase.AppointmentUsecase) *AppointmentHandler {
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
		return nil, err
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
	appointment, err := h.uc.GetAppointment(request)
	if err != nil {
		return nil, err
	}
	if appointment == nil {
		return nil, status.Error(codes.NotFound, "appointment not found")
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
	appointments, err := h.uc.ListAppointments()
	if err != nil {
		return nil, err
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
	appointment, err := h.uc.UpdateAppointmentStatus(request)
	if err != nil {
		return nil, err
	}
	if appointment == nil {
		return nil, status.Error(codes.Internal, "updated appointment is empty")
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
