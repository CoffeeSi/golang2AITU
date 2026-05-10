package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment4/doctor-service/internal/cache"
	"github.com/CoffeeSi/golang2AITU/assignment4/doctor-service/internal/event"
	"github.com/CoffeeSi/golang2AITU/assignment4/doctor-service/internal/middleware"
	"github.com/CoffeeSi/golang2AITU/assignment4/doctor-service/internal/repository"
	grpc_handler "github.com/CoffeeSi/golang2AITU/assignment4/doctor-service/internal/transport/grpc"
	"github.com/CoffeeSi/golang2AITU/assignment4/doctor-service/internal/usecase"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func Run() error {
	godotenv.Load()

	db, dsn, err := initDB()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	if err := runMigrations(dsn); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	publisher, closePublisher := initBrokerPublisher()
	defer func() {
		if err := closePublisher(); err != nil {
			log.Printf("[WARNING] failed to close event publisher: %v", err)
		}
	}()

	cacheClient, err := initCache()
	if err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}

	repo := repository.NewDoctorRepository(db)
	uc := usecase.NewDoctorUsecase(repo, publisher, cacheClient)

	return startGRPCServer(&uc, cacheClient)
}

func initDB() (*pgxpool.Pool, string, error) {
	dsn := os.Getenv("DATABASE_URL")

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse database config: %w", err)
	}

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, "", fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		return nil, "", fmt.Errorf("failed to ping database: %w", err)
	}

	return db, dsn, nil
}

func runMigrations(dsn string) error {
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
	return nil
}

func initBrokerPublisher() (event.DoctorEventPublisherInterface, func() error) {
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

	return publisher, closePublisher
}

func initCache() (*cache.CacheClient, error) {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "localhost:6379"
	}

	cacheTTLSeconds := os.Getenv("CACHE_TTL_SECONDS")
	if cacheTTLSeconds == "" {
		cacheTTLSeconds = "3600"
	}

	cacheTTL, err := time.ParseDuration(cacheTTLSeconds + "s")
	if err != nil {
		return nil, fmt.Errorf("failed to parse CACHE_TTL_SECONDS: %w", err)
	}
	cacheClient := cache.NewCacheClient(redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "",
		DB:       0,
	}), cacheTTL)
	return cacheClient, nil
}

func startGRPCServer(uc *usecase.DoctorUsecase, cacheClient cache.CacheClientInterface) error {
	port := fmt.Sprintf(":%s", os.Getenv("GRPC_PORT"))
	if port == ":" {
		port = ":50051"
	}

	limitRPM := 100
	if val, err := strconv.Atoi(os.Getenv("RATE_LIMIT_RPM")); err == nil {
		limitRPM = val
	}

	server, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.RateLimitInterceptor(cacheClient, limitRPM)),
	)
	grpc_handler.RegisterDoctorHandlers(grpcServer, uc)
	reflection.Register(grpcServer)

	fmt.Printf("Server running on localhost%s\n", port)

	return grpcServer.Serve(server)
}
