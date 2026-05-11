package main

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

type NotifyRequest struct {
	IdempotencyKey string `json:"idempotency_key"`
	Channel        string `json:"channel"`
	Recipient      string `json:"recipient"`
	Message        string `json:"message"`
}

var (
	processedKeys = make(map[string]bool)
	mu            sync.Mutex
)

func main() {
	godotenv.Load()
	port := fmt.Sprintf(":%s", os.Getenv("GATEWAY_PORT"))
	if port == ":" {
		port = ":8080"
	}

	http.HandleFunc("/notify", notifyHandler)
	http.ListenAndServe(port, nil)
}

func notifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var request NotifyRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	logRequest(request)

	w.Header().Set("Content-Type", "application/json")

	if rand.Float32() < 0.2 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
		})
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if processedKeys[request.IdempotencyKey] {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "duplicate",
		})
		return
	}
	processedKeys[request.IdempotencyKey] = true
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "accepted",
	})
}

func logRequest(req NotifyRequest) {
	logLine := map[string]interface{}{
		"time":            time.Now().Format(time.RFC3339),
		"idempotency_key": req.IdempotencyKey,
		"recipient":       req.Recipient,
		"message":         req.Message,
	}
	data, _ := json.Marshal(logLine)
	fmt.Println(string(data))
}
