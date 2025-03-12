package handlers

import (
	"fmt"
	"go-firebird/geocode"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func TestGeocode(c *gin.Context) {
	locationParam := c.Query("location")
	if locationParam == "" {
		locationParam = "Colorado"
	}

	// JSON response struct
	type LocationResponse struct {
		Location  string  `json:"location"`
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
	}

	responseData := LocationResponse{
		Location: locationParam,
	}

	results, err := geocode.GeocodeAddress(locationParam)
	if err != nil {
		log.Fatalf("Error geocoding address: %v", err)
	}
	if len(results) == 0 {
		log.Println("No geocoding results found")
	} else {
		// Print the formatted address and its latitude and longitude
		location := results[0].Geometry.Location
		fmt.Printf("Geocoding result for 'Colorado': %s\nLatitude: %f, Longitude: %f\n",
			results[0].FormattedAddress, location.Lat, location.Lng)
		responseData.Longitude = location.Lng
		responseData.Latitude = location.Lat
	}

	// Return JSON response with mock data and entities.
	c.JSON(http.StatusOK, responseData)
}
