package model

import "time"

type Appointment struct {
	ID      	string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Title       string `json:"title" gorm:"not null"`
	Description string `json:"description" gorm:"type:text"`
	DoctorID    string `json:"doctor_id" gorm:"type:uuid;not null"`
	Status      Status `json:"status" gorm:"type:varchar(20);not null"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
