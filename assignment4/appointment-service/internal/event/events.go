package event

const (
	AppointmentCreatedEventType       = "appointments.created"
	AppointmentStatusUpdatedEventType = "appointments.status_updated"
)

type AppointmentCreatedEvent struct {
	EventType  string `json:"event_type"`
	OccurredAt string `json:"occurred_at"`
	ID         string `json:"id"`
	Title      string `json:"title"`
	DoctorID   string `json:"doctor_id"`
	Status     string `json:"status"`
}

type AppointmentStatusUpdatedEvent struct {
	EventType  string `json:"event_type"`
	OccurredAt string `json:"occurred_at"`
	ID         string `json:"id"`
	OldStatus  string `json:"old_status"`
	NewStatus  string `json:"new_status"`
}
