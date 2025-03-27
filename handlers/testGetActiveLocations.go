package handlers

import (
	"go-firebird/db"
	"go-firebird/types"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

func TestGetActiveLocation(c *gin.Context, firestoreClient *firestore.Client) {
	limit := 10
	locDocs, err := db.GetTopLocationsBySkeetAmount(firestoreClient, limit)
	if err != nil {
		log.Printf("ERROR fetching top locations: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve top locations",
		})
		return
	}

	if locDocs == nil {
		locDocs = []types.LocationData{}
	}

	c.JSON(http.StatusOK, locDocs)

}
