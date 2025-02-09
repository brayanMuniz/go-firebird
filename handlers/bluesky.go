package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func FetchBluesky(c *gin.Context) {
	// Logic to fetch trending topics from Bluesky
	c.JSON(http.StatusOK, gin.H{"message": "Trending topics from Bluesky"})
}
