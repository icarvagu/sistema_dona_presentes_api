package services

import (
	"log"
	"os"
	"sync"
	"time"

	"donapresentes/repositories"
)

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

type RequestLoggerService struct {
	repo     *repositories.RequestLogRepository
	buffer   []RequestLogEntry
	mu       sync.Mutex
	done     chan struct{}
	enabled  bool
}

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

func (s *RequestLoggerService) Shutdown() {
	if s.enabled {
		close(s.done)
	}
}
