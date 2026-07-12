package services

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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

// AuthService handles authentication and authorization operations including
// login with account lockout protection, JWT token generation and validation,
// password reset flows, and refresh token rotation.
type AuthService struct {
	userRepo      *repositories.UserRepository
	refreshRepo   *repositories.RefreshTokenRepository
	auditService  *AuditService
	jwtSecret     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewAuthService creates an AuthService with the required repositories and audit trail.
// It reads the JWT secret from the JWT_SECRET environment variable and panics if it is empty.
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

// HashPassword generates a bcrypt hash of the password with the default cost.
func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash.
func (s *AuthService) CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GetJWTSecret returns the JWT signing key as raw bytes.
func (s *AuthService) GetJWTSecret() []byte {
	return s.jwtSecret
}

// ValidatePasswordComplexity enforces that the password meets the minimum requirements:
// at least 8 characters, one uppercase letter, one lowercase letter, one digit,
// and one special character.
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

// Login authenticates a user by username and password and returns JWT tokens.
// It tracks failed attempts and locks the account for 15 minutes after 5 consecutive failures.
// If the user must change their password, login is blocked until the password is updated.
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

// GenerateAccessToken creates a signed JWT access token with user claims.
// The token includes user ID, username, role, permissions, and expires in 15 minutes.
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

// GenerateRefreshToken creates a cryptographically random 256-bit refresh token
// and stores its SHA-256 hash in the database. The token expires in 7 days.
func (s *AuthService) GenerateRefreshToken(userID int) (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", apperrors.NewDatabaseError(err)
	}

	token := hex.EncodeToString(randomBytes)

	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
	expiresAt := time.Now().Add(s.refreshExpiry)

	err := s.refreshRepo.Create(userID, tokenHash, expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

// RefreshAccessToken issues a new access token using a valid refresh token.
// The old refresh token is revoked (rotation) and a new one is issued.
// If the user account is locked, the refresh is denied.
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

// Logout revokes all refresh tokens for a user, effectively ending all active sessions.
func (s *AuthService) Logout(userID int, ipAddress string) error {
	err := s.refreshRepo.RevokeAllForUser(userID)
	if err != nil {
		return err
	}
	s.auditService.LogSimple(&userID, "logout", "auth", "User logged out", ipAddress)
	return nil
}

// GenerateToken creates a JWT access token for the given user.
// Deprecated alias for GenerateAccessToken.
func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	return s.GenerateAccessToken(user)
}

// ValidateToken parses and validates a JWT token string, returning the claims
// if the token is valid and signed with the correct HS256 key.
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

// GetUserByID retrieves a user by ID with the password hash stripped.
func (s *AuthService) GetUserByID(id int) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	user.PasswordHash = ""
	return user, nil
}

// InitiatePasswordReset generates a password reset token for the given username.
// Returns the token only if the user exists (to prevent username enumeration, nil error is returned either way).
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

// GeneratePasswordResetToken creates a cryptographically random reset token
// and stores its SHA-256 hash in the database with a 1-hour expiry.
func (s *AuthService) GeneratePasswordResetToken(userID int) (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", apperrors.NewDatabaseError(err)
	}

	token := hex.EncodeToString(randomBytes)
	tokenHash := fmt.Sprintf("%x", sha256.Sum256([]byte(token)))
	expiresAt := time.Now().Add(1 * time.Hour)

	_, err := s.userRepo.CreatePasswordResetToken(userID, tokenHash, expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ResetPassword completes a password reset using a valid reset token.
// Validates password complexity, hashes the new password, marks the token as used,
// and revokes all active refresh tokens.
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

// ChangePassword updates the password for the authenticated user.
// Validates password complexity, hashes the new password, clears the must_change_password flag,
// and revokes all active refresh tokens for security.
func (s *AuthService) ChangePassword(userID int, newPassword string) error {
	newPassword = SanitizePassword(newPassword)

	if err := s.ValidatePasswordComplexity(newPassword); err != nil {
		return err
	}

	passwordHash, err := s.HashPassword(newPassword)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}

	err = s.userRepo.UpdatePassword(userID, passwordHash)
	if err != nil {
		return apperrors.NewDatabaseError(err)
	}

	_ = s.refreshRepo.RevokeAllForUser(userID)

	s.auditService.LogSimple(&userID, "password_changed", "user", "Password changed by user", "")

	return nil
}
