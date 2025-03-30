package handlers

import (
	"encoding/json" // For JSON marshaling
	"fmt"           // For error formatting
	"go-firebird/db"
	// "go-firebird/processor" // Not needed for this handler
	// "go-firebird/types" // db.GetValidLocations already returns the correct type
	"log"
	"net/http"
	"os" // For file operations
	// "strings" // Not needed
	// "sync" // Not needed

	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

// ExportLocationsHandler fetches all valid locations and saves them to a local JSON file.
func ExportLocationsHandler(c *gin.Context, firestoreClient *firestore.Client) {
	log.Println("Received request to export locations...")

	// 1. Fetch all valid locations
	validLocations, err := db.GetValidLocations(firestoreClient)
	if err != nil {
		log.Printf("Error fetching valid locations for export: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch location data",
			"details": err.Error(),
		})
		return
	}

	if len(validLocations) == 0 {
		log.Println("No valid locations found to export.")
		c.JSON(http.StatusOK, gin.H{
			"message": "No valid locations found to export.",
			"count":   0,
		})
		return
	}

	log.Printf("Fetched %d valid locations for export.", len(validLocations))

	// 2. Marshal data to JSON (with indentation for readability)
	jsonData, err := json.MarshalIndent(validLocations, "", "  ") // Use indentation ("  ")
	if err != nil {
		log.Printf("Error marshaling location data to JSON: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to format location data",
			"details": err.Error(),
		})
		return
	}

	// 3. Define filename and create/write the file
	filename := "locations_export.json"
	file, err := os.Create(filename) // Creates the file in the current working directory
	if err != nil {
		log.Printf("Error creating export file '%s': %v", filename, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create export file",
			"details": err.Error(),
		})
		return
	}
	// Ensure file is closed even if errors occur later
	defer func() {
		if cerr := file.Close(); cerr != nil {
			log.Printf("Error closing export file '%s': %v", filename, cerr)
		}
	}()

	_, err = file.Write(jsonData)
	if err != nil {
		log.Printf("Error writing JSON data to file '%s': %v", filename, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to write data to export file",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Successfully exported %d locations to %s", len(validLocations), filename)

	// 4. Return success response
	c.JSON(http.StatusOK, gin.H{
		"message":  fmt.Sprintf("Successfully exported %d locations.", len(validLocations)),
		"filename": filename, // Inform the user where the file was saved
		"count":    len(validLocations),
	})
}
