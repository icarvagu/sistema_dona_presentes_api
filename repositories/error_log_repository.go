package repositories

import (
	"database/sql"
	"time"

	apperrors "donapresentes/errors"
)

// ErrorLogRepository handles all database operations for the error_logs table.
type ErrorLogRepository struct {
	db *sql.DB
}

// ErrorLogRecord represents a row from the error_logs table.
type ErrorLogRecord struct {
	RequestID   string
	UserID      *int
	ErrorCode   int
	ErrorMessage string
	StackTrace  string
	Path        string
	Method      string
	CreatedAt   time.Time
}

// NewErrorLogRepository creates a new ErrorLogRepository with the given database connection.
func NewErrorLogRepository(db *sql.DB) *ErrorLogRepository {
	return &ErrorLogRepository{db: db}
}

// Insert inserts a new error log record into the database.
func (r *ErrorLogRepository) Insert(record *ErrorLogRecord) error {
	_, err := r.db.Exec(
		`INSERT INTO error_logs (request_id, user_id, error_code, error_message, stack_trace, path, method, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		record.RequestID, record.UserID, record.ErrorCode, record.ErrorMessage,
		record.StackTrace, record.Path, record.Method, record.CreatedAt,
	)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}
