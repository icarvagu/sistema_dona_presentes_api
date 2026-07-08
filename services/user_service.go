package services

import (
	"database/sql"
	"strings"

	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"
)

type UserService struct {
	userRepo      *repositories.UserRepository
	authService   *AuthService
	cryptoService *CryptoService
}

func NewUserService(userRepo *repositories.UserRepository, authService *AuthService, cryptoService *CryptoService) *UserService {
	return &UserService{
		userRepo:      userRepo,
		authService:   authService,
		cryptoService: cryptoService,
	}
}

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
		Username:            input.Username,
		PasswordHash:        passwordHash,
		Role:                input.Role,
		Permissions:         input.Permissions,
		FullName:            input.FullName,
		CPF:                 encryptedCPF,
		RG:                  input.RG,
		BirthDate:           input.BirthDate,
		Gender:              input.Gender,
		Status:              input.Status,
		ContactEmail:        input.ContactEmail,
		FullAddress:         input.FullAddress,
		ContactPhone:        input.ContactPhone,
		Notes:               input.Notes,
		MustChangePassword:  input.MustChangePassword,
	}

	if user.Status == "" {
		user.Status = "Ativo"
	}

	created, err := s.userRepo.Create(user)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	if cpfHash != "" {
		_ = s.userRepo.UpdateCPFHash(created.ID, cpfHash)
	}

	created.PasswordHash = ""
	created.CPF = input.CPF
	return created, nil
}

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
		_ = s.userRepo.UpdateCPFHash(updated.ID, cpfHash)
	}

	updated.PasswordHash = ""
	updated.CPF = input.CPF
	return updated, nil
}

func (s *UserService) Delete(id int) error {

	_, err := s.userRepo.GetByID(id)
	if err != nil {
		if err == sql.ErrNoRows || apperrors.IsNotFound(err) {
			return apperrors.NewNotFoundError("Usuário não encontrado")
		}
		return apperrors.NewDatabaseError(err)
	}

	return s.userRepo.Delete(id)
}

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
