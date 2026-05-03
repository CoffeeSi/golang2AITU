package model

type Status string

const (
	StatusNew        Status = "new"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

var Statuses = []Status{
	StatusNew,
	StatusInProgress,
	StatusDone,
}
