package model

import "errors"

var (
	AppointmentNotFoundError         = errors.New("appointment not found")
	InvalidArgumentCreateError       = errors.New("title and doctor_id are required")
	InvalidArgumentGetError          = errors.New("id is required")
	InvalidArgumentUpdateStatusError = errors.New("id and status are required")
	InvalidIDFormatError             = errors.New("id must be a valid UUID")
	InvalidStatusError               = errors.New("invalid status")
	DoneStatusTransitionError        = errors.New("transitioning a status from done back to new is not allowed")
	ServiceUnavailableError          = errors.New("doctor service is temporarily unavailable")
	DoctorDoesNotExistError          = errors.New("doctor does not exist")
)
