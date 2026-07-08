package services

import (
	"os"
	"testing"
)

func TestCryptoServiceEncryptDecrypt(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "test-key-for-crypto-test-1234567890!!")
	defer os.Unsetenv("ENCRYPTION_KEY")

	crypto := NewCryptoService()
	if crypto == nil {
		t.Fatalf("CryptoService não deveria ser nil com chave definida")
	}

	original := "12345678901"
	encrypted, err := crypto.Encrypt(original)
	if err != nil {
		t.Fatalf("erro ao criptografar: %v", err)
	}

	if encrypted == original {
		t.Fatalf("texto criptografado não pode ser igual ao original")
	}

	decrypted, err := crypto.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("erro ao descriptografar: %v", err)
	}

	if decrypted != original {
		t.Fatalf("descriptografia não retornou o original: esperado '%s', got '%s'", original, decrypted)
	}
}

func TestCryptoServiceHash(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "test-key-for-crypto-test-1234567890!!")
	defer os.Unsetenv("ENCRYPTION_KEY")

	crypto := NewCryptoService()

	hash1 := crypto.Hash("12345678901")
	hash2 := crypto.Hash("12345678901")
	hash3 := crypto.Hash("99999999999")

	if hash1 == "" {
		t.Fatalf("hash não pode ser vazio")
	}
	if hash1 != hash2 {
		t.Fatalf("mesmo CPF deve gerar mesmo hash")
	}
	if hash1 == hash3 {
		t.Fatalf("CPFs diferentes devem gerar hashes diferentes")
	}
}

func TestCryptoServiceEncryptEmptyString(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "test-key-for-crypto-test-1234567890!!")
	defer os.Unsetenv("ENCRYPTION_KEY")

	crypto := NewCryptoService()

	encrypted, err := crypto.Encrypt("")
	if err != nil {
		t.Fatalf("erro ao criptografar vazio: %v", err)
	}
	if encrypted != "" {
		t.Fatalf("string vazia criptografada deveria ser vazia")
	}
}

func TestCryptoServiceDecryptInvalidBase64(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "test-key-for-crypto-test-1234567890!!")
	defer os.Unsetenv("ENCRYPTION_KEY")

	crypto := NewCryptoService()

	_, err := crypto.Decrypt("isso-nao-e-base64-valido!!!")
	if err == nil {
		t.Fatalf("decrypt de base64 inválido deveria retornar erro")
	}
}

func TestCryptoServiceNilService(t *testing.T) {
	var crypto *CryptoService = nil

	result, err := crypto.Encrypt("12345")
	if err != nil {
		t.Fatalf("crypto nil não deveria retornar erro no encrypt: %v", err)
	}
	if result != "12345" {
		t.Fatalf("crypto nil deveria retornar texto original no encrypt")
	}

	result, err = crypto.Decrypt("12345")
	if err != nil {
		t.Fatalf("crypto nil não deveria retornar erro no decrypt: %v", err)
	}
	if result != "12345" {
		t.Fatalf("crypto nil deveria retornar texto original no decrypt")
	}
}

func TestCryptoServiceWithoutKey(t *testing.T) {
	os.Unsetenv("ENCRYPTION_KEY")

	crypto := NewCryptoService()
	if crypto != nil {
		t.Fatalf("CryptoService sem chave deveria ser nil")
	}
}
