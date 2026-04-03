package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var encryptionKey []byte

// EncryptionKeyLoaded reports whether the AES encryption key has been loaded.
func EncryptionKeyLoaded() bool { return encryptionKey != nil }

// LoadEncryptionKey loads a 32-byte AES key from TF_AGENT_ENCRYPTION_KEY env var,
// or from ~/.tf-agent/encryption.key, generating and persisting one if absent.
func LoadEncryptionKey() error {
	if keyHex := os.Getenv("TF_AGENT_ENCRYPTION_KEY"); keyHex != "" {
		key, err := hex.DecodeString(strings.TrimSpace(keyHex))
		if err != nil || len(key) != 32 {
			return fmt.Errorf("TF_AGENT_ENCRYPTION_KEY must be 64 hex chars (32 bytes), got %d bytes", len(key))
		}
		encryptionKey = key
		return nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("user home dir: %w", err)
	}
	keyPath := filepath.Join(home, ".tf-agent", "encryption.key")

	data, err := os.ReadFile(keyPath)
	if err == nil {
		key, err := hex.DecodeString(strings.TrimSpace(string(data)))
		if err == nil && len(key) == 32 {
			encryptionKey = key
			return nil
		}
	}

	// Generate a new key and persist it.
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("generate encryption key: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(keyPath), 0700); err != nil {
		return fmt.Errorf("create key dir: %w", err)
	}
	if err := os.WriteFile(keyPath, []byte(hex.EncodeToString(key)), 0600); err != nil {
		return fmt.Errorf("persist encryption key: %w", err)
	}
	encryptionKey = key
	fmt.Printf("generated new encryption key at %s\n", keyPath)
	return nil
}

// Encrypt encrypts plaintext with AES-256-GCM and returns a base64-encoded ciphertext.
// Returns empty string unchanged.
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	if encryptionKey == nil {
		return "", fmt.Errorf("encryption key not loaded")
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

// Decrypt decrypts a base64-encoded AES-256-GCM ciphertext.
// Returns empty string unchanged.
func Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	if encryptionKey == nil {
		return "", fmt.Errorf("encryption key not loaded")
	}
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphered := data[:gcm.NonceSize()], data[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphered, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plain), nil
}
