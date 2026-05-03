package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment3/doctor-service/internal/event"
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

	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	var publisher event.DoctorEventPublisherInterface = event.NoopPublisher{}
	closePublisher := func() error { return nil }

	var brokerPublisher *event.EventPublisher
	var err error
	backoff := 250 * time.Millisecond
	for attempt := 1; attempt <= 5; attempt++ {
		brokerPublisher, err = event.NewPublisher(amqpURL)
		if err == nil {
			break
		}
		if attempt < 5 {
			log.Printf("[WARNING] broker unavailable, retrying publish setup (attempt %d/5): %v", attempt, err)
			time.Sleep(backoff)
			backoff *= 2
		}
	}
	if err != nil {
		log.Printf("[WARNING] broker unavailable, events disabled: %v", err)
	} else {
		publisher = brokerPublisher
		closePublisher = brokerPublisher.Close
	}
	defer func() {
		if err := closePublisher(); err != nil {
			log.Printf("[WARNING] failed to close event publisher: %v", err)
		}
	}()

	// Database url configuration
	dsn := os.Getenv("DATABASE_URL")

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return fmt.Errorf("failed to parse database config: %w", err)
	}

	// Database connection
	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Verify database is accessible
	if err := db.Ping(context.Background()); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Database migration
	migrator, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	err = migrator.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migration failed: %w", err)
	}

	// Repository initialization
	repo := repository.NewDoctorRepository(db)

	// Usecase initialization
	uc := usecase.NewDoctorUsecase(repo, publisher)

	server, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	// gRPC server setup
	grpcServer := grpc.NewServer()
	grpc_handler.RegisterDoctorHandlers(grpcServer, &uc)
	reflection.Register(grpcServer)

	fmt.Printf("Server running on localhost%s\n", port)

	return grpcServer.Serve(server)
}
