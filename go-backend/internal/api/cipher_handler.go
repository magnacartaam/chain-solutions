package api

import (
	"net/http"

	"github.com/magnacartaam/chain-solutions/go-backend/internal/service/cipher"

	"github.com/gin-gonic/gin"
)

type CipherRequest struct {
	PlainText string `json:"plaintext" binding:"required"`
	Key       string `json:"key" binding:"required"`
	IV        string `json:"iv" binding:"required"`
}

type CipherResponse struct {
	OriginalText  string `json:"original_text"`
	EcbCiphertext string `json:"ecb_ciphertext"`
	CfbCiphertext string `json:"cfb_ciphertext"`
}

func CipherHandler(c *gin.Context) {
	var request CipherRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if len(request.Key) != 32 || len(request.IV) != 16 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key or IV length"})
		return
	}

	ecbHex, cfbHex, err := service.ProcessCipherRequest(
		[]byte(request.PlainText),
		[]byte(request.Key),
		[]byte(request.IV),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not process request"})
		return
	}

	response := CipherResponse{
		OriginalText:  request.PlainText,
		EcbCiphertext: ecbHex,
		CfbCiphertext: cfbHex,
	}
	c.JSON(http.StatusOK, response)
}
