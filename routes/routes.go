package routes

import (
	"cloud.google.com/go/firestore"
	language "cloud.google.com/go/language/apiv2"
	"github.com/gin-gonic/gin"
	"go-firebird/handlers"
)

func SetupRouter(firestoreClient *firestore.Client, nlpClient *language.Client) *gin.Engine {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello, welcome to Go Firebird!",
		})
	})

	r.GET("/api/firebird/bluesky", func(c *gin.Context) {
		handlers.FetchBlueskyHandler(c, firestoreClient, nlpClient)
	})

	// Testing routes
	r.GET("/api/testing/entity", func(c *gin.Context) {
		handlers.TestEntity(c, nlpClient)
	})

	r.GET("/api/testing/sentiment", func(c *gin.Context) {
		handlers.TestSentiment(c, nlpClient)
	})

	r.GET("/api/testing/classification", func(c *gin.Context) {
		handlers.ClassifyLiveTest(c)
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
