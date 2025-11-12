package stb_service

import (
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/magnacartaam/chain-solutions/services/go-api/internal/cipher/stb"
)

func ProcessCipherRequest(plainText, key, iv []byte) ([]byte, []byte, error) {
	stbStruct, err := stb.New(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize cipher: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	var ecbResult, cfbResult []byte

	go func() {
		defer wg.Done()
		ecbResult = stbStruct.EncryptECB(plainText)
	}()

	go func() {
		defer wg.Done()
		cfbResult = stbStruct.EncryptCFB(plainText, iv)
	}()

	wg.Wait()

	return ecbResult, cfbResult, nil
}

func ProcessDecipherRequest(ecbCiphertextB64, cfbCiphertextB64 string, key, iv []byte) (string, string, error) {
	stbStruct, err := stb.New(key)
	if err != nil {
		return "", "", fmt.Errorf("failed to initialize cipher: %w", err)
	}

	ecbCipherBytes, err := base64.StdEncoding.DecodeString(ecbCiphertextB64)
	if err != nil {
		return "", "", fmt.Errorf("invalid base64 for ecb ciphertext: %w", err)
	}
	cfbCipherBytes, err := base64.StdEncoding.DecodeString(cfbCiphertextB64)
	if err != nil {
		return "", "", fmt.Errorf("invalid base64 for cfb ciphertext: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)
	var ecbDecrypted, cfbDecrypted []byte
	var ecbErr, cfbErr error

	go func() {
		defer wg.Done()
		ecbDecrypted, ecbErr = stbStruct.DecryptECB(ecbCipherBytes)
	}()

	go func() {
		defer wg.Done()
		cfbDecrypted = stbStruct.DecryptCFB(cfbCipherBytes, iv)
	}()

	wg.Wait()

	if ecbErr != nil {
		return "", "", fmt.Errorf("failed to decrypt ECB: %w", ecbErr)
	}
	if cfbErr != nil {
		return "", "", fmt.Errorf("failed to decrypt CFB: %w", cfbErr)
	}

	return string(ecbDecrypted), string(cfbDecrypted), nil
}
