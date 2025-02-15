package routes

import (
	"github.com/gin-gonic/gin"
	"go-firebird/handlers"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, welcome to Go Firebird!",
		})
	})

	// api routes
	api := r.Group("/api/firebird")
	{
		api.GET("/bluesky", handlers.FetchBluesky)
		api.POST("/classify", handlers.CallOpenAI)
		api.GET("/demo", handlers.GetDemoData)
		api.GET("/testhook", handlers.TestClientHook)
	}

	return r
}
