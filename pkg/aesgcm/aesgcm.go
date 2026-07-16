package aesgcm

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// EncryptGCM and DecryptGCM use AES-GCM with a random 12-byte nonce prepended
// to the ciphertext, hex-encoded. The key must be 16, 24, or 32 bytes long
// (AES-128/192/256).

func EncryptGCM(plaintext, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nil, nonce, []byte(plaintext), nil)
	result := append(nonce, ciphertext...)

	return hex.EncodeToString(result), nil
}

func DecryptGCM(encryptedHex, key string) (string, error) {
	if encryptedHex == "" {
		return "", nil
	}

	encrypted, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(encrypted) < 12 {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce := encrypted[:12]
	ciphertext := encrypted[12:]

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	decrypted, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
