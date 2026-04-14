package model

import "errors"

var (
	InvalidCreateArgumentError = errors.New("full name and email are required")
	InvalidGetArgumentError    = errors.New("id is required")
	DoctorAlreadyExistsError   = errors.New("doctor with this email already exists")
	DoctorNotFoundError        = errors.New("doctor not found")
)
