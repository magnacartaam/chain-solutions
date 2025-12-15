package gost3411_service

import (
	"encoding/hex"
	"fmt"

	"github.com/magnacartaam/chain-solutions/services/go-api/internal/hash/gost3411"
)

func ProcessGostHash(message string, outputSize int) (string, error) {
	if outputSize != 256 && outputSize != 512 {
		return "", fmt.Errorf("output size must be 256 or 512 bits")
	}

	if message == "" {
		return "", fmt.Errorf("message cannot be empty")
	}

	messageBytes := []byte(message)

	var hashBytes []byte
	if outputSize == 512 {
		hashBytes = gost3411.Hash512(messageBytes)
	} else {
		hashBytes = gost3411.Hash256(messageBytes)
	}

	return hex.EncodeToString(hashBytes), nil
}

func ProcessGostHashBytes(data []byte, outputSize int) (string, error) {
	if outputSize != 256 && outputSize != 512 {
		return "", fmt.Errorf("output size must be 256 or 512 bits")
	}

	if len(data) == 0 {
		return "", fmt.Errorf("data cannot be empty")
	}

	var hashBytes []byte
	if outputSize == 512 {
		hashBytes = gost3411.Hash512(data)
	} else {
		hashBytes = gost3411.Hash256(data)
	}

	return hex.EncodeToString(hashBytes), nil
}

func VerifyGostHash(message string, expectedHash string, outputSize int) (bool, error) {
	computedHash, err := ProcessGostHash(message, outputSize)
	if err != nil {
		return false, err
	}

	return computedHash == expectedHash, nil
}
