package app

import (
	"fmt"
	"os"

	"github.com/CoffeeSi/golang2AITU/assignment1/appointment-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment1/appointment-service/internal/repository"
	"github.com/CoffeeSi/golang2AITU/assignment1/appointment-service/internal/transport/http"
	"github.com/CoffeeSi/golang2AITU/assignment1/appointment-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Run(port string) error {
	godotenv.Load()

	// Retrieve doctor service URL
	doctorServiceURL := os.Getenv("DOCTOR_SERVICE_URL")
	if doctorServiceURL == "" {
		doctorServiceURL = "http://localhost:8080"
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

	// Repository initialization
	repo := repository.NewAppointmentRepository(db)

	// Usecase initialization
	uc := usecase.NewAppointmentUsecase(repo, doctorServiceURL)

	// Gin router setup
	r := gin.Default()

	// Handler registration
	http.RegisterAppointmentHandlers(r, uc)

	return r.Run(port)
}
