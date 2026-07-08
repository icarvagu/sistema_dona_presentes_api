package repositories

import (
	"database/sql"
	"time"

	apperrors "donapresentes/errors"
)

type RefreshTokenRepository struct {
	db *sql.DB
}

type RefreshToken struct {
	ID        int
	UserID    int
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	Revoked   bool
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(userID int, tokenHash string, expiresAt time.Time) error {
	_, err := r.db.Exec(
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

func (r *RefreshTokenRepository) FindByHash(tokenHash string) (*RefreshToken, error) {
	var rt RefreshToken
	err := r.db.QueryRow(
		`SELECT id, user_id, token_hash, expires_at, created_at, revoked FROM refresh_tokens WHERE token_hash=$1`,
		tokenHash,
	).Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt, &rt.Revoked)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.NewNotFoundError("Refresh token não encontrado")
		}
		return nil, apperrors.NewDatabaseError(err)
	}
	return &rt, nil
}

func (r *RefreshTokenRepository) Revoke(id int) error {
	_, err := r.db.Exec(`UPDATE refresh_tokens SET revoked=TRUE WHERE id=$1`, id)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}

func (r *RefreshTokenRepository) RevokeAllForUser(userID int) error {
	_, err := r.db.Exec(`UPDATE refresh_tokens SET revoked=TRUE WHERE user_id=$1 AND revoked=FALSE`, userID)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}
	return nil
}
