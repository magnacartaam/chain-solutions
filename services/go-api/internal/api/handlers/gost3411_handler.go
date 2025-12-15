package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	service "github.com/magnacartaam/chain-solutions/services/go-api/internal/service/hash/gost3411_service"
)

type GostHashRequest struct {
	Message string `json:"message" binding:"required"`
}

type GostHashResponse struct {
	Hash       string `json:"hash"`
	OutputSize int    `json:"output_size"`
	Algorithm  string `json:"algorithm"`
}

type GostVerifyRequest struct {
	Message      string `json:"message" binding:"required"`
	ExpectedHash string `json:"expected_hash" binding:"required"`
}

type GostVerifyResponse struct {
	IsValid    bool   `json:"is_valid"`
	OutputSize int    `json:"output_size"`
	Algorithm  string `json:"algorithm"`
}

func GostHashHandler(c *gin.Context) {
	outputSizeStr := c.DefaultQuery("output_size", "512")
	outputSize, err := strconv.Atoi(outputSizeStr)
	if err != nil || (outputSize != 256 && outputSize != 512) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "output_size must be 256 or 512",
		})
		return
	}

	var req GostHashRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	hash, err := service.ProcessGostHash(req.Message, outputSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to compute hash: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GostHashResponse{
		Hash:       hash,
		OutputSize: outputSize,
		Algorithm:  "GOST 34.11",
	})
}

func GostVerifyHandler(c *gin.Context) {
	outputSizeStr := c.DefaultQuery("output_size", "512")
	outputSize, err := strconv.Atoi(outputSizeStr)
	if err != nil || (outputSize != 256 && outputSize != 512) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "output_size must be 256 or 512",
		})
		return
	}

	var req GostVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	isValid, err := service.VerifyGostHash(req.Message, req.ExpectedHash, outputSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to verify hash: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GostVerifyResponse{
		IsValid:    isValid,
		OutputSize: outputSize,
		Algorithm:  "GOST 34.11",
	})
}

func GostHash256Handler(c *gin.Context) {
	var req GostHashRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	hash, err := service.ProcessGostHash(req.Message, 256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to compute hash: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GostHashResponse{
		Hash:       hash,
		OutputSize: 256,
		Algorithm:  "GOST 34.11-256",
	})
}

func GostHash512Handler(c *gin.Context) {
	var req GostHashRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	hash, err := service.ProcessGostHash(req.Message, 512)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to compute hash: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GostHashResponse{
		Hash:       hash,
		OutputSize: 512,
		Algorithm:  "GOST 34.11-512",
	})
}
