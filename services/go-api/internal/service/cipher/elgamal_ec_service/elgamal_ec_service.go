package elgamal_ec_service

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/magnacartaam/chain-solutions/services/go-api/internal/cipher/elgamal_ec"
)

func ProcessGenerateKeyPair(curveType string) (string, string, string, error) {
	var curve *elgamal_ec.EllipticCurve

	switch curveType {
	case "P256", "256":
		curve = elgamal_ec.GetStandardCurveP256()
	case "P384", "384":
		curve = elgamal_ec.GetStandardCurveP384()
	default:
		return "", "", "", fmt.Errorf("invalid curve type, must be 'P256' or 'P384'")
	}

	if curve == nil {
		return "", "", "", fmt.Errorf("failed to initialize curve")
	}

	privateKey, err := elgamal_ec.GenerateKeyPair(curve)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate key pair: %v", err)
	}

	if privateKey == nil || privateKey.Public == nil {
		return "", "", "", fmt.Errorf("key generation returned nil key")
	}

	privateKeyBytes := privateKey.Bytes()
	publicKeyBytes := privateKey.Public.Bytes()

	if len(privateKeyBytes) == 0 || len(publicKeyBytes) == 0 {
		return "", "", "", fmt.Errorf("key export failed: empty key bytes")
	}

	privateKeyHex := hex.EncodeToString(privateKeyBytes)
	publicKeyHex := hex.EncodeToString(publicKeyBytes)

	return privateKeyHex, publicKeyHex, curveType, nil
}

func ProcessEncrypt(plaintext string, publicKeyHex string, curveType string) (string, error) {
	var curve *elgamal_ec.EllipticCurve

	switch curveType {
	case "P256", "256":
		curve = elgamal_ec.GetStandardCurveP256()
	case "P384", "384":
		curve = elgamal_ec.GetStandardCurveP384()
	default:
		return "", fmt.Errorf("invalid curve type, must be 'P256' or 'P384'")
	}

	publicKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid public key format: %v", err)
	}

	publicKey, err := elgamal_ec.PublicKeyFromBytes(curve, publicKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to load public key: %v", err)
	}

	messageBytes := []byte(plaintext)
	ciphertext, err := elgamal_ec.Encrypt(publicKey, messageBytes)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt message: %v", err)
	}

	curveByteLen := (curve.P.BitLen() + 7) / 8
	ciphertextBytes := ciphertext.Bytes(curveByteLen)
	ciphertextB64 := base64.StdEncoding.EncodeToString(ciphertextBytes)

	return ciphertextB64, nil
}

func ProcessDecrypt(ciphertextB64 string, privateKeyHex string, curveType string, messageLen int) (string, error) {
	var curve *elgamal_ec.EllipticCurve

	switch curveType {
	case "P256", "256":
		curve = elgamal_ec.GetStandardCurveP256()
	case "P384", "384":
		curve = elgamal_ec.GetStandardCurveP384()
	default:
		return "", fmt.Errorf("invalid curve type, must be 'P256' or 'P384'")
	}

	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid private key format: %v", err)
	}

	privateKey, err := elgamal_ec.PrivateKeyFromBytes(curve, privateKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to load private key: %v", err)
	}

	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", fmt.Errorf("invalid ciphertext format: %v", err)
	}

	ciphertext, err := elgamal_ec.CipherTextFromBytes(curve, ciphertextBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse ciphertext: %v", err)
	}

	messageBytes, err := elgamal_ec.Decrypt(privateKey, ciphertext, messageLen)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt message: %v", err)
	}

	return string(messageBytes), nil
}

func ProcessEncryptBytes(data []byte, publicKeyHex string, curveType string) (string, error) {
	var curve *elgamal_ec.EllipticCurve

	switch curveType {
	case "P256", "256":
		curve = elgamal_ec.GetStandardCurveP256()
	case "P384", "384":
		curve = elgamal_ec.GetStandardCurveP384()
	default:
		return "", fmt.Errorf("invalid curve type, must be 'P256' or 'P384'")
	}

	publicKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return "", fmt.Errorf("invalid public key format: %v", err)
	}

	publicKey, err := elgamal_ec.PublicKeyFromBytes(curve, publicKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to load public key: %v", err)
	}

	ciphertext, err := elgamal_ec.Encrypt(publicKey, data)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt data: %v", err)
	}

	curveByteLen := (curve.P.BitLen() + 7) / 8
	ciphertextBytes := ciphertext.Bytes(curveByteLen)
	ciphertextB64 := base64.StdEncoding.EncodeToString(ciphertextBytes)

	return ciphertextB64, nil
}

func ProcessDecryptBytes(ciphertextB64 string, privateKeyHex string, curveType string, dataLen int) ([]byte, error) {
	var curve *elgamal_ec.EllipticCurve

	switch curveType {
	case "P256", "256":
		curve = elgamal_ec.GetStandardCurveP256()
	case "P384", "384":
		curve = elgamal_ec.GetStandardCurveP384()
	default:
		return nil, fmt.Errorf("invalid curve type, must be 'P256' or 'P384'")
	}

	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key format: %v", err)
	}

	privateKey, err := elgamal_ec.PrivateKeyFromBytes(curve, privateKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %v", err)
	}

	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return nil, fmt.Errorf("invalid ciphertext format: %v", err)
	}

	ciphertext, err := elgamal_ec.CipherTextFromBytes(curve, ciphertextBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ciphertext: %v", err)
	}

	data, err := elgamal_ec.Decrypt(privateKey, ciphertext, dataLen)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %v", err)
	}

	return data, nil
}
