package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"go-firebird/cronjobs"
	"go-firebird/db"
	"go-firebird/routes"
	"log"
	"os"
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

	// Init firestore
	client, err := db.InitFirestore()
	if err != nil {
		log.Fatalf("Failed to initialize Firestore: %v", err)
	}
	defer db.CloseFirestore() // Firestore client is closed on exit

	// Initialize cron jobs
	cronjobs.InitCronJobs()

	r := routes.SetupRouter(client)
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
