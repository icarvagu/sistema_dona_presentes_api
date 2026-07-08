package repositories

import (
	"database/sql"
	"time"

	apperrors "donapresentes/errors"
)

type RequestLogRepository struct {
	db *sql.DB
}

type RequestLogRecord struct {
	RequestID  string
	Method     string
	Path       string
	Status     int
	DurationMs int
	RemoteAddr string
	UserAgent  string
	UserID     *int
	Slow       bool
	CreatedAt  time.Time
}

func NewRequestLogRepository(db *sql.DB) *RequestLogRepository {
	return &RequestLogRepository{db: db}
}

func (r *RequestLogRepository) BatchInsert(records []RequestLogRecord) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(
		`INSERT INTO request_logs (request_id, method, path, status, duration_ms, remote_addr, user_agent, user_id, slow)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
	)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	defer stmt.Close()

	for _, rec := range records {
		_, err := stmt.Exec(rec.RequestID, rec.Method, rec.Path, rec.Status, rec.DurationMs, rec.RemoteAddr, rec.UserAgent, rec.UserID, rec.Slow)
		if err != nil {
			return apperrors.NewDatabaseError(err)
		}
	}

	return tx.Commit()
}
