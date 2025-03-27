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
		// Limit(5).                            // fuck it, we ball
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
func GetValidLocation(client *firestore.Client, locationDocID string) (types.LocationData, error) {
	ctx := context.Background()
	var locationData types.LocationData

	doc, err := client.Collection("locations").Doc(locationDocID).Get(ctx)
	if err != nil {
		return locationData, err
	}

	if err := doc.DataTo(&locationData); err != nil {
		return locationData, err
	}
	return locationData, nil
}

// since its the first one, use set
func InitLocationSentiment(client *firestore.Client, locationDocID string, newAvgSentiment types.AvgLocationSentiment) error {
	ctx := context.Background()
	locDoc := client.Collection("locations").Doc(locationDocID)
	data := map[string]interface{}{
		"avgSentimentList": []types.AvgLocationSentiment{newAvgSentiment},
	}
	_, err := locDoc.Set(ctx, data, firestore.MergeAll)
	return err
}

// update the most recent object in the avgSentimentList to indicate that no new skeets were needed to be processed
func UpdateLocSentimentTimestampWithData(client *firestore.Client, locationData types.LocationData, locationDocID, newTime string) error {
	// Ensure there's at least one sentiment record.
	if len(locationData.AvgSentimentList) == 0 {
		return fmt.Errorf("AvgSentimentList is empty, cannot update timestamp")
	}

	// Update the timestamp of the most recent sentiment record.
	lastIndex := len(locationData.AvgSentimentList) - 1
	locationData.AvgSentimentList[lastIndex].TimeStamp = newTime

	// Prepare update data.
	updateData := map[string]interface{}{
		"avgSentimentList": locationData.AvgSentimentList,
	}

	ctx := context.Background()
	locDocRef := client.Collection("locations").Doc(locationDocID)
	_, err := locDocRef.Set(ctx, updateData, firestore.MergeAll)
	return err
}

func AddNewLocSentimentAvg(client *firestore.Client, locationDocID string, updatedList []types.AvgLocationSentiment) error {
	ctx := context.Background()
	locDoc := client.Collection("locations").Doc(locationDocID)

	unionItems := make([]interface{}, len(updatedList))
	for i, v := range updatedList {
		unionItems[i] = v
	}

	// could have used set since I am just gettingt them all and updating locally, but idk man that scares me
	_, err := locDoc.Set(ctx, map[string]interface{}{
		"avgSentimentList": firestore.ArrayUnion(unionItems...),
	}, firestore.MergeAll)

	return err
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

func GetLatestSkeetInSubCollection(client *firestore.Client, locationDocID string) (types.SkeetSubDoc, error) {
	ctx := context.Background()
	var latestSkeet types.SkeetSubDoc

	docs, err := client.Collection("locations").
		Doc(locationDocID).
		Collection("skeetIds").
		OrderBy("skeetData.timestamp", firestore.Desc). // latest first
		Limit(1).Documents(ctx).GetAll()
	if err != nil {
		return latestSkeet, fmt.Errorf("error executing query: %w", err)
	}

	if len(docs) == 0 {
		return latestSkeet, fmt.Errorf("no skeets found in subcollection for location %s", locationDocID)
	}
	if err := docs[0].DataTo(&latestSkeet); err != nil {
		return latestSkeet, fmt.Errorf("error converting document to SkeetSubDoc: %w", err)
	}
	return latestSkeet, nil
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

func UpdateLocationDoc(client *firestore.Client, locationID string, locationData types.LocationData) error {
	ctx := context.Background()
	locDoc := client.Collection("locations").Doc(locationID)
	_, err := locDoc.Set(ctx, locationData, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("failed to update location document %s: %w", locationID, err)
	}
	return nil
}

// UpdateLocationSentimentList replaces the entire avgSentimentList field with the provided list.
func UpdateLocationSentimentList(client *firestore.Client, locationID string, updatedList []types.AvgLocationSentiment) error {
	ctx := context.Background()
	locDocRef := client.Collection("locations").Doc(locationID)

	updates := []firestore.Update{
		{Path: "avgSentimentList", Value: updatedList},
	}

	_, err := locDocRef.Update(ctx, updates)
	if err != nil {
		return fmt.Errorf("failed to update avgSentimentList for %s: %w", locationID, err)
	}
	return nil
}

// UpdateLocationFields updates specific top-level fields using a map.
func UpdateLocationFields(client *firestore.Client, locationID string, fieldsToUpdate map[string]interface{}) error {
	ctx := context.Background()
	locDocRef := client.Collection("locations").Doc(locationID)

	_, err := locDocRef.Set(ctx, fieldsToUpdate, firestore.MergeAll)
	if err != nil {
		return fmt.Errorf("failed to update fields for location %s: %w", locationID, err)
	}
	return nil
}

func GetTopLocationsBySkeetAmount(client *firestore.Client, limit int) ([]types.LocationData, error) {
	ctx := context.Background()
	var topLocations []types.LocationData

	query := client.Collection("locations").
		Where("formattedAddress", "!=", "").
		OrderBy("latestSkeetsAmount", firestore.Desc).
		Limit(limit)

	iter := query.Documents(ctx)
	defer iter.Stop() // Ensure iterator resources are released

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break // End of results
		}
		if err != nil {
			return nil, fmt.Errorf("error iterating top locations: %w", err)
		}

		var location types.LocationData
		if err := doc.DataTo(&location); err != nil {
			return nil, fmt.Errorf("error converting document %s to LocationData: %w", doc.Ref.ID, err)
		}
		topLocations = append(topLocations, location)
	}

	return topLocations, nil
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
