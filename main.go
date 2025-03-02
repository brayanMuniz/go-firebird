package main

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go"
	"fmt"
	"github.com/joho/godotenv"
	"go-firebird/cronjobs"
	"go-firebird/routes"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"log"
	"os"
)

// Structure to represent your skeet data
type Skeet struct {
	Content string `firestore:"content"`
	Author  string `firestore:"author"`
}

func readSkeets(client *firestore.Client) {
	ctx := context.Background()
	iter := client.Collection("skeets").Documents(ctx) //get all documents
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

func writeSkeet(client *firestore.Client) {
	ctx := context.Background()
	newSkeet := &Skeet{
		Author:  "kanye1",
		Content: "Got banned for NO Reason",
	}

	ref, _, err := client.Collection("skeets").Add(ctx, newSkeet) //autogenerate ID

	if err != nil {
		log.Fatalf("Error adding document: %v", err)
	}
	fmt.Printf("Added document with ID: %s\n", ref.ID)
}

func deleteSkeet(client *firestore.Client, skeetID string) (*firestore.WriteResult, error) {
	ctx := context.Background()
	docRef := client.Collection("skeets").Doc(skeetID)
	writeResult, err := docRef.Delete(ctx)

	if err != nil {
		log.Fatalf("Error deleting document: %v", err)
	}

	fmt.Printf("Skeet with ID '%s' deleted successfully.\n", skeetID)

	return writeResult, nil

}

func updateSkeetContent(client *firestore.Client, skeetID string) (*firestore.WriteResult, error) {
	ctx := context.Background()
	docRef := client.Collection("skeets").Doc(skeetID)
	updates := []firestore.Update{
		{Path: "content", Value: ""}, // Update the content field to an empty string
	}

	// Perform the update
	res, err := docRef.Update(ctx, updates)
	if err != nil {
		return nil, fmt.Errorf("error updating document: %w", err)
	}

	fmt.Printf("Skeet with ID '%s' updated successfully.\n", skeetID)

	return res, nil

}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Print and check env
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey != "" {
		fmt.Println("OPENAI_API_KEY loaded")

	}

	clientURL := os.Getenv("CLIENT_URL")
	fmt.Println("CLIENT_URL: ", clientURL)

	// Initialize cron jobs
	cronjobs.InitCronJobs()

	// firebase
	opt := option.WithCredentialsFile("./firebird-firebase.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		fmt.Errorf("error initializing app: %v", err)
	}

	// firestore
	client, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatalf("error getting firestore client: %v", err)
	}
	defer client.Close()

	r := routes.SetupRouter()
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
