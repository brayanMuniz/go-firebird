package db

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"go-firebird/geocode"
	"go-firebird/types"
	"google.golang.org/api/iterator"
	"log"
)

// locations who have been able to be geocoded
func GetValidLocations(client *firestore.Client) ([]types.LocationData, error) {
	var validLocations []types.LocationData
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
		var location types.LocationData
		if err := doc.DataTo(&location); err != nil {
			return nil, err
		}
		validLocations = append(validLocations, location)
	}

	return validLocations, nil

}

// single valid location based on id
func GetValidLocation(client *firestore.Client, locId string) (types.LocationData, error) {
	ctx := context.Background()
	var locationData types.LocationData

	doc, err := client.Collection("locations").Doc(locId).Get(ctx)
	if err != nil {
		return locationData, err
	}

	if err := doc.DataTo(&locationData); err != nil {
		return locationData, err
	}
	return locationData, nil
}

func GetSkeetsSubCollection(client *firestore.Client, locationDocID string, start, end string) ([]types.SkeetSubDoc, error) {
	ctx := context.Background()
	var skeets []types.SkeetSubDoc

	iter := client.Collection("locations").
		Doc(locationDocID).
		Collection("skeetIds").
		Where("skeetData.timestamp", ">=", start).
		Where("skeetData.timestamp", "<=", end).
		Documents(ctx)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error iterating skeets: %w", err)
		}
		var s types.SkeetSubDoc
		if err := doc.DataTo(&s); err != nil {
			return nil, fmt.Errorf("error converting document to SubSkeet: %w", err)
		}
		skeets = append(skeets, s)
	}

	return skeets, nil

}

func GetNewLocations(client *firestore.Client) ([]types.LocationData, error) {
	var newLocations []types.LocationData
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
		var location types.LocationData
		if err := doc.DataTo(&location); err != nil {
			return nil, err
		}
		newLocations = append(newLocations, location)
	}

	return newLocations, nil
}

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
		"newLocation":      false, // uses newLocation flag to determine if an update is needed
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
