package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"go-firebird/cronjobs"
	"go-firebird/db"
	"go-firebird/nlp"
	"go-firebird/routes"
	"log"
	"os"

	"googlemaps.github.io/maps"
)

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

	mapsCredentials := os.Getenv("MAPS_CREDENTIALS")

	// Create a new Google Maps client
	mapsClient, err := maps.NewClient(maps.WithAPIKey(mapsCredentials))
	if err != nil {
		log.Fatalf("fatal error creating maps client: %s", err)
	}

	// Build a GeocodingRequest with the desired address.
	geocodeReq := &maps.GeocodingRequest{
		Address: "Colorado",
	}

	// Call Geocode to perform forward geocoding.
	geocodeResp, err := mapsClient.Geocode(context.Background(), geocodeReq)
	if err != nil {
		log.Fatalf("fatal error calling Geocode: %s", err)
	}

	// Check that we have at least one result.
	if len(geocodeResp) == 0 {
		log.Println("No results found")
		return
	}

	// Get the location from the first result.
	location := geocodeResp[0].Geometry.Location
	fmt.Printf("Address: %s\nLatitude: %f, Longitude: %f\n", geocodeResp[0].FormattedAddress, location.Lat, location.Lng)

	// Init firestore
	firestoreClient, err := db.InitFirestore()
	if err != nil {
		log.Fatalf("Failed to initialize Firestore: %v", err)
	}
	defer db.CloseFirestore() // Firestore client is closed on exit

	// Init language client
	languageClient, err := nlp.InitLanguageClient()
	if err != nil {
		log.Fatalf("Failed to initialize nlp: %v", err)
	}
	defer nlp.CloseLanguageClient()

	// Init cron jobs
	cronjobs.InitCronJobs()

	r := routes.SetupRouter(firestoreClient, languageClient)
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
