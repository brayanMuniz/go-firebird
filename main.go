package main

import (
	"fmt"
	"go-firebird/cronjobs"
	"go-firebird/routes"
	"log"
	"os"

	"github.com/joho/godotenv"
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

	// Initialize cron jobs
	cronjobs.InitCronJobs()

	r := routes.SetupRouter()
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
