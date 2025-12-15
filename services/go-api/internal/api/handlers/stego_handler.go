package handlers

import (
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/service/stenography_service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StegoHandler struct {
	service *stenography_service.AlgoService
}

func NewStegoHandler() *StegoHandler {
	return &StegoHandler{
		service: stenography_service.NewAlgoService(),
	}
}

func (h *StegoHandler) Hide(c *gin.Context) {
	message := c.PostForm("message")
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}

	resultImg, err := h.service.HideData(file, message)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=hidden.jpg")
	c.Data(http.StatusOK, "image/jpeg", resultImg)
}

func (h *StegoHandler) Extract(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image file is required"})
		return
	}

	message, err := h.service.ExtractData(file)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "failed to extract: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": message,
	})
}
