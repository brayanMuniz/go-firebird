package db

import (
	"cloud.google.com/go/firestore"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
