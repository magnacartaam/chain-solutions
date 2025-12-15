package gost3410_service

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/magnacartaam/chain-solutions/services/go-api/internal/hash/gost3411"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/signature/gost3410"
)

func ProcessGenerateKeyPair(curveType string) (string, string, string, error) {
	var curve *gost3410.EllipticCurve

	switch curveType {
	case "256":
		curve = gost3410.GetStandardCurve256()
	case "512":
		curve = gost3410.GetStandardCurve512()
	default:
		return "", "", "", fmt.Errorf("invalid curve type, must be '256' or '512'")
	}

	privateKey, err := gost3410.GenerateKeyPair(curve)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate key pair: %v", err)
	}

	privateKeyBytes := privateKey.Bytes()
	publicKeyBytes := privateKey.Public.Bytes()

	privateKeyHex := hex.EncodeToString(privateKeyBytes)
	publicKeyHex := hex.EncodeToString(publicKeyBytes)

	return privateKeyHex, publicKeyHex, curveType, nil
}

func ProcessSign(message string, privateKeyHex string, curveType string) (string, string, error) {
	var curve *gost3410.EllipticCurve

	switch curveType {
	case "256":
		curve = gost3410.GetStandardCurve256()
	case "512":
		curve = gost3410.GetStandardCurve512()
	default:
		return "", "", fmt.Errorf("invalid curve type, must be '256' or '512'")
	}

	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", "", fmt.Errorf("invalid private key format: %v", err)
	}

	privateKey, err := gost3410.PrivateKeyFromBytes(curve, privateKeyBytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to load private key: %v", err)
	}

	var hash []byte
	if curveType == "256" {
		hash = gost3411.Hash256([]byte(message))
	} else {
		hash = gost3411.Hash512([]byte(message))
	}

	signature, err := gost3410.Sign(privateKey, hash)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign message: %v", err)
	}

	rHex := hex.EncodeToString(signature.R.Bytes())
	sHex := hex.EncodeToString(signature.S.Bytes())

	return rHex, sHex, nil
}

func ProcessVerify(message string, publicKeyHex string, rHex string, sHex string, curveType string) (bool, error) {
	var curve *gost3410.EllipticCurve

	switch curveType {
	case "256":
		curve = gost3410.GetStandardCurve256()
	case "512":
		curve = gost3410.GetStandardCurve512()
	default:
		return false, fmt.Errorf("invalid curve type, must be '256' or '512'")
	}

	publicKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return false, fmt.Errorf("invalid public key format: %v", err)
	}

	publicKey, err := gost3410.PublicKeyFromBytes(curve, publicKeyBytes)
	if err != nil {
		return false, fmt.Errorf("failed to load public key: %v", err)
	}

	rBytes, err := hex.DecodeString(rHex)
	if err != nil {
		return false, fmt.Errorf("invalid signature R format: %v", err)
	}

	sBytes, err := hex.DecodeString(sHex)
	if err != nil {
		return false, fmt.Errorf("invalid signature S format: %v", err)
	}

	signature := &gost3410.Signature{
		R: new(big.Int).SetBytes(rBytes),
		S: new(big.Int).SetBytes(sBytes),
	}

	var hash []byte
	if curveType == "256" {
		hash = gost3411.Hash256([]byte(message))
	} else {
		hash = gost3411.Hash512([]byte(message))
	}

	isValid := gost3410.Verify(publicKey, hash, signature)

	return isValid, nil
}

func ProcessSignBytes(data []byte, privateKeyHex string, curveType string) (string, string, error) {
	var curve *gost3410.EllipticCurve

	switch curveType {
	case "256":
		curve = gost3410.GetStandardCurve256()
	case "512":
		curve = gost3410.GetStandardCurve512()
	default:
		return "", "", fmt.Errorf("invalid curve type, must be '256' or '512'")
	}

	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return "", "", fmt.Errorf("invalid private key format: %v", err)
	}

	privateKey, err := gost3410.PrivateKeyFromBytes(curve, privateKeyBytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to load private key: %v", err)
	}

	var hash []byte
	if curveType == "256" {
		hash = gost3411.Hash256(data)
	} else {
		hash = gost3411.Hash512(data)
	}

	signature, err := gost3410.Sign(privateKey, hash)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign data: %v", err)
	}

	rHex := hex.EncodeToString(signature.R.Bytes())
	sHex := hex.EncodeToString(signature.S.Bytes())

	return rHex, sHex, nil
}

func ProcessVerifyBytes(data []byte, publicKeyHex string, rHex string, sHex string, curveType string) (bool, error) {
	var curve *gost3410.EllipticCurve

	switch curveType {
	case "256":
		curve = gost3410.GetStandardCurve256()
	case "512":
		curve = gost3410.GetStandardCurve512()
	default:
		return false, fmt.Errorf("invalid curve type, must be '256' or '512'")
	}

	publicKeyBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return false, fmt.Errorf("invalid public key format: %v", err)
	}

	publicKey, err := gost3410.PublicKeyFromBytes(curve, publicKeyBytes)
	if err != nil {
		return false, fmt.Errorf("failed to load public key: %v", err)
	}

	rBytes, err := hex.DecodeString(rHex)
	if err != nil {
		return false, fmt.Errorf("invalid signature R format: %v", err)
	}

	sBytes, err := hex.DecodeString(sHex)
	if err != nil {
		return false, fmt.Errorf("invalid signature S format: %v", err)
	}

	signature := &gost3410.Signature{
		R: new(big.Int).SetBytes(rBytes),
		S: new(big.Int).SetBytes(sBytes),
	}

	var hash []byte
	if curveType == "256" {
		hash = gost3411.Hash256(data)
	} else {
		hash = gost3411.Hash512(data)
	}

	isValid := gost3410.Verify(publicKey, hash, signature)

	return isValid, nil
}
