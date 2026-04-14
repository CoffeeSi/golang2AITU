package client

import (
	"context"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/model"
	pb_doctor "github.com/CoffeeSi/golang2AITU/assignment2/proto/doctor-service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DoctorClient struct {
	pbDoctorClient pb_doctor.DoctorServiceClient
}

const defaultDoctorExistsTimeout = 2 * time.Second

func NewDoctorClient(grpcConn *grpc.ClientConn) DoctorClient {
	return DoctorClient{
		pbDoctorClient: pb_doctor.NewDoctorServiceClient(grpcConn),
	}
}

func (c *DoctorClient) DoctorExists(ctx context.Context, doctorID string) error {
	requestCtx := ctx
	cancel := func() {}
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		requestCtx, cancel = context.WithTimeout(ctx, defaultDoctorExistsTimeout)
	}
	defer cancel()

	_, err := c.pbDoctorClient.GetDoctor(requestCtx, &pb_doctor.GetDoctorRequest{Id: doctorID})
	if err != nil {
		grpcStatus, ok := status.FromError(err)
		if !ok {
			return model.ServiceUnavailableError
		}

		if grpcStatus.Code() == codes.NotFound {
			return model.DoctorDoesNotExistError
		}
		return model.ServiceUnavailableError

	}
	return nil
}
