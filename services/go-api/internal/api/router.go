package api

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.Engine) {
	apiGroup := router.Group("/api")
	{
		stbGroup := apiGroup.Group("/stb")
		{
			stbGroup.POST("/cipher", CipherHandler)
			stbGroup.POST("/decipher", DecipherHandler)
		}

		rabinGroup := apiGroup.Group("/rabin")
		{
			rabinGroup.GET("/keygen", RabinKeygenHandler)
			rabinGroup.POST("/encrypt", RabinEncryptHandler)
			rabinGroup.POST("/decrypt", RabinDecryptHandler)
		}
	}
}
