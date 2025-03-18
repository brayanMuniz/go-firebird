package handlers

import (
	"fmt"
	"go-firebird/db"
	"go-firebird/nlp"
	"go-firebird/types"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

func TestLocationSentimentUpdate(c *gin.Context, firestoreClient *firestore.Client) {

	allGood := true
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

			newSentiment := types.AvgLocationSentiment{
				TimeStamp:        end,
				SkeetsAmount:     len(skeets),
				AverageSentiment: avg,
			}
			log.Printf("TimeStamp is: %v", newSentiment.TimeStamp)
			log.Printf("Skeets amount is: %v", newSentiment.SkeetsAmount)
			log.Printf("Average is: %v", newSentiment.AverageSentiment)
			newLocationErr := db.InitLocationSentiment(firestoreClient, testLocationId, newSentiment)
			if newLocationErr != nil {
				log.Printf("Error adding new location sentiment")
				allGood = false
			}

		} else {
			// TODO: check if there are any new skeets
			// if there are calculate the new sentiment average and add a new one to the list
			// else update the most recent sentiment with a new timestamp

			log.Printf("Sentiment list is not empty, will not INIT")
			err := db.UpdateLocSentimentTimestampWithData(firestoreClient, locationData, testLocationId, end)
			if err != nil {
				log.Printf("Failed to update avgLocation field")
			}
		}

	} else {

		// get all the valid locations
		validLocations, err := db.GetValidLocations(firestoreClient)
		if err != nil {
			log.Printf("Error fetching valid locations: %v", err)
			return
		}

		// TODO: 2: do the same as the test one but on ALL valid documents
		for _, v := range validLocations {
			l := len(v.AvgSentimentList)
			if l == 0 {

			} else {
				// latestSentiment := v.AvgSentimentList[l-1]
			}

		}

	}

	mock := map[string]interface{}{
		"data": allGood,
	}

	// Return JSON
	c.JSON(http.StatusOK, mock)
}
