package handlers

import (
	"cloud.google.com/go/firestore"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-firebird/db"
	"log"
	"net/http"
)

func TestLocationSentimentUpdate(c *gin.Context, firestoreClient *firestore.Client) {
	// get all the valid locations
	validLocations, err := db.GetValidLocations(firestoreClient)
	if err != nil {
		log.Printf("Error fetching valid locations: %v", err)
		return
	}
	for v, x := range validLocations {
		fmt.Print(v, x)
	}

	// TODO:
	// for each one of those get their /skeets subcollection and return the average with other data (# skeets etx)
	// calculate the new data
	// update the location doc with a new post

	mock := map[string]interface{}{
		"data": validLocations,
	}

	// Return JSON
	c.JSON(http.StatusOK, mock)
}
