package app

import (
	"fmt"
	"net"
	"os"

	"github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/internal/repository"
	grpc_handler "github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/internal/transport/grpc"
	"github.com/CoffeeSi/golang2AITU/assignment2/doctor-service/internal/usecase"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Run() error {
	godotenv.Load()

	// Port configuration
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	if port == ":" {
		port = ":8080"
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

	db.AutoMigrate(&model.Doctor{})

	// Repository initialization
	repo := repository.NewDoctorRepository(db)

	// Usecase initialization
	uc := usecase.NewDoctorUsecase(repo)

	server, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	// gRPC server setup
	grpcServer := grpc.NewServer()
	grpc_handler.RegisterDoctorHandlers(grpcServer, &uc)
	reflection.Register(grpcServer)

	fmt.Printf("Server running on localhost%s\n", port)

	return grpcServer.Serve(server)
}
