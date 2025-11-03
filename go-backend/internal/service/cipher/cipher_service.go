package service

import (
	"fmt"
	"github.com/magnacartaam/chain-solutions/go-backend/internal/ciphers"
	"sync"
)

func ProcessCipherRequest(plainText, key, iv []byte) (string, string, error) {
	stb, err := cipher.New(key)
	if err != nil {
		return "", "", fmt.Errorf("failed to initialize cipher: %w", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	var ecbResult, cfbResult []byte

	go func() {
		defer wg.Done()
		ecbResult = stb.EncryptECB(plainText)
	}()

	go func() {
		defer wg.Done()
		cfbResult = stb.EncryptCFB(plainText, iv)
	}()

	wg.Wait()

	ecbHex := fmt.Sprintf("%x", ecbResult)
	cfbHex := fmt.Sprintf("%x", cfbResult)

	return ecbHex, cfbHex, nil
}
