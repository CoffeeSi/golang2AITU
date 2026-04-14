package app

import (
	"fmt"
	"net"
	"os"

	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/client"
	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/repository"
	grpc_handler "github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/transport/grpc"
	"github.com/CoffeeSi/golang2AITU/assignment2/appointment-service/internal/usecase"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Run(port string) error {
	godotenv.Load()

	// Retrieve doctor service URL
	doctorServiceURL := os.Getenv("DOCTOR_SERVICE_URL")
	if doctorServiceURL == "" {
		doctorServiceURL = "localhost:8080"
	}

	// Database url configuration
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_TIMEZONE"),
	)

	// Database connection
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	db.AutoMigrate(&model.Appointment{})

	conn, err := grpc.NewClient(doctorServiceURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	doctorClient := client.NewDoctorClient(conn)

	// Repository initialization
	repo := repository.NewAppointmentRepository(db)

	// Usecase initialization
	uc := usecase.NewAppointmentUsecase(repo, doctorClient)

	server, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	// gRPC server setup
	grpcServer := grpc.NewServer()
	grpc_handler.RegisterAppointmentHandlers(grpcServer, &uc)
	reflection.Register(grpcServer)

	fmt.Printf("Server running on localhost%s\n", port)

	return grpcServer.Serve(server)
}
