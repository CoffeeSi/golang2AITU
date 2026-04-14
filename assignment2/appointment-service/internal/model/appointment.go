package model

import "time"

type Appointment struct {
	ID          string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Title       string    `gorm:"not null"`
	Description string    `gorm:"type:text"`
	DoctorID    string    `gorm:"type:uuid;not null"`
	Status      Status    `gorm:"type:varchar(20);default:'new';not null"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   time.Time `gorm:"not null"`
}
