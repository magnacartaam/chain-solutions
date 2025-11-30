package mceliece_service

import (
	"encoding/base64"
	"fmt"

	"github.com/magnacartaam/chain-solutions/services/go-api/internal/cipher/mceliece"
)

func ProcessMcElieceEncrypt(plainText string, pubKey *mceliece.PublicKey) (string, error) {
	messageBytes := []byte(plainText)

	cipherBytes, err := mceliece.Encrypt(messageBytes, pubKey)
	if err != nil {
		return "", fmt.Errorf("encryption failed: %w", err)
	}

	return base64.StdEncoding.EncodeToString(cipherBytes), nil
}

func ProcessMcElieceDecrypt(cipherTextB64 string, privKey *mceliece.PrivateKey) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(cipherTextB64)
	if err != nil {
		return "", fmt.Errorf("invalid base64 ciphertext: %w", err)
	}

	decryptedBytes, err := mceliece.Decrypt(cipherBytes, privKey)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(decryptedBytes), nil
}
