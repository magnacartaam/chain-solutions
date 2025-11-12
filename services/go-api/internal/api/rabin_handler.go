package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/cipher"
	service "github.com/magnacartaam/chain-solutions/services/go-api/internal/service/cipher/rabin_service"
)

type RabinKeygenResponse struct {
	PublicKeyN  string `json:"public_key_n"`
	PrivateKeyP string `json:"private_key_p"`
	PrivateKeyQ string `json:"private_key_q"`
}

type RabinEncryptRequest struct {
	PlainText  string `json:"plaintext" binding:"required"`
	PublicKeyN string `json:"public_key_n" binding:"required"`
}

type RabinEncryptResponse struct {
	CipherTextB64 string `json:"ciphertext_b64"`
}

type RabinDecryptRequest struct {
	CipherTextB64 string `json:"ciphertext_b64" binding:"required"`
	PublicKeyN    string `json:"public_key_n" binding:"required"`
	PrivateKeyP   string `json:"private_key_p" binding:"required"`
	PrivateKeyQ   string `json:"private_key_q" binding:"required"`
}

type RabinDecryptResponse struct {
	Candidates []string `json:"candidates"`
}

func RabinKeygenHandler(c *gin.Context) {
	bitsStr := c.DefaultQuery("bits", "1024")
	bits, err := strconv.Atoi(bitsStr)
	if err != nil || bits <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bits parameter"})
		return
	}

	keys, err := rabin.GenerateRabinKeys(bits)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate keys"})
		return
	}

	c.JSON(http.StatusOK, RabinKeygenResponse{
		PublicKeyN:  keys.N.String(),
		PrivateKeyP: keys.P.String(),
		PrivateKeyQ: keys.Q.String(),
	})
}

func RabinEncryptHandler(c *gin.Context) {
	var req RabinEncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	cipherTextB64, err := service.ProcessRabinEncrypt(req.PlainText, req.PublicKeyN)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, RabinEncryptResponse{CipherTextB64: cipherTextB64})
}

func RabinDecryptHandler(c *gin.Context) {
	var req RabinDecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	candidates, err := service.ProcessRabinDecrypt(req.CipherTextB64, req.PublicKeyN, req.PrivateKeyP, req.PrivateKeyQ)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, RabinDecryptResponse{Candidates: candidates})
}
