package db

import (
	"cloud.google.com/go/firestore"
	"context"
	"go-firebird/geocode"
	"log"
)

// TODO: update this struct in order to have a sentiment
type LocationData struct {
	LocationName     string  `firestore:"locationName"`
	FormattedAddress string  `firestore:"formattedAddress"`
	Lat              float64 `firestore:"lat"`
	Long             float64 `firestore:"long"`
	Type             string  `firestore:"type"`
}

type NewLocationMetaData struct {
	LocationName string
	Type         string
	NewLocation  bool
}

// locations who have been able to be geocoded
func GetValidLocations(client *firestore.Client) ([]LocationData, error) {
	var validLocations []LocationData
	ctx := context.Background()

	// Query all valid documents
	docs, err := client.Collection("locations").
		Where("formattedAddress", "!=", ""). // processed to be invalid
		Documents(ctx).
		GetAll()
	if err != nil {
		return nil, err
	}

	// Convert each doc to LocationData struct
	for _, doc := range docs {
		var location LocationData
		if err := doc.DataTo(&location); err != nil {
			return nil, err
		}
		validLocations = append(validLocations, location)
	}

	return validLocations, nil

}

func GetNewLocations(client *firestore.Client) ([]LocationData, error) {
	var newLocations []LocationData
	ctx := context.Background()

	// Query all documents where newlocation == true
	docs, err := client.Collection("locations").
		Where("newLocation", "==", true).
		Documents(ctx).
		GetAll()
	if err != nil {
		return nil, err
	}

	// Convert each doc to LocationData struct
	for _, doc := range docs {
		var location LocationData
		if err := doc.DataTo(&location); err != nil {
			return nil, err
		}
		newLocations = append(newLocations, location)
	}

	return newLocations, nil
}

// uses newLocation flag to determine if an update is needed
func UpdateLocationGeocoding(client *firestore.Client, locationName string) {
	hashedLocationID := HashString(locationName)
	results, err := geocode.GeocodeAddress(locationName)
	if err != nil {
		log.Printf("Failed to geocode %s: %v", locationName, err)
		return
	}

	result := ""
	if len(results) > 0 {
		result = results[0].FormattedAddress

	}
	geoData := map[string]interface{}{
		"formattedAddress": result,
		"lat":              0,
		"long":             0,
		"newLocation":      false,
	}

	if len(results) == 0 {
		log.Printf("No geocode results for %s", locationName)
		log.Println("Setting result to null")
	} else {
		loc := results[0].Geometry.Location
		geoData = map[string]interface{}{
			"formattedAddress": results[0].FormattedAddress,
			"lat":              loc.Lat,
			"long":             loc.Lng,
			"newLocation":      false,
		}
	}

	ctx := context.Background()
	_, err = client.Collection("locations").Doc(hashedLocationID).Set(ctx, geoData, firestore.MergeAll)
	if err != nil {
		log.Printf("\nFailed to update geocoding data for %s: %v", locationName, err)
		return
	}
	log.Printf("\nSuccessfully updated geocoding data for %s", locationName)
}
