package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

// deriveKey returns a 32-byte key from env MASTER_KEY (base64 or raw)
func deriveKey() ([]byte, error) {
	raw := os.Getenv("AUTH_MASTER_KEY")
	if raw == "" {
		return nil, errors.New("AUTH_MASTER_KEY not set")
	}
	// Try base64 first
	if kb, err := base64.StdEncoding.DecodeString(raw); err == nil && (len(kb) == 32 || len(kb) == 16 || len(kb) == 24) {
		return kb, nil
	}
	// Otherwise hash to 32 bytes
	sum := sha256.Sum256([]byte(raw))
	return sum[:], nil
}

// Encrypt encrypts plaintext using AES-GCM
func Encrypt(plaintext []byte) ([]byte, error) {
	key, err := deriveKey()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ct := aead.Seal(nonce, nonce, plaintext, nil)
	return ct, nil
}

// Decrypt decrypts ciphertext using AES-GCM
func Decrypt(ciphertext []byte) ([]byte, error) {
	key, err := deriveKey()
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ns := aead.NonceSize()
	if len(ciphertext) < ns {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ct := ciphertext[:ns], ciphertext[ns:]
	return aead.Open(nil, nonce, ct, nil)
}
