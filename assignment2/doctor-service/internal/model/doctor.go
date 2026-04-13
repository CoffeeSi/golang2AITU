package model

type Doctor struct {
	ID             string `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FullName       string `json:"full_name" gorm:"not null"`
	Specialization string `json:"specialization" gorm:"not null"`
	Email          string `json:"email" gorm:"not null;unique"`
}