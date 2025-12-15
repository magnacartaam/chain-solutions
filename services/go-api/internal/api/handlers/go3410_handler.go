package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	service "github.com/magnacartaam/chain-solutions/services/go-api/internal/service/signature/gost3410_service"
)

type GOST3410KeygenResponse struct {
	PrivateKey string `json:"private_key"`
	PublicKey  string `json:"public_key"`
	CurveType  string `json:"curve_type"`
	Algorithm  string `json:"algorithm"`
}

type GOST3410SignRequest struct {
	Message    string `json:"message" binding:"required"`
	PrivateKey string `json:"private_key" binding:"required"`
	CurveType  string `json:"curve_type" binding:"required"`
}

type GOST3410SignResponse struct {
	SignatureR string `json:"signature_r"`
	SignatureS string `json:"signature_s"`
	CurveType  string `json:"curve_type"`
	Algorithm  string `json:"algorithm"`
}

type GOST3410VerifyRequest struct {
	Message    string `json:"message" binding:"required"`
	PublicKey  string `json:"public_key" binding:"required"`
	SignatureR string `json:"signature_r" binding:"required"`
	SignatureS string `json:"signature_s" binding:"required"`
	CurveType  string `json:"curve_type" binding:"required"`
}

type GOST3410VerifyResponse struct {
	IsValid   bool   `json:"is_valid"`
	CurveType string `json:"curve_type"`
	Algorithm string `json:"algorithm"`
}

func GOST3410KeygenHandler(c *gin.Context) {
	curveType := c.DefaultQuery("curve_type", "256")

	if curveType != "256" && curveType != "512" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "curve_type must be '256' or '512'",
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

	c.JSON(http.StatusOK, GOST3410KeygenResponse{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		CurveType:  curve,
		Algorithm:  "GOST 34.10-2012",
	})
}

func GOST3410SignHandler(c *gin.Context) {
	var req GOST3410SignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	if req.CurveType != "256" && req.CurveType != "512" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "curve_type must be '256' or '512'",
		})
		return
	}

	signatureR, signatureS, err := service.ProcessSign(req.Message, req.PrivateKey, req.CurveType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to sign message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GOST3410SignResponse{
		SignatureR: signatureR,
		SignatureS: signatureS,
		CurveType:  req.CurveType,
		Algorithm:  "GOST 34.10-2012",
	})
}

func GOST3410VerifyHandler(c *gin.Context) {
	var req GOST3410VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body: " + err.Error(),
		})
		return
	}

	if req.CurveType != "256" && req.CurveType != "512" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "curve_type must be '256' or '512'",
		})
		return
	}

	isValid, err := service.ProcessVerify(
		req.Message,
		req.PublicKey,
		req.SignatureR,
		req.SignatureS,
		req.CurveType,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to verify signature: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GOST3410VerifyResponse{
		IsValid:   isValid,
		CurveType: req.CurveType,
		Algorithm: "GOST 34.10-2012",
	})
}

func GOST3410Keygen256Handler(c *gin.Context) {
	privateKey, publicKey, curve, err := service.ProcessGenerateKeyPair("256")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate key pair: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GOST3410KeygenResponse{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		CurveType:  curve,
		Algorithm:  "GOST 34.10-2012 (256-bit)",
	})
}

func GOST3410Keygen512Handler(c *gin.Context) {
	privateKey, publicKey, curve, err := service.ProcessGenerateKeyPair("512")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate key pair: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, GOST3410KeygenResponse{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		CurveType:  curve,
		Algorithm:  "GOST 34.10-2012 (512-bit)",
	})
}
