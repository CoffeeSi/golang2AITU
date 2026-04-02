package app

import (
	"fmt"
	"os"

	"github.com/CoffeeSi/golang2AITU/assignment1/doctor-service/internal/model"
	"github.com/CoffeeSi/golang2AITU/assignment1/doctor-service/internal/repository"
	"github.com/CoffeeSi/golang2AITU/assignment1/doctor-service/internal/transport/http"
	"github.com/CoffeeSi/golang2AITU/assignment1/doctor-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Run(port string) error {
	godotenv.Load()
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

	// Gin router setup
	r := gin.Default()

	// Handler registration
	http.RegisterDoctorHandlers(r, uc)

	return r.Run(port)
}
