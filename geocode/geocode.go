package geocode

import (
	"context"
	"fmt"
	"googlemaps.github.io/maps"
	"log"
	"os"
	"sync"
)

// mapsClient is a singleton maps client instance.
var (
	mapsClient *maps.Client
	clientOnce sync.Once
)

// InitMapsClient initializes and returns a singleton Google Maps client.
func InitMapsClient() (*maps.Client, error) {
	var err error
	clientOnce.Do(func() {
		apiKey := os.Getenv("MAPS_CREDENTIALS")
		if apiKey == "" {
			err = fmt.Errorf("MAPS_CREDENTIALS environment variable not set")
			return
		}
		mapsClient, err = maps.NewClient(maps.WithAPIKey(apiKey))
		if err != nil {
			log.Fatalf("Failed to create maps client: %v", err)
		}
	})
	return mapsClient, err
}

// GeocodeAddress takes an address string and returns geocoding results.
func GeocodeAddress(address string) ([]maps.GeocodingResult, error) {
	client, err := InitMapsClient()
	if err != nil {
		return nil, err
	}

	req := &maps.GeocodingRequest{
		Address: address,
	}

	// Forward geocode: get latitude and longitude for the given address.
	results, err := client.Geocode(context.Background(), req)
	if err != nil {
		return nil, err
	}

	return results, nil
}
