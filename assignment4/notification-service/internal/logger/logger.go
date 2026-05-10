package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type EventLog struct {
	Time      string `json:"time"`
	Level     string `json:"level"`
	Service   string `json:"service"`
	EventType string `json:"event_type"`
	Payload   any    `json:"payload"`
}

type EventLogger struct {
	serviceName string
}

func NewEventLogger(serviceName string) *EventLogger {
	return &EventLogger{
		serviceName: serviceName,
	}
}

func (l *EventLogger) LogEvent(eventType string, payload any) {
	entry := EventLog{
		Time:      time.Now().Format(time.RFC3339),
		Level:     "info",
		Service:   l.serviceName,
		EventType: eventType,
		Payload:   payload,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, `{"time":"%s","level":"error","msg":"failed to marshal event log","error":"%s"}`+"\n",
			time.Now().Format(time.RFC3339), err.Error())
		return
	}

	fmt.Println(string(data))
}
