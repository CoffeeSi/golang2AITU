package model

type Status string

const (
	StatusScheduled Status = "new"
	StatusCompleted Status = "in_progress"
	StatusCancelled Status = "done"
)

var Statuses = []Status{
	StatusScheduled,
	StatusCompleted,
	StatusCancelled,
}