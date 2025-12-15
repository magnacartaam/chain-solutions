package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	service "github.com/magnacartaam/chain-solutions/services/go-api/internal/service/hash/sha1_service"
)

type SHA1HashRequest struct {
	Message string `json:"message"` // binding:"required"`
}

type SHA1HashResponse struct {
	Hash      string `json:"hash"`
	Algorithm string `json:"algorithm"`
	Size      int    `json:"size"`
}

type SHA1VerifyRequest struct {
	Message      string `json:"message" binding:"required"`
	ExpectedHash string `json:"expected_hash" binding:"required"`
}

type SHA1VerifyResponse struct {
	IsValid   bool   `json:"is_valid"`
	Algorithm string `json:"algorithm"`
}

type SHA1MultipleHashRequest struct {
	Message    string `json:"message" binding:"required"`
	Iterations int    `json:"iterations" binding:"required,min=1"`
}

type SHA1MultipleHashResponse struct {
	Hash       string `json:"hash"`
	Iterations int    `json:"iterations"`
	Algorithm  string `json:"algorithm"`
}

func SHA1HashHandler(c *gin.Context) {
	var req SHA1HashRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	hash, err := service.ProcessSHA1Hash(req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to compute hash: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SHA1HashResponse{
		Hash:      hash,
		Algorithm: "SHA-1",
		Size:      160,
	})
}

func SHA1VerifyHandler(c *gin.Context) {
	var req SHA1VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	isValid, err := service.VerifySHA1Hash(req.Message, req.ExpectedHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to verify hash: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SHA1VerifyResponse{
		IsValid:   isValid,
		Algorithm: "SHA-1",
	})
}

func SHA1MultipleHashHandler(c *gin.Context) {
	var req SHA1MultipleHashRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	if req.Iterations < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "iterations must be at least 1",
		})
		return
	}

	maxIterations := 10000
	if req.Iterations > maxIterations {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "iterations exceeds maximum allowed (" + strconv.Itoa(maxIterations) + ")",
		})
		return
	}

	hash, err := service.ProcessSHA1HashMultiple(req.Message, req.Iterations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to compute hash: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SHA1MultipleHashResponse{
		Hash:       hash,
		Iterations: req.Iterations,
		Algorithm:  "SHA-1",
	})
}

func SHA1CompareHandler(c *gin.Context) {
	hash1 := c.Query("hash1")
	hash2 := c.Query("hash2")

	if hash1 == "" || hash2 == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "both hash1 and hash2 query parameters are required",
		})
		return
	}

	isEqual := service.CompareSHA1Hashes(hash1, hash2)

	c.JSON(http.StatusOK, gin.H{
		"are_equal": isEqual,
		"algorithm": "SHA-1",
	})
}
