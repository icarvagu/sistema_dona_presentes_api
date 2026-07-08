package services

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	apperrors "donapresentes/errors"
	"donapresentes/models"
	"donapresentes/repositories"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var passwordComplexity = regexp.MustCompile(`^.{8,}$`)
var passwordUpper = regexp.MustCompile(`[A-Z]`)
var passwordLower = regexp.MustCompile(`[a-z]`)
var passwordDigit = regexp.MustCompile(`[0-9]`)
var passwordSpecial = regexp.MustCompile(`[^a-zA-Z0-9]`)

type AuthService struct {
	userRepo      *repositories.UserRepository
	refreshRepo   *repositories.RefreshTokenRepository
	auditService  *AuditService
	jwtSecret     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewAuthService(userRepo *repositories.UserRepository, refreshRepo *repositories.RefreshTokenRepository, auditService *AuditService) *AuthService {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	return &AuthService{
		userRepo:      userRepo,
		refreshRepo:   refreshRepo,
		auditService:  auditService,
		jwtSecret:     []byte(secret),
		accessExpiry:  15 * time.Minute,
		refreshExpiry: 7 * 24 * time.Hour,
	}
}

func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *AuthService) GetJWTSecret() []byte {
	return s.jwtSecret
}

func (s *AuthService) ValidatePasswordComplexity(password string) error {
	if !passwordComplexity.MatchString(password) {
		return apperrors.NewValidationError("Senha deve ter pelo menos 8 caracteres")
	}
	if !passwordUpper.MatchString(password) {
		return apperrors.NewValidationError("Senha deve conter pelo menos uma letra maiúscula")
	}
	if !passwordLower.MatchString(password) {
		return apperrors.NewValidationError("Senha deve conter pelo menos uma letra minúscula")
	}
	if !passwordDigit.MatchString(password) {
		return apperrors.NewValidationError("Senha deve conter pelo menos um dígito")
	}
	if !passwordSpecial.MatchString(password) {
		return apperrors.NewValidationError("Senha deve conter pelo menos um caractere especial")
	}
	return nil
}

func (s *AuthService) Login(input *models.LoginInput, ipAddress string) (*models.LoginResponse, error) {
	input.Username = SanitizeUsername(input.Username)
	if input.Username == "" {
		return nil, apperrors.NewValidationError("Usuário ou senha inválidos")
	}

	user, err := s.userRepo.GetByUsername(input.Username)
	if err != nil {
		if err == sql.ErrNoRows || apperrors.IsNotFound(err) {
			return nil, apperrors.NewValidationError("Usuário ou senha inválidos")
		}
		return nil, err
	}

	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		s.auditService.LogSimple(&user.ID, "login_blocked", "user", "Account locked due to too many failed attempts", ipAddress)
		return nil, apperrors.NewValidationError("Conta temporariamente bloqueada por muitas tentativas. Tente novamente mais tarde")
	}

	if user.MustChangePassword {
		passwordOK := s.CheckPassword(input.Password, user.PasswordHash)
		if passwordOK {
			return nil, apperrors.NewValidationError("Você precisa alterar sua senha antes de continuar")
		}
	}

	if !s.CheckPassword(input.Password, user.PasswordHash) {
		newAttempts := user.FailedLoginAttempts + 1
		lockUntil := (*time.Time)(nil)
		if newAttempts >= 5 {
			t := time.Now().Add(15 * time.Minute)
			lockUntil = &t
		}
		_ = s.userRepo.UpdateFailedAttempts(user.ID, newAttempts, lockUntil)
		s.auditService.LogSimple(&user.ID, "login_failed", "user", fmt.Sprintf("Failed login attempt %d/5", newAttempts), ipAddress)
		return nil, apperrors.NewValidationError("Usuário ou senha inválidos")
	}

	_ = s.userRepo.ResetFailedAttempts(user.ID)

	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	refreshToken, err := s.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	user.PasswordHash = ""

	s.auditService.LogSimple(&user.ID, "login_success", "user", "Login successful", ipAddress)

	return &models.LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) GenerateAccessToken(user *models.User) (string, error) {
	expirationTime := time.Now().Add(s.accessExpiry)

	permissions := user.Permissions
	if permissions == nil {
		permissions = []string{}
	}

	claims := jwt.MapClaims{
		"user_id":     user.ID,
		"username":    user.Username,
		"role":        user.Role,
		"permissions": permissions,
		"type":        "access",
		"exp":         expirationTime.Unix(),
		"iat":         time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *AuthService) GenerateRefreshToken(userID int) (string, error) {
	tokenBytes := make([]byte, 32)
	for i := range tokenBytes {
		tokenBytes[i] = byte(time.Now().UnixNano() & 0xFF)
	}

	randomData := fmt.Sprintf("%d-%d", userID, time.Now().UnixNano())
	token := fmt.Sprintf("%x", sha256.Sum256([]byte(randomData)))[:64]

	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
	expiresAt := time.Now().Add(s.refreshExpiry)

	err := s.refreshRepo.Create(userID, tokenHash, expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) RefreshAccessToken(refreshToken, ipAddress string) (*models.LoginResponse, error) {
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(refreshToken)))

	rt, err := s.refreshRepo.FindByHash(tokenHash)
	if err != nil {
		return nil, apperrors.NewUnauthorizedError("Refresh token inválido ou expirado")
	}

	if rt.Revoked || time.Now().After(rt.ExpiresAt) {
		return nil, apperrors.NewUnauthorizedError("Refresh token inválido ou expirado")
	}

	_ = s.refreshRepo.Revoke(rt.ID)

	user, err := s.userRepo.GetByID(rt.UserID)
	if err != nil {
		return nil, apperrors.NewUnauthorizedError("Usuário não encontrado")
	}

	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, apperrors.NewValidationError("Conta temporariamente bloqueada")
	}

	accessToken, err := s.GenerateAccessToken(user)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	newRefreshToken, err := s.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, apperrors.NewDatabaseError(err)
	}

	user.PasswordHash = ""

	s.auditService.LogSimple(&user.ID, "token_refreshed", "auth", "Access token refreshed", ipAddress)

	return &models.LoginResponse{
		Token:        accessToken,
		RefreshToken: newRefreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) Logout(userID int, ipAddress string) error {
	err := s.refreshRepo.RevokeAllForUser(userID)
	if err != nil {
		return err
	}
	s.auditService.LogSimple(&userID, "logout", "auth", "User logged out", ipAddress)
	return nil
}

func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	return s.GenerateAccessToken(user)
}

func (s *AuthService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("método de assinatura inválido")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("token inválido")
}

func (s *AuthService) GetUserByID(id int) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

func (s *AuthService) InitiatePasswordReset(username string) (string, error) {
	username = SanitizeUsername(username)
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return "", nil
	}

	token, err := s.GeneratePasswordResetToken(user.ID)
	if err != nil {
		return "", err
	}

	s.auditService.LogSimple(&user.ID, "password_reset_requested", "user", "Password reset requested", "")

	return token, nil
}

func (s *AuthService) GeneratePasswordResetToken(userID int) (string, error) {
	token := fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d-%d", userID, time.Now().UnixNano()))))[:64]
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err := s.userRepo.CreatePasswordResetToken(userID, tokenHash, expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) ResetPassword(token, newPassword string) error {
	newPassword = SanitizePassword(newPassword)
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))

	prt, err := s.userRepo.FindPasswordResetToken(tokenHash)
	if err != nil {
		return apperrors.NewValidationError("Token de reset inválido ou expirado")
	}

	if prt.Used || time.Now().After(prt.ExpiresAt) {
		return apperrors.NewValidationError("Token de reset inválido ou expirado")
	}

	if err := s.ValidatePasswordComplexity(newPassword); err != nil {
		return err
	}

	passwordHash, err := s.HashPassword(newPassword)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}

	err = s.userRepo.UpdatePassword(prt.UserID, passwordHash)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}

	_ = s.userRepo.MarkPasswordResetTokenUsed(prt.ID)

	_ = s.refreshRepo.RevokeAllForUser(prt.UserID)

	s.auditService.LogSimple(&prt.UserID, "password_reset_completed", "user", "Password reset completed", "")

	return nil
}
