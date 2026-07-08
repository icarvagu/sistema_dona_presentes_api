package repositories

import (
	"database/sql"
	"time"

	apperrors "donapresentes/errors"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

type AuditLog struct {
	ID        int       `json:"id"`
	UserID    *int      `json:"user_id,omitempty"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource,omitempty"`
	Details   string    `json:"details,omitempty"`
	IPAddress string    `json:"ip_address,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *AuditRepository) Insert(userID *int, action, resource, details, ipAddress string, createdAt time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO audit_logs (user_id, action, resource, details, ip_address, created_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		userID, action, resource, details, ipAddress, createdAt,
	)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}
