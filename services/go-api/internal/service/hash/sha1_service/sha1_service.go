package sha1_service

import (
	"encoding/hex"
	"fmt"

	"github.com/magnacartaam/chain-solutions/services/go-api/internal/hash/sha1"
)

func ProcessSHA1Hash(message string) (string, error) {
	//if message == "" {
	//	return "", fmt.Errorf("message cannot be empty")
	//}

	messageBytes := []byte(message)
	hashBytes := sha1.Hash(messageBytes)

	return hex.EncodeToString(hashBytes), nil
}

func ProcessSHA1HashBytes(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("data cannot be empty")
	}

	hashBytes := sha1.Hash(data)
	return hex.EncodeToString(hashBytes), nil
}

func VerifySHA1Hash(message string, expectedHash string) (bool, error) {
	computedHash, err := ProcessSHA1Hash(message)
	if err != nil {
		return false, err
	}

	return computedHash == expectedHash, nil
}

func ProcessSHA1HashMultiple(message string, iterations int) (string, error) {
	//if message == "" {
	//	return "", fmt.Errorf("message cannot be empty")
	//}

	if iterations <= 0 {
		return "", fmt.Errorf("iterations must be positive")
	}

	current := []byte(message)
	for i := 0; i < iterations; i++ {
		current = sha1.Hash(current)
	}

	return hex.EncodeToString(current), nil
}

func CompareSHA1Hashes(hash1, hash2 string) bool {
	if len(hash1) != len(hash2) {
		return false
	}

	bytes1, err1 := hex.DecodeString(hash1)
	bytes2, err2 := hex.DecodeString(hash2)

	if err1 != nil || err2 != nil {
		return false
	}

	if len(bytes1) != sha1.Size || len(bytes2) != sha1.Size {
		return false
	}

	var result byte
	for i := 0; i < len(bytes1); i++ {
		result |= bytes1[i] ^ bytes2[i]
	}

	return result == 0
}

func GetSHA1Digest() *sha1.Digest {
	return sha1.New()
}
