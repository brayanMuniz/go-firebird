package db

import (
	"cloud.google.com/go/firestore"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	firebase "firebase.google.com/go"
	"fmt"
	"go-firebird/geocode"
	"go-firebird/nlp"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"os"
	"sync"
)

// hashString hashes a given string using SHA-256 and returns its hex representation.
func HashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// FirestoreClient is a singleton Firestore client instance.
var (
	client     *firestore.Client
	clientOnce sync.Once
)

// InitFirestore initializes and returns a Firestore client.
func InitFirestore() (*firestore.Client, error) {
	var err error

	clientOnce.Do(func() {
		// Decode credentials
		encodedCreds := os.Getenv("FIREBASE_CREDENTIALS")
		creds, err := base64.StdEncoding.DecodeString(encodedCreds)
		if err != nil {
			log.Fatalf("Failed to decode Firestore credentials: %v", err)
		}

		// Initialize Firebase App
		opt := option.WithCredentialsJSON(creds)
		app, err := firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			log.Fatalf("Error initializing Firestore: %v", err)
		}

		// Get Firestore Client
		client, err = app.Firestore(context.Background())
		if err != nil {
			log.Fatalf("Error getting Firestore client: %v", err)
		}
	})

	return client, err
}

// CloseFirestore closes the Firestore client.
func CloseFirestore() {
	if client != nil {
		client.Close()
	}
}

// Skeet represents a post stored in Firestore
type Skeet struct {
	Avatar      string `firestore:"avatar"`
	Content     string `firestore:"content"`
	Timestamp   string `firestore:"timestamp"`
	Handle      string `firestore:"handle"`
	DisplayName string `firestore:"displayName"`
	UID         string `firestore:"uid"`
}

// This was made out of necessity because of how long the parameters would have been lol
type SaveCompleteSkeetType struct {
	NewSkeet       Skeet
	Classification []float64
	Entities       []nlp.Entity
	Sentiment      nlp.Sentiment
}

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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Returns new location names
func SaveCompleteSkeet(client *firestore.Client, data SaveCompleteSkeetType) ([]string, error) {
	ctx := context.Background()

	// Build arrays for ADDRESS and LOCATION entities.
	var addresses, locations []nlp.Entity
	for _, entity := range data.Entities {
		if entity.Type == "ADDRESS" {
			addresses = append(addresses, entity)
		} else if entity.Type == "LOCATION" {
			locations = append(locations, entity)
		}
	}

	// hashedLocation Id's that have "invalid" location data
	invalidLocations := []string{}

	// Prepare main skeet data.
	skeetData := map[string]interface{}{
		"avatar":          data.NewSkeet.Avatar,
		"content":         data.NewSkeet.Content,
		"timestamp":       data.NewSkeet.Timestamp,
		"handle":          data.NewSkeet.Handle,
		"displayName":     data.NewSkeet.DisplayName,
		"uid":             data.NewSkeet.UID,
		"classification":  data.Classification,
		"entity_ADDRESS":  addresses,
		"entity_LOCATION": locations,
		"sentiment":       data.Sentiment,
	}

	hashedSkeetID := HashString(data.NewSkeet.UID)

	fmt.Printf("\nRunning transaction for: %s, hashed: %s\n", data.NewSkeet.UID, hashedSkeetID)

	// Run a transaction to perform all writes atomically.
	newLocationsData := []NewLocationMetaData{}
	err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {

		// For each entity of type LOCATION or ADDRESS, check if its new and add it to newLocationsData
		for _, entity := range data.Entities {
			if entity.Type == "LOCATION" || entity.Type == "ADDRESS" {

				// Retrieve the current value of "newLocation"
				hashedLocationID := HashString(entity.Name)
				locationDocRef := client.Collection("locations").Doc(hashedLocationID)
				locationDoc, err := tx.Get(locationDocRef)
				if err != nil {
					if status.Code(err) == codes.NotFound {
						// If the document doesn't exist, create it and set newLocation to true.
						newLocationsData = append(newLocationsData, NewLocationMetaData{
							LocationName: entity.Name,
							Type:         entity.Type,
							NewLocation:  true,
						})

					} else {
						return fmt.Errorf("error getting location doc for %s: %w", entity.Name, err)
					}

				} else {
					// Extract document data and check its geocoding fields to see if its invalid

					// NOTE: if it is "invalid" the value is 0, it will return an err, so set to default
					dataMap := locationDoc.Data()
					lat, latOk := dataMap["lat"].(float64)
					if !latOk {
						lat = 0
					}
					long, longOk := dataMap["long"].(float64)
					if !longOk {
						long = 0
					}
					newLocation, nlOk := dataMap["newLocation"].(bool)
					// If the document is validly structured, check if it's invalid.
					if nlOk {
						if lat == 0 && long == 0 && newLocation == false {
							// Mark as invalid by skipping addition.
							invalidLocations = append(invalidLocations, hashedLocationID)
						}
					}

				}

			}

		}

		// Set the main skeet document.
		skeetDocRef := client.Collection("skeets").Doc(hashedSkeetID)
		if err := tx.Set(skeetDocRef, skeetData, firestore.MergeAll); err != nil {
			return fmt.Errorf("failed to set skeet document: %w", err)
		}

		// create the new location docs
		for _, value := range newLocationsData {
			hashedLocationID := HashString(value.LocationName)
			locationDocRef := client.Collection("locations").Doc(hashedLocationID)

			// Convert struct to map.
			locationDataMap := map[string]interface{}{
				"locationName": value.LocationName,
				"type":         value.Type,
				"newLocation":  value.NewLocation,
			}

			if err := tx.Set(locationDocRef, locationDataMap, firestore.MergeAll); err != nil {
				return fmt.Errorf("failed to set location doc for %s: %w", value.LocationName, err)
			}

		}

		// loop over again and set all the skeet data in /locations/location/skeets
		for _, entity := range data.Entities {
			hashedLocationID := HashString(entity.Name)
			if entity.Type == "LOCATION" || entity.Type == "ADDRESS" {
				if contains(invalidLocations, hashedLocationID) {
					fmt.Printf("Skipping skeet subcollection write for invalid location: %s\n", entity.Name)
					continue
				}

				locationDocRef := client.Collection("locations").Doc(hashedLocationID)

				// In the subcollection location/locId/skeetIds, store the skeet data with entity details.
				subDocRef := locationDocRef.Collection("skeetIds").Doc(hashedSkeetID)
				subData := map[string]interface{}{
					"mentions":     entity.Mentions,
					"locationName": entity.Name,
					"type":         entity.Type,
					"skeetData":    skeetData,
				}
				if err := tx.Set(subDocRef, subData, firestore.MergeAll); err != nil {
					return fmt.Errorf("failed to set subcollection for %s: %w", entity.Name, err)
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Printf("Transaction failed: %v", err)
		return nil, err
	}

	// Extract the new location names to return.
	var newLocationNames []string
	for _, locMeta := range newLocationsData {
		newLocationNames = append(newLocationNames, locMeta.LocationName)
	}

	fmt.Printf("\nSuccessfully saved complete skeet with hashed ID: %s\n", hashedSkeetID)
	return newLocationNames, nil
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
		var ld LocationData
		if err := doc.DataTo(&ld); err != nil {
			return nil, err
		}
		newLocations = append(newLocations, ld)
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

// ReadSkeets retrieves and prints all skeets from Firestore.
func ReadSkeets(client *firestore.Client) {
	ctx := context.Background()
	iter := client.Collection("skeets").Documents(ctx) // Get all documents

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error iterating documents: %v", err)
		}

		var skeet Skeet
		if err := doc.DataTo(&skeet); err != nil {
			log.Fatalf("Error converting to struct: %v", err)
		}
		fmt.Printf("Read Skeet: %+v\n", skeet)
	}
}

// WriteSkeet adds a new skeet to Firestore.
func WriteSkeet(client *firestore.Client, newSkeet Skeet) {
	ctx := context.Background()
	hashedSkeetID := HashString(newSkeet.UID)
	_, err := client.Collection("skeets").Doc(hashedSkeetID).Set(ctx, newSkeet)
	if err != nil {
		log.Fatalf("Error adding document: %v", err)
	}
	fmt.Printf("Added document with hashed ID: %s\n", hashedSkeetID)
}

// DeleteSkeet removes a skeet from Firestore using its document ID.
func DeleteSkeet(client *firestore.Client, skeetID string) (*firestore.WriteResult, error) {
	ctx := context.Background()
	hashedSkeetID := HashString(skeetID)
	docRef := client.Collection("skeets").Doc(hashedSkeetID)
	writeResult, err := docRef.Delete(ctx)

	if err != nil {
		log.Fatalf("Error deleting document: %v", err)
		return nil, err
	}

	fmt.Printf("Skeet with hashed ID '%s' deleted successfully.\n", hashedSkeetID)
	return writeResult, nil
}

// UpdateSkeetContent updates the content of a skeet in Firestore.
func UpdateSkeetContent(client *firestore.Client, skeetID, newContent string) (*firestore.WriteResult, error) {
	ctx := context.Background()
	hashedSkeetID := HashString(skeetID)
	docRef := client.Collection("skeets").Doc(hashedSkeetID)
	updates := []firestore.Update{
		{Path: "content", Value: newContent},
	}

	// Perform the update
	res, err := docRef.Update(ctx, updates)
	if err != nil {
		return nil, fmt.Errorf("error updating document: %w", err)
	}

	fmt.Printf("Skeet with ID '%s' updated successfully.\n", hashedSkeetID)
	return res, nil
}
