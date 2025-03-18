package handlers

import (
	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"go-firebird/db"
	"go-firebird/processor"
	"go-firebird/types"
	"log"
	"net/http"
	"strings"
	"sync"
)

func TestLocationSentimentUpdate(c *gin.Context, firestoreClient *firestore.Client) {
	goodSaving := make([]string, 0)
	failureSaving := make([]string, 0)

	testLocationId := strings.TrimSpace(c.Query("docId"))
	if testLocationId != "" {
		err := processor.ProcessLocationAvgSentiment(firestoreClient, testLocationId)
		if err != nil {
			log.Printf("Error processing the location average sentiment save: %v", err)
			failureSaving = append(failureSaving, testLocationId)
		} else {
			goodSaving = append(goodSaving, testLocationId)
		}
	} else {

		// get all the valid locations
		validLocations, err := db.GetValidLocations(firestoreClient)
		if err != nil {
			log.Printf("Error fetching valid locations: %v", err)
			return
		}

		var wg sync.WaitGroup
		var mu sync.Mutex // To protect shared state updates.
		// My little variables cant be this cute!

		for _, locData := range validLocations {
			wg.Add(1)
			go func(location types.LocationData) {
				defer wg.Done()

				docId := db.HashString(locData.LocationName)
				if err := processor.ProcessLocationAvgSentiment(firestoreClient, docId); err != nil {
					log.Printf("Error processing the location average sentiment save: %v", err)
					mu.Lock()
					failureSaving = append(failureSaving, docId)
					mu.Unlock()
				} else {
					mu.Lock()
					goodSaving = append(goodSaving, docId)
					mu.Unlock()

				}
			}(locData)
		}

		wg.Wait()

	}

	mock := map[string]interface{}{
		"goodSaving":    goodSaving,
		"failureSaving": failureSaving,
	}

	// Return JSON
	c.JSON(http.StatusOK, mock)
}
