package services

import (
	"time"

	"donapresentes/models"
	"donapresentes/repositories"
)

type UserRepositoryStub struct{}

func (s *UserRepositoryStub) GetByUsername(username string) (*models.User, error) {
	return &models.User{ID: 1, Username: username, Role: "admin"}, nil
}
func (s *UserRepositoryStub) GetByID(id int) (*models.User, error) {
	return &models.User{ID: id, Username: "testuser", Role: "admin"}, nil
}
func (s *UserRepositoryStub) Create(u *models.User) (*models.User, error) { return u, nil }
func (s *UserRepositoryStub) GetAll() ([]models.User, error)             { return nil, nil }
func (s *UserRepositoryStub) Update(id int, u *models.User) (*models.User, error) { return u, nil }
func (s *UserRepositoryStub) UpdatePassword(id int, passwordHash string) error { return nil }
func (s *UserRepositoryStub) Delete(id int) error                               { return nil }
func (s *UserRepositoryStub) GetByCPF(cpf string) (*models.User, error)         { return nil, nil }
func (s *UserRepositoryStub) SearchByFilter(filter string) ([]models.User, error) { return nil, nil }
func (s *UserRepositoryStub) UpdateFailedAttempts(id int, attempts int, lockedUntil *time.Time) error {
	return nil
}
func (s *UserRepositoryStub) ResetFailedAttempts(id int) error { return nil }
func (s *UserRepositoryStub) UpdateCPFHash(id int, cpfHash string) error { return nil }
func (s *UserRepositoryStub) GetByCPFHash(cpfHash string) (*models.User, error) { return nil, nil }
func (s *UserRepositoryStub) CreatePasswordResetToken(userID int, tokenHash string, expiresAt time.Time) (*repositories.PasswordResetToken, error) {
	return &repositories.PasswordResetToken{ID: 1, UserID: userID, TokenHash: tokenHash, ExpiresAt: expiresAt}, nil
}
func (s *UserRepositoryStub) FindPasswordResetToken(tokenHash string) (*repositories.PasswordResetToken, error) {
	return nil, nil
}
func (s *UserRepositoryStub) MarkPasswordResetTokenUsed(id int) error { return nil }

type RefreshTokenRepositoryStub struct{}

func (s *RefreshTokenRepositoryStub) Create(userID int, tokenHash string, expiresAt time.Time) error {
	return nil
}
func (s *RefreshTokenRepositoryStub) FindByHash(tokenHash string) (*repositories.RefreshToken, error) {
	return &repositories.RefreshToken{ID: 1, UserID: 1, TokenHash: tokenHash, ExpiresAt: time.Now().Add(7 * 24 * time.Hour)}, nil
}
func (s *RefreshTokenRepositoryStub) Revoke(id int) error             { return nil }
func (s *RefreshTokenRepositoryStub) RevokeAllForUser(userID int) error { return nil }
