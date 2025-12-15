package mceliece_service

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/magnacartaam/chain-solutions/services/go-api/internal/cipher/mceliece"
)

func ProcessMcElieceEncrypt(plainText string, pubKey *mceliece.PublicKey) (string, error) {
	messageBytes := []byte(plainText)
	k := pubKey.Params.K

	blockSize := (k / 8)
	if blockSize <= 0 {
		return "", fmt.Errorf("key size k=%d is too small to encrypt any data", k)
	}

	var encryptedBlocks [][]byte

	for i := 0; i < len(messageBytes); i += blockSize {
		end := i + blockSize
		if end > len(messageBytes) {
			end = len(messageBytes)
		}
		block := messageBytes[i:end]

		encryptedBlock, err := mceliece.Encrypt(block, pubKey)
		if err != nil {
			return "", fmt.Errorf("failed to encrypt block %d: %w", i/blockSize, err)
		}
		encryptedBlocks = append(encryptedBlocks, encryptedBlock)
	}

	var encodedBlocks []string
	for _, block := range encryptedBlocks {
		encodedBlocks = append(encodedBlocks, base64.StdEncoding.EncodeToString(block))
	}

	return strings.Join(encodedBlocks, ":"), nil
}

func ProcessMcElieceDecrypt(cipherText string, privKey *mceliece.PrivateKey) (string, error) {
	encodedBlocks := strings.Split(cipherText, ":")
	var decryptedBytes []byte

	for i, blockB64 := range encodedBlocks {
		if blockB64 == "" {
			continue
		}

		cipherBytes, err := base64.StdEncoding.DecodeString(blockB64)
		if err != nil {
			return "", fmt.Errorf("invalid base64 in block %d: %w", i, err)
		}

		decryptedBlock, err := mceliece.Decrypt(cipherBytes, privKey)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt block %d: %w", i, err)
		}

		decryptedBytes = append(decryptedBytes, decryptedBlock...)
	}

	trimmedBytes := bytes.TrimRight(decryptedBytes, "\x00")

	return string(trimmedBytes), nil
}
