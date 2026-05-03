package app

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/CoffeeSi/golang2AITU/assignment3/appointment-service/internal/client"
	"github.com/CoffeeSi/golang2AITU/assignment3/appointment-service/internal/repository"
	grpc_handler "github.com/CoffeeSi/golang2AITU/assignment3/appointment-service/internal/transport/grpc"
	"github.com/CoffeeSi/golang2AITU/assignment3/appointment-service/internal/usecase"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func Run() error {
	godotenv.Load()

	// Port configuration
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	if port == ":" {
		port = ":8081"
	}

	// Retrieve doctor service URL
	doctorServiceURL := os.Getenv("DOCTOR_SERVICE_URL")
	if doctorServiceURL == "" {
		doctorServiceURL = "localhost:8080"
	}

	// Database url configuration
	dsn := os.Getenv("DB_URL")

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return err
	}

	// Database connection
	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return err
	}
	defer db.Close()

	// Database migration
	migrator, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		return err
	}

	err = migrator.Up()
	if err != nil {
		fmt.Printf("Migration: %v\n", err)
	}

	// gRPC client
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
