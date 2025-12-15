package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/cipher/mceliece"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/service/cipher/mceliece_service"
)

type McEliecePublicKeyJSON struct {
	G      [][]int             `json:"g"`
	T      int                 `json:"t"`
	Params mceliece.Parameters `json:"params"`
}

type McEliecePrivateKeyJSON struct {
	S      [][]int             `json:"s"`
	G      [][]int             `json:"g"`
	P      [][]int             `json:"p"`
	Params mceliece.Parameters `json:"params"`
}

type McElieceEncryptRequest struct {
	PlainText string                `json:"plaintext" binding:"required"`
	PublicKey McEliecePublicKeyJSON `json:"public_key" binding:"required"`
}

type McElieceEncryptResponse struct {
	CipherTextB64 string `json:"ciphertext_b64"`
}

type McElieceDecryptRequest struct {
	CipherText string                 `json:"ciphertext" binding:"required"`
	PrivateKey McEliecePrivateKeyJSON `json:"private_key" binding:"required"`
}

type McElieceDecryptResponse struct {
	DecryptedText string `json:"decrypted_text"`
}

// --- Handlers ---

func McElieceKeygenHandler(c *gin.Context) {
	n, _ := strconv.Atoi(c.DefaultQuery("n", "32"))
	k, _ := strconv.Atoi(c.DefaultQuery("k", "16"))
	t, _ := strconv.Atoi(c.DefaultQuery("t", "2"))

	params := mceliece.Parameters{N: n, K: k, T: t}

	keyPair, err := mceliece.GenerateKeys(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate keys: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"public_key":  keyPair.Public,
		"private_key": keyPair.Private,
	})
}

func McElieceEncryptHandler(c *gin.Context) {
	var req McElieceEncryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	fmt.Printf("\n--- McEliece Encrypt ---\n")
	fmt.Printf("[HANDLER] Received PlainText: %q\n", req.PlainText)
	fmt.Printf("[HANDLER] Received Public Key N (start): %.30s...\n", req.PublicKey)

	pubKey := &mceliece.PublicKey{
		G:      req.PublicKey.G,
		T:      req.PublicKey.T,
		Params: req.PublicKey.Params,
	}

	cipherTextB64, err := mceliece_service.ProcessMcElieceEncrypt(req.PlainText, pubKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("[HANDLER] Sending CipherTextB64: %s\n", cipherTextB64)
	c.JSON(http.StatusOK, McElieceEncryptResponse{CipherTextB64: cipherTextB64})
}

func McElieceDecryptHandler(c *gin.Context) {
	var req McElieceDecryptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	privKey := &mceliece.PrivateKey{
		S:      req.PrivateKey.S,
		G:      req.PrivateKey.G,
		P:      req.PrivateKey.P,
		Params: req.PrivateKey.Params,
	}

	decryptedText, err := mceliece_service.ProcessMcElieceDecrypt(req.CipherText, privKey)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, McElieceDecryptResponse{DecryptedText: decryptedText})
}
