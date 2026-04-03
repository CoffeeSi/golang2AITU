package dto

import (
	"time"

	"github.com/CoffeeSi/golang2AITU/assignment1/appointment-service/internal/model"
)

type CreateAppointmentRequest struct {
	Title       string 		 `json:"title" binding:"required"`
	Description string 		 `json:"description" binding:"required"`
	DoctorID    string 		 `json:"doctor_id" binding:"required,uuid"`
}

type AppointmentResponse struct {
	ID          string    		`json:"id"`
	Title       string    		`json:"title"`
	Description string    		`json:"description"`
	DoctorID    string    		`json:"doctor_id"`
	Status      model.Status    `json:"status"`
	CreatedAt   time.Time 		`json:"created_at"`
	UpdatedAt   time.Time 		`json:"updated_at"`
}

type UpdateAppointmentStatusRequest struct {
	Status model.Status `json:"status" binding:"required"`
}

