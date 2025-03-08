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
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// hashString hashes a given string using SHA-256 and returns its hex representation.
func hashString(s string) string {
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
	Author  string `firestore:"author"`
	Content string `firestore:"content"`
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
func WriteSkeet(client *firestore.Client, author, content string) {
	ctx := context.Background()
	newSkeet := &Skeet{
		Author:  author,
		Content: content,
	}

	ref, _, err := client.Collection("skeets").Add(ctx, newSkeet) // Auto-generate ID
	if err != nil {
		log.Fatalf("Error adding document: %v", err)
	}
	fmt.Printf("Added document with ID: %s\n", ref.ID)
}

// DeleteSkeet removes a skeet from Firestore using its document ID.
func DeleteSkeet(client *firestore.Client, skeetID string) (*firestore.WriteResult, error) {
	ctx := context.Background()
	docRef := client.Collection("skeets").Doc(skeetID)
	writeResult, err := docRef.Delete(ctx)

	if err != nil {
		log.Fatalf("Error deleting document: %v", err)
		return nil, err
	}

	fmt.Printf("Skeet with ID '%s' deleted successfully.\n", skeetID)
	return writeResult, nil
}

// UpdateSkeetContent updates the content of a skeet in Firestore.
func UpdateSkeetContent(client *firestore.Client, skeetID, newContent string) (*firestore.WriteResult, error) {
	ctx := context.Background()
	docRef := client.Collection("skeets").Doc(skeetID)
	updates := []firestore.Update{
		{Path: "content", Value: newContent}, // Update content field
	}

	// Perform the update
	res, err := docRef.Update(ctx, updates)
	if err != nil {
		return nil, fmt.Errorf("error updating document: %w", err)
	}

	fmt.Printf("Skeet with ID '%s' updated successfully.\n", skeetID)
	return res, nil
}

// SaveLocationEntity saves or updates a location entity in Firestore using a nested structure.
// The root document in "locations" uses a hashed location name as its ID and stores the location metadata.
// Under that, the subcollection "skeetIds" will have a document with a hashed skeet ID, storing the entity data.
func SaveLocationEntity(client *firestore.Client, locationName, skeetID string, entityData nlp.Entity) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Hash the location name and skeet ID to not have to deal with ill formed document names
	hashedLocationID := hashString(locationName)
	hashedSkeetID := hashString(skeetID)

	locationMetadata := map[string]interface{}{
		"name": entityData.Name,
		"type": entityData.Type,
	}

	// Update the root document for location metadata.
	_, err := client.Collection("locations").Doc(hashedLocationID).Set(ctx, locationMetadata, firestore.MergeAll)
	if err != nil {
		log.Printf("Failed to save location metadata for '%s' (hashed: %s): %v", locationName, hashedLocationID, err)
		return err
	}

	// Prepare data for the skeet document.
	subData := map[string]interface{}{
		"skeetID":   skeetID,
		"Mentions":  entityData.Mentions,
		"Name":      entityData.Name,
		"Type":      entityData.Type,
		"timestamp": firestore.ServerTimestamp,
	}

	// Save or update the skeet document under the location's subcollection.
	_, err = client.Collection("locations").Doc(hashedLocationID).Collection("skeetIds").Doc(hashedSkeetID).Set(ctx, subData, firestore.MergeAll)
	if err != nil {
		log.Printf("Failed to save skeet data for '%s' (hashed: %s) under location '%s': %v", skeetID, hashedSkeetID, hashedLocationID, err)
		return err
	}

	return nil
}
