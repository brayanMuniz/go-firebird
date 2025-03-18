package handlers

import (
	"fmt"
	"go-firebird/db"
	"go-firebird/nlp"
	"go-firebird/types"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

func TestLocationSentimentUpdate(c *gin.Context, firestoreClient *firestore.Client) {

	allGood := true
	testLocationId := strings.TrimSpace(c.Query("docId"))
	if testLocationId != "" {
		start := "1970-01-01T00:00:00Z" // big bang of computers
		end := time.Now().UTC().Format(time.RFC3339)

		log.Printf("Running avg sentiment on docId: %s", testLocationId)
		log.Printf("Using range %v | %v", start, end)

		locationData, err := db.GetValidLocation(firestoreClient, testLocationId)
		if err != nil {
			log.Printf("Error fetching valid location: %v", err)
			return
		}
		log.Printf("Location name: %v", locationData.FormattedAddress)

		if len(locationData.AvgSentimentList) == 0 {
			fmt.Println()
			log.Printf("Sentiment list is empty. Will INIT")
			skeets, err := db.GetSkeetsSubCollection(firestoreClient, testLocationId, start, end)
			if err != nil {
				log.Printf("Error fetching skeets subcollection: %v", err)
				return
			}
			fmt.Printf("Fetched %d skeets\n", len(skeets))

			avg := nlp.ComputeSimpleAverageSentiment(skeets)
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
			log.Printf("Sentiment list is not empty")
			latestSentiment := locationData.AvgSentimentList[len(locationData.AvgSentimentList)-1]
			newSkeets, err := db.GetSkeetsSubCollection(firestoreClient, testLocationId, latestSentiment.TimeStamp, end)
			if err != nil {
				log.Printf("Failed to get skeets subcolleciton", err)
			}

			if len(newSkeets) > 0 {
				log.Printf("New skeets to average. Adding new average")
				newSkeetsAvg := nlp.ComputeSimpleAverageSentiment(newSkeets)
				newAvg := ((latestSentiment.AverageSentiment * float32(latestSentiment.SkeetsAmount)) + newSkeetsAvg) / ((float32(latestSentiment.SkeetsAmount)) + float32(len(newSkeets)))

				newSentiment := types.AvgLocationSentiment{
					TimeStamp:        end,
					SkeetsAmount:     latestSentiment.SkeetsAmount + len(newSkeets),
					AverageSentiment: newAvg,
				}
				log.Printf("TimeStamp is: %v", newSentiment.TimeStamp)
				log.Printf("Skeets amount is: %v", newSentiment.SkeetsAmount)
				log.Printf("Average is: %v", newSentiment.AverageSentiment)

				locationData.AvgSentimentList = append(locationData.AvgSentimentList, newSentiment)
				e := db.AddNewLocSentimentAvg(firestoreClient, testLocationId, locationData.AvgSentimentList)
				if e != nil {
					log.Printf("Failed to add new avgLocation.", e)
				}
				log.Printf("Added new average to list")

			} else {
				log.Printf("Updating timestamp of latest sentiment")
				e := db.UpdateLocSentimentTimestampWithData(firestoreClient, locationData, testLocationId, end)
				if e != nil {
					log.Printf("Failed to update avgLocation field")
				}
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
