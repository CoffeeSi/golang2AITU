package model

import "time"

type Appointment struct {
	ID          string
	Title       string
	Description string
	DoctorID    string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
