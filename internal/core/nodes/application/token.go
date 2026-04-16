package application

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func GenerateNodeToken() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return "ntk_" + hex.EncodeToString(buf), nil
}

func HashNodeToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
