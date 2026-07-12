package repositories

import (
	"time"

	"donapresentes/models"
)

// UserRepositoryInterface defines the contract for database operations on the users table.
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

// RefreshTokenRepositoryInterface defines the contract for database operations on the refresh_tokens table.
type RefreshTokenRepositoryInterface interface {
	Create(userID int, tokenHash string, expiresAt time.Time) error
	FindByHash(hash string) (*RefreshToken, error)
	Revoke(id int) error
	RevokeAllForUser(userID int) error
}

// CustomerRepositoryInterface defines the contract for database operations on the customers table.
type CustomerRepositoryInterface interface {
	GetAll() ([]models.Customer, error)
	GetByID(id int) (*models.Customer, error)
	Create(c *models.Customer) error
	Update(id int, c *models.Customer) error
	Delete(id int) error
}

// SupplierRepositoryInterface defines the contract for database operations on the suppliers table.
type SupplierRepositoryInterface interface {
	GetAll() ([]models.Supplier, error)
	GetByID(id int) (*models.Supplier, error)
	Create(s *models.Supplier) error
	Update(id int, s *models.Supplier) error
	Delete(id int) error
}

// CarrierRepositoryInterface defines the contract for database operations on the carriers table.
type CarrierRepositoryInterface interface {
	GetAll() ([]models.Carrier, error)
	GetByID(id int) (*models.Carrier, error)
	Create(c *models.Carrier) error
	Update(id int, c *models.Carrier) error
	Delete(id int) error
}

// ProductRepositoryInterface defines the contract for database operations on the products table.
type ProductRepositoryInterface interface {
	GetAll() ([]models.Product, error)
	GetByID(id int) (*models.Product, error)
	Create(p *models.Product) (*models.Product, error)
	Update(id int, p *models.Product) (*models.Product, error)
	Delete(id int) error
	GetAllPaginated(filter string, page, limit int) (*models.PaginatedProductResponse, error)
	SearchByFilter(filter string, limit int) ([]models.Product, error)
	GetGroups() ([]string, error)
	GetPendingApproval() ([]models.Product, error)
	ApproveProduct(id int, origin string) error
	BulkApproveAll(origin string) (int64, error)
	UpdateLastCost(id int, cost float64, date *time.Time, qty1 int, qty2 int, qty3 int, val1 float64, val2 float64, val3 float64, user string) error
	GetByGroup(group string) ([]models.Product, error)
	GetFinancialReport() ([]models.FinancialReportItem, error)
	CreateProductItem(item *models.ProductItem) error
	DeleteProductItems(parentID int) error
	GetNewlyImported() ([]int, error)
}

// AuditServiceInterface defines the contract for audit logging operations on the audit_logs table.
type AuditServiceInterface interface {
	LogSimple(userID *int, action, entity, detail, ip string)
}
