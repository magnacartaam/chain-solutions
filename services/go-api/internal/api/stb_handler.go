package api

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/service/cipher/stb_service"
)

type EncryptRequest struct {
	PlainText string `json:"plaintext" binding:"required"`
	Key       string `json:"key" binding:"required"`
	IV        string `json:"iv" binding:"required"`
}

type EncryptResponse struct {
	OriginalText     string `json:"original_text"`
	EcbCiphertextB64 string `json:"ecb_ciphertext_b64"`
	CfbCiphertextB64 string `json:"cfb_ciphertext_b64"`
}

type DecryptRequest struct {
	EcbCiphertextB64 string `json:"ecb_ciphertext_b64" binding:"required"`
	CfbCiphertextB64 string `json:"cfb_ciphertext_b64" binding:"required"`
	Key              string `json:"key" binding:"required"`
	IV               string `json:"iv" binding:"required"`
}

type DecryptResponse struct {
	EcbDecryptedText string `json:"ecb_decrypted_text"`
	CfbDecryptedText string `json:"cfb_decrypted_text"`
}

func CipherHandler(c *gin.Context) {
	var request EncryptRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if len(request.Key) != 32 || len(request.IV) != 16 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key or IV length"})
		return
	}

	ecbResult, cfbResult, err := stb_service.service.ProcessCipherRequest(
		[]byte(request.PlainText),
		[]byte(request.Key),
		[]byte(request.IV),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not process request"})
		return
	}

	response := EncryptResponse{
		OriginalText:     request.PlainText,
		EcbCiphertextB64: base64.StdEncoding.EncodeToString(ecbResult),
		CfbCiphertextB64: base64.StdEncoding.EncodeToString(cfbResult),
	}
	c.JSON(http.StatusOK, response)
}

func DecipherHandler(c *gin.Context) {
	var request DecryptRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if len(request.Key) != 32 || len(request.IV) != 16 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key or IV length"})
		return
	}

	ecbDecrypted, cfbDecrypted, err := stb_service.service.ProcessDecipherRequest(
		request.EcbCiphertextB64,
		request.CfbCiphertextB64,
		[]byte(request.Key),
		[]byte(request.IV),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := DecryptResponse{
		EcbDecryptedText: ecbDecrypted,
		CfbDecryptedText: cfbDecrypted,
	}
	c.JSON(http.StatusOK, response)
}
