package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/joho/godotenv"
	"go-firebird/cronjobs"
	"go-firebird/db"
	"go-firebird/routes"
	"io"
	"log"
	"os"

	language "cloud.google.com/go/language/apiv2"
	"cloud.google.com/go/language/apiv2/languagepb"
	"google.golang.org/api/option"
)

// analyzeEntities sends text to the Cloud Natural Language API to extract named entities.
func analyzeEntities(w io.Writer, client *language.Client, text string) error {
	ctx := context.Background()
	req := &languagepb.AnalyzeEntitiesRequest{
		Document: &languagepb.Document{
			Source: &languagepb.Document_Content{
				Content: text,
			},
			Type: languagepb.Document_PLAIN_TEXT,
		},
		EncodingType: languagepb.EncodingType_UTF8,
	}
	resp, err := client.AnalyzeEntities(ctx, req)
	if err != nil {
		return fmt.Errorf("AnalyzeEntities error: %w", err)
	}
	fmt.Fprintf(w, "Entities: %v\n", resp.Entities)
	return nil
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

	// Init firestore
	firestoreClient, err := db.InitFirestore()
	if err != nil {
		log.Fatalf("Failed to initialize Firestore: %v", err)
	}
	defer db.CloseFirestore() // Firestore client is closed on exit

	// Initialize cron jobs
	cronjobs.InitCronJobs()

	// Decode credentials
	encodedCreds := os.Getenv("NATURAL_LANGUAGE_CREDENTIALS")
	naturalLangCred, err := base64.StdEncoding.DecodeString(encodedCreds)
	if err != nil {
		log.Fatalf("Failed to decode Natural language credentials credentials: %v", err)
	}

	// Create the Natural Language API client using the decoded credentials
	opt := option.WithCredentialsJSON(naturalLangCred)
	langClient, err := language.NewClient(context.Background(), opt)
	if err != nil {
		log.Fatalf("Failed to create Natural Language client: %v", err)
	}
	defer langClient.Close()

	fmt.Println("\nAnalyzing Entities...")
	if err := analyzeEntities(os.Stdout, langClient, "The address of the White House is 1600 Pennsylvania Avenue NW, Washington, D.C. 20500, United States."); err != nil {
		log.Fatalf("Error analyzing entities: %v", err)
	}

	r := routes.SetupRouter(firestoreClient)
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
