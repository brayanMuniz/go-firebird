package routes

import (
	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"go-firebird/handlers"
)

func SetupRouter(firestoreClient *firestore.Client) *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, welcome to Go Firebird!",
		})
	})

	// Inject Firestore client into handler
	r.GET("/api/firebird/bluesky", func(c *gin.Context) {
		handlers.FetchBlueskyHandler(c, firestoreClient)
	})

	// api routes
	api := r.Group("/api/firebird")
	{
		api.POST("/classify", handlers.CallOpenAI)
		api.GET("/demo", handlers.GetDemoData)
		api.GET("/testhook", handlers.TestClientHook)
		api.GET("/simulate", handlers.SimulateDisasterTweets)
	}

	return r
}
