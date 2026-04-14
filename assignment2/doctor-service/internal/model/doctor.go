package model

type Doctor struct {
	ID             string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FullName       string `gorm:"not null"`
	Specialization string `gorm:"not null"`
	Email          string `gorm:"not null;unique"`
}
