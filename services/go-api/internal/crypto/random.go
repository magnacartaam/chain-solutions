package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateSeed creates a cryptographically secure 32-byte random hex string.
func GenerateSeed() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashStringSHA256 returns the SHA256 hash of a string.
func HashStringSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// HashDataSHA256 hashes raw bytes.
func HashDataSHA256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}
