package handlers

import (
	"cloud.google.com/go/firestore"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-firebird/db"
	"log"
	"net/http"
	"sync"
)

func TestUpdateGeocodingDBTest(c *gin.Context, firestoreClient *firestore.Client) {
	newLocations, err := db.GetNewLocations(firestoreClient)
	if err != nil {
		log.Printf("Error fetching new locations: %v", err)
		return
	}
	for v, x := range newLocations {
		fmt.Print(v, x)
	}

	var wg sync.WaitGroup
	for v, loc := range newLocations {
		println(v)
		wg.Add(1)
		go func(location db.LocationData) {
			defer wg.Done()
			log.Printf("Updating geocode for location hash: %s", location.LocationName)
			db.UpdateLocationGeocoding(firestoreClient, location.LocationName)
		}(loc)
	}
	wg.Wait() // Wait for all updates to finish

	mock := map[string]interface{}{
		"good": 0,
	}

	// Return JSON
	c.JSON(http.StatusOK, mock)
}
