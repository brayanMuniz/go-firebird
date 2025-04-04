package db

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"go-firebird/types"
	"google.golang.org/api/iterator"
	"log"
)

const disastersCollection = "disasters"

// SaveDisasters saves a slice of DisasterData objects to the 'disasters' collection
// using BulkWriter for efficient non-transactional writes.
// It uses the DisasterData.ID field as the Firestore document ID.
func SaveDisasters(client *firestore.Client, disasters []types.DisasterData) error {
	if len(disasters) == 0 {
		log.Println("No disasters to save.")
		return nil
	}

	ctx := context.Background() // Use background context for the bulk write operation
	bw := client.BulkWriter(ctx)
	disastersCollectionRef := client.Collection(disastersCollection)

	log.Printf("Preparing to save %d disasters using BulkWriter to collection '%s'...", len(disasters), disastersCollection)

	savedCount := 0
	for i := range disasters {
		disaster := disasters[i]

		if disaster.ID == "" {
			log.Printf("Warning: Skipping disaster with empty ID: %+v", disaster)
			continue // Cannot save without an ID
		}
		// Use the pre-generated disaster.ID as the document ID
		docRef := disastersCollectionRef.Doc(disaster.ID)

		_, err := bw.Set(docRef, disaster)
		if err != nil {
			log.Printf("Error enqueueing disaster %s for save: %v", disaster.ID, err)
		} else {
			savedCount++
		}
	}

	if savedCount == 0 {
		log.Println("No valid disasters were enqueued for saving.")
		return nil
	}

	// NOTE: Flush sends any remaining writes and waits for them to complete.
	// It should be called before the BulkWriter goes out of scope.
	log.Printf("Flushing BulkWriter for %d enqueued disaster saves...", savedCount)
	bw.Flush() // Flush ensures all writes are sent

	log.Printf("BulkWriter flushed. Attempted to save %d disasters.", savedCount)

	return nil
}

// GetAllDisasters retrieves all documents from the 'disasters' collection.
func GetAllDisasters(client *firestore.Client) ([]types.DisasterData, error) {
	ctx := context.Background()
	var allDisasters []types.DisasterData

	iter := client.Collection(disastersCollection).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error iterating disasters collection: %w", err)
		}

		var disaster types.DisasterData
		if err := doc.DataTo(&disaster); err != nil {
			log.Printf("Warning: Error converting document %s to DisasterData: %v. Skipping.", doc.Ref.ID, err)
			continue
		}
		disaster.ID = doc.Ref.ID
		allDisasters = append(allDisasters, disaster)
	}
	log.Printf("Retrieved %d disasters from the database.", len(allDisasters))
	return allDisasters, nil
}

// GetDisasterByID retrieves a single disaster document by its ID.
func GetDisasterByID(client *firestore.Client, disasterID string) (types.DisasterData, error) {
	ctx := context.Background()
	var disaster types.DisasterData

	docSnap, err := client.Collection(disastersCollection).Doc(disasterID).Get(ctx)
	if err != nil {
		return disaster, fmt.Errorf("error getting disaster %s: %w", disasterID, err)
	}

	if err := docSnap.DataTo(&disaster); err != nil {
		return disaster, fmt.Errorf("error converting document %s to DisasterData: %w", disasterID, err)
	}
	disaster.ID = docSnap.Ref.ID

	return disaster, nil
}
