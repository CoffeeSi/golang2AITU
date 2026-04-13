package grpc

import (
	"context"

	"github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/internal/usecase"
	pb "github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DoctorHandler struct {
	pb.UnimplementedDoctorServiceServer
	uc *usecase.DoctorUsecase
}

func RegisterDoctorHandlers(s *grpc.Server, uc *usecase.DoctorUsecase) *DoctorHandler {
	handler := &DoctorHandler{uc: uc}
	pb.RegisterDoctorServiceServer(s, handler)
	return handler
}

func (h *DoctorHandler) CreateDoctor(ctx context.Context, request *pb.CreateDoctorRequest) (*pb.DoctorResponse, error) {
	newDoctor := model.Doctor{
		FullName:       request.FullName,
		Specialization: request.Specialization,
		Email:          request.Email,
	}

	err := h.uc.CreateDoctor(ctx, &newDoctor)
	if err != nil {
		return nil, err
	}
	return &pb.DoctorResponse{
		Id:             newDoctor.ID,
		FullName:       newDoctor.FullName,
		Specialization: newDoctor.Specialization,
		Email:          newDoctor.Email,
	}, nil
}

func (h *DoctorHandler) GetDoctor(ctx context.Context, request *pb.GetDoctorRequest) (*pb.DoctorResponse, error) {
	doctor, err := h.uc.GetDoctor(ctx, request)
	if err != nil {
		return nil, err
	} else if doctor == nil {
		return nil, status.Error(codes.NotFound, "Doctor not found")
	}
	return &pb.DoctorResponse{
		Id:             doctor.ID,
		FullName:       doctor.FullName,
		Specialization: doctor.Specialization,
		Email:          doctor.Email,
	}, nil
}

func (h *DoctorHandler) ListDoctors(ctx context.Context, request *pb.ListDoctorsRequest) (*pb.ListDoctorsResponse, error) {
	doctors, err := h.uc.ListDoctors(ctx)
	if err != nil {
		return nil, err
	}
	var doctorResponses []*pb.DoctorResponse
	for _, doctor := range doctors {
		doctorResponses = append(doctorResponses, &pb.DoctorResponse{
			Id:             doctor.ID,
			FullName:       doctor.FullName,
			Specialization: doctor.Specialization,
			Email:          doctor.Email,
		})
	}
	return &pb.ListDoctorsResponse{Doctors: doctorResponses}, nil
}
