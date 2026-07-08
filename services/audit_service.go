package services

import (
	"database/sql"
	"time"

	"donapresentes/repositories"
)

type AuditService struct {
	auditRepo *repositories.AuditRepository
}

func NewAuditService(db *sql.DB) *AuditService {
	return &AuditService{
		auditRepo: repositories.NewAuditRepository(db),
	}
}

func (s *AuditService) Log(userID *int, action, resource, details, ipAddress string) {
	if s == nil || s.auditRepo == nil {
		return
	}
	go func() {
		_ = s.auditRepo.Insert(userID, action, resource, details, ipAddress, time.Now())
	}()
}
