package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

var cipherKey string

func InitEncryptionKey(key string) {
	cipherKey = key
}

// MaskData encrypts plain text using AES-GCM
func MaskData(data any) ([]byte, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error serializing data: %w", err)
	}

	keyBytes := []byte(cipherKey)
	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	cipherText := aesGCM.Seal(nonce, nonce, raw, nil)

	return cipherText, nil
}

// UnmaskData decrypts AES-GCM encrypted text into out
func UnmaskData(encryptedText string, out any) error {

	keyBytes := []byte(cipherKey)
	cipherText := []byte(encryptedText)

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	// extract nonce from encrypted data
	nonceSize := aesGCM.NonceSize()
	if len(cipherText) < nonceSize {
		return errors.New("invalid cipher text")
	}
	nonce, cipherText := cipherText[:nonceSize], cipherText[nonceSize:]

	// Decrypt the data
	raw, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("error deserialising data: %w", err)
	}
	return nil
}

func VerifyUserAuthenticatorCode(userID int, code string) error {
	return nil
}
