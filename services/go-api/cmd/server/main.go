package main

import (
	"github.com/gin-gonic/gin"
	"github.com/magnacartaam/chain-solutions/services/go-api/internal/api"
)

func main() {
	router := gin.Default()

	api.RegisterRoutes(router)

	err := router.Run(":8080")
	if err != nil {
		return
	}
}
