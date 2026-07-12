package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

// CryptoService provides AES-GCM encryption/decryption for sensitive data
// and HMAC-SHA256 hashing for unique lookups. The encryption key is derived
// from the ENCRYPTION_KEY environment variable using SHA-256.
type CryptoService struct {
	key  []byte
	salt []byte
}

// NewCryptoService creates a CryptoService if the ENCRYPTION_KEY environment
// variable is set. Returns nil if the key is empty (crypto is disabled).
func NewCryptoService() *CryptoService {
	keyHex := os.Getenv("ENCRYPTION_KEY")
	if keyHex == "" {
		return nil
	}
	key := sha256.Sum256([]byte(keyHex))
	salt := sha256.Sum256([]byte(keyHex + ":salt"))
	return &CryptoService{key: key[:], salt: salt[:]}
}

// Encrypt encrypts plaintext using AES-GCM with a random nonce.
// Returns the base64-encoded ciphertext. If the service is nil, returns plaintext unchanged.
func (s *CryptoService) Encrypt(plaintext string) (string, error) {
	if s == nil {
		return plaintext, nil
	}
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded ciphertext using AES-GCM.
// If the service is nil, returns the input unchanged.
func (s *CryptoService) Decrypt(cipherB64 string) (string, error) {
	if s == nil {
		return cipherB64, nil
	}
	if cipherB64 == "" {
		return "", nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// Hash computes an HMAC-SHA256 hash of the value using the service's salt.
// Used for deterministic one-way hashing (e.g., CPF deduplication lookups).
func (s *CryptoService) Hash(value string) string {
	mac := hmac.New(sha256.New, s.salt)
	mac.Write([]byte(value))
	return fmt.Sprintf("%x", mac.Sum(nil))
}
