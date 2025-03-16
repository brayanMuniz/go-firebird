package handlers

import (
	"cloud.google.com/go/firestore"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-firebird/db"
	"go-firebird/nlp"
	"log"
	"net/http"
	"time"
)

func TestLocationSentimentUpdate(c *gin.Context, firestoreClient *firestore.Client) {
	testLocationId := c.Query("docId")
	if testLocationId != "" {
		start := "1970-01-01T00:00:00Z" // big bang of computers
		end := time.Now().UTC().Format(time.RFC3339)

		fmt.Println("Running avg sentiment on docId: %s", testLocationId)
		locationData, err := db.GetValidLocation(firestoreClient, testLocationId)
		if err != nil {
			log.Printf("Error fetching valid locations: %v", err)
		}

		if len(locationData.AvgSentimentList) == 0 {
			skeets, err := db.GetSkeetsSubCollection(firestoreClient, testLocationId, start, end)
			if err != nil {
				log.Printf("Error fetching skeets subcollection: %v", err)
				return
			}
			fmt.Printf("Fetched %d skeets\n", len(skeets))

			avg, err := nlp.ComputeSimpleAverageSentiment(skeets)
			if err != nil {
				log.Printf("Error fetching skeets subcollection: %v", err)
				return
			}
			// TODO: 1. update the document with the new avg sentiment and all the fields
			log.Printf("Average is: %v", avg)

		} else {
		}

	} else {

		// get all the valid locations
		validLocations, err := db.GetValidLocations(firestoreClient)
		if err != nil {
			log.Printf("Error fetching valid locations: %v", err)
			return
		}

		// TODO:2: do the same as the test one but on ALL valid documents
		for _, v := range validLocations {
			l := len(v.AvgSentimentList)
			if l == 0 {

			} else {
				// latestSentiment := v.AvgSentimentList[l-1]
			}

		}

	}
	mock := map[string]interface{}{
		"data": true,
	}

	// Return JSON
	c.JSON(http.StatusOK, mock)
}
