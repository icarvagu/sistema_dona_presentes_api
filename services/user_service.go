package services

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

// UserService manages user accounts including CRUD, password hashing,
// CPF encryption/hashing, duplicate detection, and input validation.
type UserService struct {
	userRepo      *repositories.UserRepository
	authService   *AuthService
	cryptoService *CryptoService
	auditService  *AuditService
}

// NewUserService creates a UserService with the required repositories,
// auth service, crypto service, and audit trail.
func NewUserService(userRepo *repositories.UserRepository, authService *AuthService, cryptoService *CryptoService, auditService *AuditService) *UserService {
	return &UserService{
		userRepo:      userRepo,
		authService:   authService,
		cryptoService: cryptoService,
		auditService:  auditService,
	}
}

// GetAll returns all users with password hashes stripped and CPF decrypted.
func (s *UserService) GetAll() ([]models.User, error) {
	users, err := s.userRepo.GetAll()
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i].PasswordHash = ""
		if s.cryptoService != nil && users[i].CPF != "" {
			decrypted, err := s.cryptoService.Decrypt(users[i].CPF)
			if err == nil {
				users[i].CPF = decrypted
			}
		}
	}
	return users, nil
}

// GetByID returns a single user by ID with password hash stripped and CPF decrypted.
func (s *UserService) GetByID(id int) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows || apperrors.IsNotFound(err) {
			return nil, apperrors.NewNotFoundError("Usuário não encontrado")
		}
		return nil, apperrors.NewDatabaseError(err)
	}

	user.PasswordHash = ""

	if s.cryptoService != nil && user.CPF != "" {
		decrypted, err := s.cryptoService.Decrypt(user.CPF)
		if err == nil {
			user.CPF = decrypted
		}
	}

	return user, nil
}

// Create registers a new user with sanitized input, duplicate checks
// (username and CPF), password hashing, and CPF encryption.
func (s *UserService) Create(input *models.UserInput) (*models.User, error) {
	input.Username = SanitizeUsername(input.Username)
	input.FullName = SanitizeString(input.FullName)
	input.CPF = SanitizeCPF(input.CPF)
	input.Password = SanitizePassword(input.Password)

	if err := s.validateUserInput(input, true); err != nil {
		return nil, err
	}

	_, err := s.userRepo.GetByUsername(input.Username)
	if err == nil {
		return nil, apperrors.NewValidationError("Username já existe")
	}

	if input.CPF != "" {
		cpfHash := s.cryptoService.Hash(input.CPF)
		_, err := s.userRepo.GetByCPFHash(cpfHash)
		if err == nil {
			return nil, apperrors.NewValidationError("CPF já cadastrado")
		}
	}

	passwordHash, err := s.authService.HashPassword(input.Password)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	encryptedCPF := input.CPF
	if s.cryptoService != nil && input.CPF != "" {
		encryptedCPF, err = s.cryptoService.Encrypt(input.CPF)
		if err != nil {
			return nil, apperrors.NewDatabaseError(err)
		}
	}

	cpfHash := ""
	if input.CPF != "" {
		cpfHash = s.cryptoService.Hash(input.CPF)
	}

	user := &models.User{
		Username:           input.Username,
		PasswordHash:       passwordHash,
		Role:               input.Role,
		Permissions:        input.Permissions,
		FullName:           input.FullName,
		CPF:                encryptedCPF,
		RG:                 input.RG,
		BirthDate:          input.BirthDate,
		Gender:             input.Gender,
		Status:             input.Status,
		ContactEmail:       input.ContactEmail,
		FullAddress:        input.FullAddress,
		ContactPhone:       input.ContactPhone,
		Notes:              input.Notes,
		MustChangePassword: input.MustChangePassword,
	}

	if user.Status == "" {
		user.Status = "Ativo"
	}

	created, err := s.userRepo.Create(user)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	if cpfHash != "" {
		if err := s.userRepo.UpdateCPFHash(created.ID, cpfHash); err != nil {
			log.Printf("failed to update CPF hash for user %d: %v", created.ID, err)
		}
	}

	created.PasswordHash = ""
	created.CPF = input.CPF
	if created != nil {
		s.auditService.LogSimple(&created.ID, "user_created", "user", fmt.Sprintf("username=%s role=%s", created.Username, created.Role), "")
	}
	return created, nil
}

// Update modifies an existing user by ID. Validates input, checks for
// duplicate username/CPF, re-encrypts CPF if changed, and optionally
// updates the password.
func (s *UserService) Update(id int, input *models.UserInput) (*models.User, error) {
	input.Username = SanitizeUsername(input.Username)
	input.FullName = SanitizeString(input.FullName)
	input.CPF = SanitizeCPF(input.CPF)
	input.Password = SanitizePassword(input.Password)

	existingUser, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows || apperrors.IsNotFound(err) {
			return nil, apperrors.NewNotFoundError("Usuário não encontrado")
		}
		return nil, apperrors.NewDatabaseError(err)
	}

	if err := s.validateUserInput(input, input.Password == ""); err != nil {
		return nil, err
	}

	if input.Username != existingUser.Username {
		_, err := s.userRepo.GetByUsername(input.Username)
		if err == nil {
			return nil, apperrors.NewValidationError("Username já existe")
		}
	}

	encryptedCPF := input.CPF
	cpfHash := ""
	if input.CPF != "" && input.CPF != existingUser.CPF {
		cpfHash = s.cryptoService.Hash(input.CPF)
		_, err := s.userRepo.GetByCPFHash(cpfHash)
		if err == nil {
			return nil, apperrors.NewValidationError("CPF já cadastrado")
		}
		if s.cryptoService != nil {
			encryptedCPF, err = s.cryptoService.Encrypt(input.CPF)
			if err != nil {
				return nil, apperrors.NewDatabaseError(err)
			}
		}
	} else if input.CPF == existingUser.CPF {
		encryptedCPF = existingUser.CPF
	}

	user := &models.User{
		Username:     input.Username,
		Role:         input.Role,
		Permissions:  input.Permissions,
		FullName:     input.FullName,
		CPF:          encryptedCPF,
		RG:           input.RG,
		BirthDate:    input.BirthDate,
		Gender:       input.Gender,
		Status:       input.Status,
		ContactEmail: input.ContactEmail,
		FullAddress:  input.FullAddress,
		ContactPhone: input.ContactPhone,
		Notes:        input.Notes,
	}

	if user.Status == "" {
		user.Status = existingUser.Status
	}

	if input.Password != "" {
		passwordHash, err := s.authService.HashPassword(input.Password)
		if err != nil {
			return nil, apperrors.NewDatabaseError(err)
		}
		err = s.userRepo.UpdatePassword(id, passwordHash)
		if err != nil {
			return nil, apperrors.NewDatabaseError(err)
		}
	}

	updated, err := s.userRepo.Update(id, user)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	if cpfHash != "" {
		if err := s.userRepo.UpdateCPFHash(updated.ID, cpfHash); err != nil {
			log.Printf("failed to update CPF hash for user %d: %v", updated.ID, err)
		}
	}

	updated.PasswordHash = ""
	updated.CPF = input.CPF
	if updated != nil {
		s.auditService.LogSimple(&updated.ID, "user_updated", "user", fmt.Sprintf("id=%d username=%s", id, updated.Username), "")
	}
	return updated, nil
}

// Delete removes a user by ID after verifying it exists.
func (s *UserService) Delete(id int) error {

	_, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows || apperrors.IsNotFound(err) {
			return apperrors.NewNotFoundError("Usuário não encontrado")
		}
		return apperrors.NewDatabaseError(err)
	}

	err = s.userRepo.Delete(id)
	if err == nil {
		s.auditService.LogSimple(nil, "user_deleted", "user", fmt.Sprintf("id=%d", id), "")
	}
	return err
}

// SearchByFilter searches users by a text filter, returning matches with password hashes stripped.
func (s *UserService) SearchByFilter(filter string) ([]models.User, error) {
	users, err := s.userRepo.SearchByFilter(filter)
	if err != nil {
		return nil, err
	}
	for i := range users {
		users[i].PasswordHash = ""
	}
	return users, nil
}

// validateUserInput enforces user field rules:
// username must be at least 3 characters; password complexity (when required);
// role must be admin/gerente/standard; status must be Ativo/Inativo;
// gender must be Masculino/Feminino/Outro; CPF must have at least 11 digits.
func (s *UserService) validateUserInput(input *models.UserInput, requirePassword bool) error {

	if strings.TrimSpace(input.Username) == "" {
		return apperrors.NewValidationError("Username é obrigatório")
	}
	if len(input.Username) < 3 {
		return apperrors.NewValidationError("Username deve ter pelo menos 3 caracteres")
	}

	if requirePassword {
		if strings.TrimSpace(input.Password) == "" {
			return apperrors.NewValidationError("Senha é obrigatória")
		}
		if err := s.authService.ValidatePasswordComplexity(input.Password); err != nil {
			return err
		}
	}

	if input.Role != "admin" && input.Role != "standard" && input.Role != "gerente" {
		return apperrors.NewValidationError("Role deve ser 'admin', 'gerente' ou 'standard'")
	}

	if input.Status != "" && input.Status != "Ativo" && input.Status != "Inativo" {
		return apperrors.NewValidationError("Status deve ser 'Ativo' ou 'Inativo'")
	}

	if input.Gender != nil && *input.Gender != "Masculino" && *input.Gender != "Feminino" && *input.Gender != "Outro" {
		return apperrors.NewValidationError("Gender deve ser 'Masculino', 'Feminino' ou 'Outro'")
	}

	if input.CPF != "" && len(input.CPF) < 11 {
		return apperrors.NewValidationError("CPF inválido")
	}

	return nil
}
