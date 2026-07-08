package logger

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Entry struct {
	Timestamp     string `json:"timestamp"`
	Level         string `json:"level"`
	RequestID     string `json:"request_id,omitempty"`
	Method        string `json:"method,omitempty"`
	Path          string `json:"path,omitempty"`
	Status        int    `json:"status,omitempty"`
	Duration      string `json:"duration,omitempty"`
	RemoteAddr    string `json:"remote_addr,omitempty"`
	UserAgent     string `json:"user_agent,omitempty"`
	UserID        int    `json:"user_id,omitempty"`
	Message       string `json:"message"`
	Slow          bool   `json:"slow,omitempty"`
	Error         string `json:"error,omitempty"`
	Stack         string `json:"stack,omitempty"`
	DBQuery       string `json:"db_query,omitempty"`
	DBDuration    string `json:"db_duration,omitempty"`
}

var jsonLogger *log.Logger

func init() {
	flags := 0
	if os.Getenv("LOG_FORMAT") == "json" {
		flags = 0
	} else {
		flags = log.LstdFlags
	}
	jsonLogger = log.New(os.Stdout, "", flags)
}

func Log(entry Entry) {
	if os.Getenv("LOG_FORMAT") == "json" {
		entry.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
		data, _ := json.Marshal(entry)
		jsonLogger.Println(string(data))
	} else {
		msg := entry.Message
		if entry.RequestID != "" {
			msg = "[" + entry.RequestID + "] " + msg
		}
		if entry.Slow {
			msg = "[SLOW] " + msg
		}
		if entry.Level == "error" || entry.Level == "warn" {
			msg = "[" + entry.Level + "] " + msg
		}
		jsonLogger.Println(msg)
	}
}
