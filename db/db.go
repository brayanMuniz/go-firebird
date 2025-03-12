package db

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"go-firebird/nlp"
	"log"
	"os"
	"sync"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
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

func SaveCompleteSkeet(client *firestore.Client, data SaveCompleteSkeetType) error {
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

	// Compute the hashed skeet ID.
	hashedSkeetID := HashString(data.NewSkeet.UID)

	fmt.Println("Running transaction for: ", hashedSkeetID)

	// Run a transaction to perform all writes atomically.
	err := client.RunTransaction(ctx, func(ctx context.Context, tx *firestore.Transaction) error {

		// Set the main skeet document.
		skeetDocRef := client.Collection("skeets").Doc(hashedSkeetID)
		if err := tx.Set(skeetDocRef, skeetData, firestore.MergeAll); err != nil {
			return fmt.Errorf("failed to set skeet document: %w", err)
		}

		// For each entity of type LOCATION or ADDRESS, update its location doc and nested subcollection.
		for _, entity := range data.Entities {
			if entity.Type == "LOCATION" || entity.Type == "ADDRESS" {
				hashedLocationID := HashString(entity.Name)
				locationMetadata := map[string]interface{}{
					"locationName": entity.Name,
					"type":         entity.Type,
				}
				locationDocRef := client.Collection("locations").Doc(hashedLocationID)
				if err := tx.Set(locationDocRef, locationMetadata, firestore.MergeAll); err != nil {
					return fmt.Errorf("failed to set location doc for %s: %w", entity.Name, err)
				}

				// In the subcollection "skeetIds", store the skeet data with entity details.
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
		return err
	}

	fmt.Printf("Successfully saved complete skeet with hashed ID: %s\n", hashedSkeetID)
	return nil
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
