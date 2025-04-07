package routes

import (
	"cloud.google.com/go/firestore"
	language "cloud.google.com/go/language/apiv2"
	"github.com/gin-gonic/gin"
	"go-firebird/handlers"
	"googlemaps.github.io/maps"
)

func SetupRouter(firestoreClient *firestore.Client, nlpClient *language.Client, geocodeClient *maps.Client) *gin.Engine {
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

	r.GET("/api/testing/geocoding", func(c *gin.Context) {
		handlers.TestGeocode(c)
	})

	r.GET("/api/testing/updateGeocoding", func(c *gin.Context) {
		handlers.TestUpdateGeocodingDBTest(c, firestoreClient)
	})

	r.GET("/api/testing/updateLocationSentiment", func(c *gin.Context) {
		handlers.TestLocationSentimentUpdate(c, firestoreClient)
	})

	r.GET("/api/testing/getActiveLocations", func(c *gin.Context) {
		handlers.TestGetActiveLocation(c, firestoreClient)
	})

	r.GET("/api/export/locations", func(c *gin.Context) {
		handlers.ExportLocationsHandler(c, firestoreClient)
	})

	r.GET("/api/test/disasterDetection", func(c *gin.Context) {
		handlers.RunDisasterDetection(c, firestoreClient)
	})

	r.GET("/api/demo/disaster/add", func(c *gin.Context) {
		handlers.AddDisasterDemoData(c, firestoreClient, nlpClient)
	})

	r.GET("/api/demo/disaster/delete", func(c *gin.Context) {
		handlers.DeleteDisasterDemoData(c, firestoreClient)
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
