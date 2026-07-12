package controllers

import (
	"encoding/json"
	"net/http"
	"time"

	"donapresentes/controllers/config"
)

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Database  string `json:"database"`
	Version   string `json:"version"`
}

// HealthCheck handles GET /health — returns the health status of the API and its database connection.
func HealthCheck(w http.ResponseWriter, r *http.Request) {
	dbStatus := "ok"
	if config.DB == nil {
		dbStatus = "disconnected"
	} else if err := config.DB.Ping(); err != nil {
		dbStatus = "error: " + err.Error()
	}

	status := "healthy"
	if dbStatus != "ok" {
		status = "degraded"
	}

	resp := HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Database:  dbStatus,
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	if status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(resp)
}
