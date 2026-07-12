package services

import (
	"log"
	"os"
	"sync"
	"time"

	"donapresentes/repositories"
)

// RequestLogEntry represents a single HTTP request log record with timing,
// status, and caller metadata.
type RequestLogEntry struct {
	RequestID  string
	Method     string
	Path       string
	Status     int
	DurationMs int
	RemoteAddr string
	UserAgent  string
	UserID     *int
	Slow       bool
}

// RequestLoggerService buffers HTTP request logs in memory and periodically
// flushes them to the database in batches for efficient persistence.
// Enabled only when the LOG_DB environment variable is set to "true".
type RequestLoggerService struct {
	repo    *repositories.RequestLogRepository
	buffer  []RequestLogEntry
	mu      sync.Mutex
	done    chan struct{}
	enabled bool
}

// NewRequestLoggerService creates a RequestLoggerService backed by the given repository.
// If LOG_DB env is "true", starts a background goroutine that flushes logs every 5 seconds.
func NewRequestLoggerService(repo *repositories.RequestLogRepository) *RequestLoggerService {
	enabled := os.Getenv("LOG_DB") == "true"
	s := &RequestLoggerService{
		repo:    repo,
		buffer:  make([]RequestLogEntry, 0, 100),
		done:    make(chan struct{}),
		enabled: enabled,
	}
	if enabled {
		go s.flusher()
	}
	return s
}

// Log enqueues a request log entry into the buffer. If the buffer reaches 100 entries,
// it flushes immediately. Silently discards entries when logging is disabled.
func (s *RequestLoggerService) Log(entry RequestLogEntry) {
	if !s.enabled {
		return
	}

	s.mu.Lock()
	s.buffer = append(s.buffer, entry)
	if len(s.buffer) >= 100 {
		batch := s.buffer
		s.buffer = make([]RequestLogEntry, 0, 100)
		s.mu.Unlock()
		s.flush(batch)
		return
	}
	s.mu.Unlock()
}

// flusher runs a ticker every 5 seconds to flush accumulated logs.
func (s *RequestLoggerService) flusher() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.mu.Lock()
			if len(s.buffer) == 0 {
				s.mu.Unlock()
				continue
			}
			batch := s.buffer
			s.buffer = make([]RequestLogEntry, 0, 100)
			s.mu.Unlock()
			s.flush(batch)
		case <-s.done:
			s.mu.Lock()
			if len(s.buffer) > 0 {
				s.flush(s.buffer)
				s.buffer = nil
			}
			s.mu.Unlock()
			return
		}
	}
}

// flush converts log entries to repository records and batch-inserts them.
func (s *RequestLoggerService) flush(entries []RequestLogEntry) {
	records := make([]repositories.RequestLogRecord, len(entries))
	for i, e := range entries {
		records[i] = repositories.RequestLogRecord{
			RequestID:  e.RequestID,
			Method:     e.Method,
			Path:       e.Path,
			Status:     e.Status,
			DurationMs: e.DurationMs,
			RemoteAddr: e.RemoteAddr,
			UserAgent:  e.UserAgent,
			UserID:     e.UserID,
			Slow:       e.Slow,
		}
	}

	if err := s.repo.BatchInsert(records); err != nil {
		log.Printf("[LOG_DB] erro ao salvar request logs: %v", err)
	}
}

// Shutdown flushes any remaining buffered logs and stops the background flusher.
func (s *RequestLoggerService) Shutdown() {
	if s.enabled {
		close(s.done)
	}
}
