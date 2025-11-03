package main

import (
	"github.com/magnacartaam/chain-solutions/go-backend/internal/api"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	api.RegisterRoutes(router)

	err := router.Run(":8080")
	if err != nil {
		return
	}
}
