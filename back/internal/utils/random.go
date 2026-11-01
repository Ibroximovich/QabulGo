package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// GenerateRandomBytes generates n cryptographically secure random bytes
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomToken generates a URL-safe cryptographically secure random string
func GenerateRandomToken(byteLength int) (string, error) {
	bytes, err := GenerateRandomBytes(byteLength)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Generate6DigitCode generates a cryptographically secure 6-digit numeric OTP code
func Generate6DigitCode() (string, error) {
	maxVal := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, maxVal)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// HashToken returns the SHA-256 hash of a token as a hex string
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
