package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

// skEncryptionKey is derived from the API_KEY for encrypting SK in Redis.
// This allows SK to be decrypted for signature verification while
// not storing plaintext SK in Redis.
var skEncryptionKey []byte

// InitSKEncryption initializes the SK encryption key from the API key
func InitSKEncryption(apiKey string) {
	hash := sha256.Sum256([]byte(apiKey))
	skEncryptionKey = hash[:]
}

// encryptSK encrypts the SK using AES-GCM
func encryptSK(sk string) (string, error) {
	if len(skEncryptionKey) == 0 {
		return "", fmt.Errorf("SK encryption not initialized")
	}

	block, err := aes.NewCipher(skEncryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(sk), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptSK decrypts the SK using AES-GCM
func decryptSK(encrypted string) (string, error) {
	if len(skEncryptionKey) == 0 {
		return "", fmt.Errorf("SK encryption not initialized")
	}

	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted SK: %w", err)
	}

	block, err := aes.NewCipher(skEncryptionKey)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt SK: %w", err)
	}

	return string(plaintext), nil
}
