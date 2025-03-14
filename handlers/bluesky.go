package handlers

import (
	"context"
	"fmt"
	"go-firebird/db"
	"go-firebird/mlmodel"
	"go-firebird/nlp"
	"go-firebird/types"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cloud.google.com/go/firestore"
	language "cloud.google.com/go/language/apiv2"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/gin-gonic/gin"
)

// documentation: https://github.com/bluesky-social/indigo/blob/2503553ea604ea7f0bfa6a021b284b371ac2ac96/xrpc/xrpc.go#L114
const (
	feedMethod = "app.bsky.feed.getFeed"
)

type SaveSkeetResult struct {
	SavedSkeetID         string        `json:"savedSkeetId"`
	Content              string        `json:"content"`
	NewLocationNames     []string      `json:"newLocationNames"`
	ProcessedEntityCount int           `json:"processedEntityCount"`
	Classification       []float64     `json:"classification"`
	Sentiment            nlp.Sentiment `json:"sentiment"`
	AlreadyExist         bool          `json:"alreadyExist"`
	ErrorSaving          bool          `json:"errorSaving"`
}

func saveSkeet(newSkeet db.Skeet, firestoreClient *firestore.Client, nlpClient *language.Client) (SaveSkeetResult, error) {
	ctx := context.Background()
	hashedSkeetID := db.HashString(newSkeet.UID)

	var result SaveSkeetResult
	result.SavedSkeetID = hashedSkeetID
	result.Content = newSkeet.Content
	result.AlreadyExist = false
	result.ErrorSaving = false

	// Check if the skeet already exists.
	_, err := firestoreClient.Collection("skeets").Doc(hashedSkeetID).Get(ctx)
	if err == nil {
		// The document exists, so return early.
		fmt.Printf("Doc already exists for: %s. Hash: %s\n", newSkeet.UID, hashedSkeetID)
		result.AlreadyExist = true
		result.SavedSkeetID = hashedSkeetID
		return result, nil
	} else if status.Code(err) != codes.NotFound {
		// An error other than not found occurred.
		result.ErrorSaving = true
		return result, err
	}

	// Prepare the ML input.
	mlInputs := mlmodel.MLRequest{
		newSkeet.UID: newSkeet.Content,
	}

	// We'll run ML model call, entity extraction, and sentiment analysis concurrently.
	var (
		classification         []float64
		nlpEntities            []nlp.Entity
		sentiment              nlp.Sentiment
		mlErr, nlpErr, sentErr error
	)
	var wg sync.WaitGroup
	wg.Add(3)

	// ML Model call.
	go func() {
		defer wg.Done()
		mlResp, err := mlmodel.CallModel(mlInputs)
		if err != nil {
			mlErr = err
			return
		}
		classification = mlResp[newSkeet.UID]
	}()

	// Entity extraction.
	go func() {
		defer wg.Done()
		var err error
		nlpEntities, err = nlp.AnalyzeEntities(nlpClient, newSkeet.Content)
		if err != nil {
			// Log the error but allow processing to continue with an empty slice.
			log.Printf("Error analyzing entities: %v", err)
			nlpEntities = []nlp.Entity{}
			nlpErr = err
		}
	}()

	// Sentiment analysis.
	go func() {
		defer wg.Done()
		var err error
		sentiment, err = nlp.AnalyzeSentiment(nlpClient, newSkeet.Content)
		if err != nil {
			log.Printf("Error analyzing sentiment: %v", err)
			sentErr = err
		}
	}()

	// Wait for all goroutines to finish.
	wg.Wait()
	// Mark as used
	_ = nlpErr
	_ = sentErr

	// Optionally, you can decide how to handle errors from ML, entity extraction or sentiment.
	if mlErr != nil {
		return result, mlErr
	}
	// Even if nlpErr or sentErr occurred, you might continue if you want to allow partial processing.

	// Set result values.
	result.Classification = classification
	result.Sentiment = sentiment

	// Prepare data for saving.
	data := db.SaveCompleteSkeetType{
		NewSkeet:       newSkeet,
		Classification: classification,
		Entities:       nlpEntities,
		Sentiment:      sentiment,
	}

	// Save everything in one transaction. This returns an array of new location names.
	newLocations, err := db.SaveCompleteSkeet(firestoreClient, data)
	if err != nil {
		result.ErrorSaving = true
		return result, err
	}
	result.NewLocationNames = newLocations

	// Count the number of entities that are either LOCATION or ADDRESS.
	count := 0
	for _, entity := range data.Entities {
		if entity.Type == "LOCATION" || entity.Type == "ADDRESS" {
			count++
		}
	}
	result.ProcessedEntityCount = count

	// Launch a goroutine for each new location geocoding update.
	var geoWg sync.WaitGroup
	for _, locationName := range newLocations {
		geoWg.Add(1)
		go func(loc string) {
			defer geoWg.Done()
			db.UpdateLocationGeocoding(firestoreClient, loc)
		}(locationName)
	}
	geoWg.Wait()

	return result, nil
}

// FetchBlueskyHandler fetches a hydrated feed using the Bluesky API.
func FetchBlueskyHandler(c *gin.Context, firestoreClient *firestore.Client, nlpClient *language.Client) {
	feedParam := c.DefaultQuery("feed", "f")
	var feedAtURI string
	switch feedParam {
	case "e":
		feedAtURI = "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejxlobe474" // earthquake feed
	case "h":
		feedAtURI = "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejwgffwqky" // hurricane feed
	default:
		feedAtURI = "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejsyozb6iq" // fire feed
	}

	client := &xrpc.Client{
		Client:    &http.Client{Timeout: 10 * time.Second},
		Host:      "https://public.api.bsky.app", // public endpoint for unauthenticated requests.
		UserAgent: nil,                           // can set a custom agent
	}

	// Read query parameters
	limit := c.DefaultQuery("limit", "10") // Default to 10 if not specified
	cursor := c.Query("cursor")            // Cursor for pagination, empty if not provided
	// Request removes the +, so we add it back
	if cursor != "" {
		cursor = strings.ReplaceAll(cursor, " ", "+")
	}

	// The limit can be adjusted (min 1, max 100, default 50).
	params := map[string]interface{}{
		"feed":   feedAtURI,
		"limit":  limit,
		"cursor": cursor,
	}

	var out types.FeedResponse // Using the structured response schema in types

	// Call the Bluesky API
	err := client.Do(context.Background(), xrpc.Query, "json", feedMethod, params, nil, &out)
	if err != nil {
		log.Printf("Error fetching feed via xrpc: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Fetched feed from /api/firebird/blusky")

	resultsChan := make(chan SaveSkeetResult, len(out.Feed))
	var wg sync.WaitGroup

	for _, v := range out.Feed {
		if v.Post.URI != "" {
			wg.Add(1)
			feedItem := v
			go func() {
				defer wg.Done()

				newSkeet := db.Skeet{
					Avatar:      feedItem.Post.Author.Avatar,
					Content:     feedItem.Post.Record.Text,
					Handle:      feedItem.Post.Author.Handle,
					DisplayName: feedItem.Post.Author.DisplayName,
					UID:         feedItem.Post.URI,
					Timestamp:   feedItem.Post.Record.CreatedAt,
				}

				savedSkeetResult, err := saveSkeet(newSkeet, firestoreClient, nlpClient)
				if err != nil {
					// Instead of aborting, assign a default result with error flag.
					savedSkeetResult = SaveSkeetResult{
						SavedSkeetID:         "",
						NewLocationNames:     nil,
						ProcessedEntityCount: 0,
						Classification:       nil,
						Sentiment:            nlp.Sentiment{},
						AlreadyExist:         false,
						Content:              "",
						ErrorSaving:          true,
					}
				}
				resultsChan <- savedSkeetResult
			}()
		}
	}

	wg.Wait()
	close(resultsChan)

	resultsList := make([]SaveSkeetResult, 0, len(out.Feed))
	for result := range resultsChan {
		resultsList = append(resultsList, result)
	}

	c.JSON(http.StatusOK, resultsList)

}
