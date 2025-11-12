package rabin_service

import (
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/magnacartaam/chain-solutions/services/go-api/internal/cipher"
)

func ProcessRabinEncrypt(plainText string, nStr string) (string, error) {
	publicKeyN := new(big.Int)
	publicKeyN, ok := publicKeyN.SetString(nStr, 10)
	if !ok {
		return "", fmt.Errorf("invalid public key format")
	}

	message := new(big.Int).SetBytes([]byte(plainText))

	if message.Cmp(publicKeyN) >= 0 {
		return "", fmt.Errorf("plaintext is too long for the given key size")
	}

	cipherInt := cipher.Encrypt(message, publicKeyN)

	return base64.StdEncoding.EncodeToString(cipherInt.Bytes()), nil
}

func ProcessRabinDecrypt(cipherTextB64 string, nStr, pStr, qStr string) ([]string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(cipherTextB64)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 ciphertext: %w", err)
	}
	cipherInt := new(big.Int).SetBytes(cipherBytes)

	keys := &cipher.RabinKeys{}
	keys.N, _ = new(big.Int).SetString(nStr, 10)
	keys.P, _ = new(big.Int).SetString(pStr, 10)
	keys.Q, _ = new(big.Int).SetString(qStr, 10)

	if keys.N == nil || keys.P == nil || keys.Q == nil {
		return nil, fmt.Errorf("invalid key format")
	}

	candidates := cipher.Decrypt(cipherInt, keys)

	results := make([]string, 4)
	for i, c := range candidates {
		results[i] = string(c.Bytes())
	}

	return results, nil
}
