package services

import (
	"database/sql"
	"time"

	"donapresentes/logger"
	"donapresentes/repositories"
)

// AuditService records audit trail entries for critical business operations.
// It logs actions to both the structured logger and the database asynchronously.
type AuditService struct {
	auditRepo *repositories.AuditRepository
}

// NewAuditService creates an AuditService backed by the given database connection.
func NewAuditService(db *sql.DB) *AuditService {
	return &AuditService{
		auditRepo: repositories.NewAuditRepository(db),
	}
}

// Log records an audit entry with full context (user, action, resource, details,
// IP address, and request ID). It writes to the structured logger synchronously
// and persists to the database asynchronously via a goroutine.
func (s *AuditService) Log(userID *int, action, resource, details, ipAddress, requestID string) {
	if s == nil || s.auditRepo == nil {
		return
	}

	logger.Log(logger.Entry{
		Level:      "audit",
		Message:    action + " " + resource,
		RemoteAddr: ipAddress,
		RequestID:  requestID,
		Error:      details,
	})

	go func() {
		if err := s.auditRepo.Insert(userID, action, resource, details, ipAddress, requestID, time.Now()); err != nil {
			logger.Log(logger.Entry{
				Level:   "error",
				Message: "falha ao persistir audit log: " + action + " " + resource,
				Error:   err.Error(),
			})
		}
	}()
}

// LogSimple records an audit entry without a request ID context.
// It delegates to Log with an empty requestID.
func (s *AuditService) LogSimple(userID *int, action, resource, details, ipAddress string) {
	s.Log(userID, action, resource, details, ipAddress, "")
}
