package app

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/CoffeeSi/golang2AITU/assignment3/doctor-service/internal/repository"
	grpc_handler "github.com/CoffeeSi/golang2AITU/assignment3/doctor-service/internal/transport/grpc"
	"github.com/CoffeeSi/golang2AITU/assignment3/doctor-service/internal/usecase"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Run() error {
	godotenv.Load()

	// Port configuration
	port := fmt.Sprintf(":%s", os.Getenv("PORT"))
	if port == ":" {
		port = ":8080"
	}

	// Database url configuration
	dsn := os.Getenv("DB_URL")

	config, err := pgxpool.ParseConfig(dsn)

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
