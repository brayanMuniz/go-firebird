package handlers

import (
	"context"
	"fmt"
	"go-firebird/db"
	"go-firebird/detection"
	"go-firebird/summarization"
	"go-firebird/types"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
)

// RunDisasterDetection fetches candidate locations, runs detection, generates summaries, and saves results.
func RunDisasterDetection(c *gin.Context, firestoreClient *firestore.Client) {
	const detectionSentimentThreshold float32 = 0.0 // -0.3

	log.Println("Handler: Starting disaster detection process...")

	// 1. Fetch candidate locations
	locations, err := db.GetLocationsForDisasterCheck(firestoreClient, detectionSentimentThreshold)
	if err != nil {
		log.Printf("ERROR fetching locations for disaster check: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve locations for analysis"})
		return
	}
	if len(locations) == 0 {
		log.Println("Handler: No locations found meeting the disaster check criteria.")
		c.JSON(http.StatusOK, gin.H{"message": "No locations met criteria", "disasters": []types.DisasterData{}})
		return
	}

	log.Printf("Handler: Running detection logic on %d locations...", len(locations))

	// 2. Run the detection logic
	disasters, detectErr := detection.DetectDisastersFromList(locations)
	if detectErr != nil {
		log.Printf("ERROR during disaster detection logic: %v", detectErr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An error occurred during disaster analysis"})
		return
	}
	if len(disasters) == 0 {
		log.Println("Handler: Detection logic identified 0 disaster clusters.")
		c.JSON(http.StatusOK, gin.H{"message": "No disaster clusters identified", "disasters": []types.DisasterData{}})
		return
	}

	log.Printf("Handler: Detection complete. Found %d disaster clusters. Proceeding to summarization...", len(disasters))

	// 3. Generate Summaries (if disasters were found)
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("Warning: OPENAI_API_KEY environment variable not set. Skipping summary generation.")
	} else {
		openaiClient := openai.NewClient(apiKey)
		ctxSummarize, cancelSummarize := context.WithTimeout(context.Background(), 2*time.Minute) // Use background context for potentially long task
		defer cancelSummarize()

		err = summarization.GenerateSummaries(ctxSummarize, disasters, firestoreClient, openaiClient)
		if err != nil {
			// Log the error but continue, some summaries might be missing
			log.Printf("Warning: Error occurred during summary generation: %v", err)
		}
	}

	// 4. Save the detected disasters (with or without summaries)
	log.Printf("Handler: Saving %d detected disasters to the database...", len(disasters))
	err = db.SaveDisasters(firestoreClient, disasters)
	if err != nil {
		log.Printf("ERROR saving disasters to database: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Failed to save detected disasters",
			"disasters": disasters,
		})
		return
	}

	log.Println("Handler: Process complete. Disasters saved.")
	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Processed %d locations, identified and saved %d disasters.", len(locations), len(disasters)),
		"disasters": disasters,
	})
}
