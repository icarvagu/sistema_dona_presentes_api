package repositories

import (
	"time"

	"donapresentes/models"
)

type UserRepositoryInterface interface {
	GetByUsername(username string) (*models.User, error)
	GetByID(id int) (*models.User, error)
	Create(u *models.User) (*models.User, error)
	GetAll() ([]models.User, error)
	Update(id int, u *models.User) (*models.User, error)
	UpdatePassword(id int, passwordHash string) error
	Delete(id int) error
	GetByCPF(cpf string) (*models.User, error)
	SearchByFilter(filter string) ([]models.User, error)
	UpdateFailedAttempts(id int, attempts int, lockedUntil *time.Time) error
	ResetFailedAttempts(id int) error
	UpdateCPFHash(id int, cpfHash string) error
	GetByCPFHash(cpfHash string) (*models.User, error)
	CreatePasswordResetToken(userID int, tokenHash string, expiresAt time.Time) (*PasswordResetToken, error)
	FindPasswordResetToken(tokenHash string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(id int) error
}

type RefreshTokenRepositoryInterface interface {
	Create(userID int, tokenHash string, expiresAt time.Time) error
	FindByHash(hash string) (*RefreshToken, error)
	Revoke(id int) error
	RevokeAllForUser(userID int) error
}

type CustomerRepositoryInterface interface {
	GetAll() ([]models.Customer, error)
	GetByID(id int) (*models.Customer, error)
	Create(c *models.Customer) error
	Update(id int, c *models.Customer) error
	Delete(id int) error
}

type SupplierRepositoryInterface interface {
	GetAll() ([]models.Supplier, error)
	GetByID(id int) (*models.Supplier, error)
	Create(s *models.Supplier) error
	Update(id int, s *models.Supplier) error
	Delete(id int) error
}

type CarrierRepositoryInterface interface {
	GetAll() ([]models.Carrier, error)
	GetByID(id int) (*models.Carrier, error)
	Create(c *models.Carrier) error
	Update(id int, c *models.Carrier) error
	Delete(id int) error
}

type ProductRepositoryInterface interface {
	GetAll() ([]models.Product, error)
	GetByID(id int) (*models.Product, error)
	Create(p *models.Product) (*models.Product, error)
	Update(id int, p *models.Product) (*models.Product, error)
	Delete(id int) error
}

type AuditServiceInterface interface {
	LogSimple(userID *int, action, entity, detail, ip string)
}
