package server

import (
	"encoding/hex"
	"strings"
	"testing"
)

// resetKey clears the package-level encryption key for test isolation.
func resetKey(t *testing.T) {
	t.Helper()
	prev := encryptionKey
	t.Cleanup(func() { encryptionKey = prev })
	encryptionKey = nil
}

func TestEncryptionKeyLoaded_FalseWhenNil(t *testing.T) {
	resetKey(t)
	if EncryptionKeyLoaded() {
		t.Error("expected EncryptionKeyLoaded() == false when key is nil")
	}
}

func TestEncryptionKeyLoaded_TrueAfterLoad(t *testing.T) {
	resetKey(t)
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	t.Setenv("TF_AGENT_ENCRYPTION_KEY", hex.EncodeToString(key))

	if err := LoadEncryptionKey(); err != nil {
		t.Fatalf("LoadEncryptionKey: %v", err)
	}
	if !EncryptionKeyLoaded() {
		t.Error("expected EncryptionKeyLoaded() == true after LoadEncryptionKey")
	}
}

func TestLoadEncryptionKey_FromEnv(t *testing.T) {
	resetKey(t)
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 7)
	}
	t.Setenv("TF_AGENT_ENCRYPTION_KEY", hex.EncodeToString(key))

	if err := LoadEncryptionKey(); err != nil {
		t.Fatalf("LoadEncryptionKey: %v", err)
	}
	if !EncryptionKeyLoaded() {
		t.Error("key should be loaded after reading from env")
	}
}

func TestLoadEncryptionKey_FromEnv_BadHex(t *testing.T) {
	resetKey(t)
	t.Setenv("TF_AGENT_ENCRYPTION_KEY", "not-valid-hex!!")

	err := LoadEncryptionKey()
	if err == nil {
		t.Error("expected error for invalid hex key")
	}
}

func TestLoadEncryptionKey_FromEnv_WrongLength(t *testing.T) {
	resetKey(t)
	// 16 bytes (32 hex chars) — too short for AES-256.
	key := make([]byte, 16)
	t.Setenv("TF_AGENT_ENCRYPTION_KEY", hex.EncodeToString(key))

	err := LoadEncryptionKey()
	if err == nil {
		t.Error("expected error for key that is not 32 bytes")
	}
}

// setTestKey sets a deterministic 32-byte encryption key for the duration of the test.
func setTestKey(t *testing.T) {
	t.Helper()
	resetKey(t)
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	encryptionKey = key
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	setTestKey(t)
	plaintext := "hello, terraform world"

	ciphertext, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if ciphertext == "" {
		t.Fatal("Encrypt returned empty ciphertext for non-empty input")
	}

	recovered, err := Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if recovered != plaintext {
		t.Errorf("round-trip mismatch: got %q, want %q", recovered, plaintext)
	}
}

func TestEncryptDecrypt_EmptyString(t *testing.T) {
	setTestKey(t)

	ciphertext, err := Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt empty: %v", err)
	}
	if ciphertext != "" {
		t.Errorf("Encrypt empty string should return empty, got %q", ciphertext)
	}

	recovered, err := Decrypt("")
	if err != nil {
		t.Fatalf("Decrypt empty: %v", err)
	}
	if recovered != "" {
		t.Errorf("Decrypt empty string should return empty, got %q", recovered)
	}
}

func TestEncrypt_ProducesUniqueOutputs(t *testing.T) {
	setTestKey(t)
	// Each call uses a fresh random nonce, so two encryptions of the same
	// plaintext should not be identical.
	c1, err := Encrypt("same input")
	if err != nil {
		t.Fatalf("Encrypt #1: %v", err)
	}
	c2, err := Encrypt("same input")
	if err != nil {
		t.Fatalf("Encrypt #2: %v", err)
	}
	if c1 == c2 {
		t.Error("two Encrypt calls with the same input produced identical ciphertext (nonce re-use?)")
	}
}

func TestDecrypt_InvalidCiphertext(t *testing.T) {
	setTestKey(t)

	_, err := Decrypt("this-is-not-valid-base64!!!")
	if err == nil {
		t.Error("expected error decrypting garbage input")
	}
}

func TestDecrypt_TooShort(t *testing.T) {
	setTestKey(t)

	// Valid base64 but fewer bytes than the GCM nonce size (12 bytes).
	short := "aGk=" // base64("hi") — only 2 bytes
	_, err := Decrypt(short)
	if err == nil {
		t.Error("expected error for ciphertext that is too short")
	}
	if !strings.Contains(err.Error(), "too short") {
		t.Errorf("expected 'too short' in error, got: %v", err)
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	setTestKey(t)
	ciphertext, err := Encrypt("secret data")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	// Switch to a different key.
	newKey := make([]byte, 32)
	for i := range newKey {
		newKey[i] = byte(255 - i)
	}
	encryptionKey = newKey

	_, err = Decrypt(ciphertext)
	if err == nil {
		t.Error("expected error when decrypting with a different key")
	}
}

func TestEncrypt_NoKeyLoaded(t *testing.T) {
	resetKey(t)

	_, err := Encrypt("something")
	if err == nil {
		t.Error("expected error from Encrypt when key is not loaded")
	}
}

func TestDecrypt_NoKeyLoaded(t *testing.T) {
	resetKey(t)

	_, err := Decrypt("c29tZXRoaW5n") // valid base64, but no key
	if err == nil {
		t.Error("expected error from Decrypt when key is not loaded")
	}
}
