package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	service "github.com/magnacartaam/chain-solutions/services/go-api/internal/service/cipher/elgamal_ec_service"
)

type ElGamalECKeygenResponse struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	CurveType  string `json:"curve_type"`
	Algorithm  string `json:"algorithm"`
}

type ElGamalECEncryptRequest struct {
	Plaintext string `json:"plaintext" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
	CurveType string `json:"curve_type" binding:"required"`
}

type ElGamalECEncryptResponse struct {
	CiphertextB64 string `json:"ciphertext_b64"`
	CurveType     string `json:"curve_type"`
	Algorithm     string `json:"algorithm"`
}

type ElGamalECDecryptRequest struct {
	CiphertextB64 string `json:"ciphertext_b64" binding:"required"`
	PrivateKey    string `json:"private_key" binding:"required"`
	CurveType     string `json:"curve_type" binding:"required"`
	MessageLen    int    `json:"message_len"`
}

type ElGamalECDecryptResponse struct {
	Plaintext string `json:"plaintext"`
	CurveType string `json:"curve_type"`
	Algorithm string `json:"algorithm"`
}

// ElGamalECKeygenHandler generates a new ElGamal EC key pair
// Query params:
//   - curve_type: "P256" or "P384" (default: "P256")
func ElGamalECKeygenHandler(c *gin.Context) {
	curveType := c.DefaultQuery("curve_type", "P256")

	if curveType != "P256" && curveType != "P384" && curveType != "256" && curveType != "384" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "curve_type must be 'P256' or 'P384'",
		})
		return
	}

	privateKey, publicKey, curve, err := service.ProcessGenerateKeyPair(curveType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate key pair: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ElGamalECKeygenResponse{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		CurveType:  curve,
		Algorithm:  "ElGamal on Elliptic Curves",
	})
}

// ElGamalECEncryptHandler encrypts a message
// Body:
//   - plaintext: message to encrypt
//   - public_key: recipient's public key in hex format
//   - curve_type: "P256" or "P384"
func ElGamalECEncryptHandler(c *gin.Context) {
	var req ElGamalECEncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	if req.CurveType != "P256" && req.CurveType != "P384" && req.CurveType != "256" && req.CurveType != "384" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "curve_type must be 'P256' or 'P384'",
		})
		return
	}

	ciphertextB64, err := service.ProcessEncrypt(req.Plaintext, req.PublicKey, req.CurveType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to encrypt message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ElGamalECEncryptResponse{
		CiphertextB64: ciphertextB64,
		CurveType:     req.CurveType,
		Algorithm:     "ElGamal on Elliptic Curves",
	})
}

// ElGamalECDecryptHandler decrypts a ciphertext
// Body:
//   - ciphertext_b64: encrypted message in base64 format
//   - private_key: private key in hex format
//   - curve_type: "P256" or "P384"
//   - message_len: optional, original message length for better decoding
func ElGamalECDecryptHandler(c *gin.Context) {
	var req ElGamalECDecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	if req.CurveType != "P256" && req.CurveType != "P384" && req.CurveType != "256" && req.CurveType != "384" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "curve_type must be 'P256' or 'P384'",
		})
		return
	}

	plaintext, err := service.ProcessDecrypt(req.CiphertextB64, req.PrivateKey, req.CurveType, req.MessageLen)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to decrypt message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ElGamalECDecryptResponse{
		Plaintext: plaintext,
		CurveType: req.CurveType,
		Algorithm: "ElGamal on Elliptic Curves",
	})
}

// ElGamalECKeygenP256Handler is a convenience endpoint for P-256 key generation
func ElGamalECKeygenP256Handler(c *gin.Context) {
	privateKey, publicKey, curve, err := service.ProcessGenerateKeyPair("P256")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate key pair: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ElGamalECKeygenResponse{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		CurveType:  curve,
		Algorithm:  "ElGamal on Elliptic Curves (P-256)",
	})
}

// ElGamalECKeygenP384Handler is a convenience endpoint for P-384 key generation
func ElGamalECKeygenP384Handler(c *gin.Context) {
	privateKey, publicKey, curve, err := service.ProcessGenerateKeyPair("P384")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate key pair: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, ElGamalECKeygenResponse{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		CurveType:  curve,
		Algorithm:  "ElGamal on Elliptic Curves (P-384)",
	})
}
