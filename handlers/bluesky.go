package handlers

import (
	"context"
	"fmt"
	"go-firebird/db"
	"go-firebird/nlp"
	"go-firebird/types"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	language "cloud.google.com/go/language/apiv2"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/gin-gonic/gin"
)

// documentation: https://github.com/bluesky-social/indigo/blob/2503553ea604ea7f0bfa6a021b284b371ac2ac96/xrpc/xrpc.go#L114
const (
	feedMethod = "app.bsky.feed.getFeed"
)

// FetchBlueskyHandler fetches a hydrated feed using the Bluesky API.
func FetchBlueskyHandler(c *gin.Context, firestoreClient *firestore.Client, nlpClient *language.Client) {
	client := &xrpc.Client{
		Client:    &http.Client{Timeout: 10 * time.Second},
		Host:      "https://public.api.bsky.app", // public endpoint for unauthenticated requests.
		UserAgent: nil,                           // can set a custom agent
	}

	// Interchange these for testing

	// fire: https://bsky.app/profile/did:plc:qiknc4t5rq7yngvz7g4aezq7/feed/aaaejsyozb6iq
	feedAtURI := "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejsyozb6iq"

	// earthquake: https://bsky.app/profile/did:plc:qiknc4t5rq7yngvz7g4aezq7/feed/aaaejxlobe474
	// feedAtURI := "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejxlobe474"

	// hurricane: https://bsky.app/profile/faineg.bsky.social/feed/aaaejwgffwqky
	// feedAtURI := "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejwgffwqky"

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

	first := out.Feed[0]
	if first.Post.URI != "" {
		// Parse the createdAt string into a time.Time
		parsedTime, err := time.Parse(time.RFC3339, first.Post.Record.CreatedAt)
		if err != nil {
			log.Printf("Failed to parse createdAt: %v", err)
			parsedTime = time.Now() // fallback to current time if parsing fails
		}

		// Save the first skeet from the feed into the database
		newSkeet := db.Skeet{
			Author:      first.Post.Author.DisplayName,
			Content:     first.Post.Record.Text,
			Handle:      first.Post.Author.Handle,
			DisplayName: first.Post.Author.DisplayName,
			UID:         first.Post.URI,
			Timestamp:   parsedTime,
		}

		db.WriteSkeet(firestoreClient, newSkeet)

		// Perform NLP analysis on the skeet content.
		fmt.Println(newSkeet.Content)
		nlpEntities, err := nlp.AnalyzeEntities(nlpClient, newSkeet.Content)
		if err != nil {
			log.Printf("Error analyzing entities: %v", err)
		} else {
			// Loop over NLP entities
			for _, entity := range nlpEntities {
				fmt.Printf("Entity Name: %s, Type: %s\n", entity.Name, entity.Type)
				if entity.Type == "LOCATION" || entity.Type == "ADDRESS" {
					// Save the location metadata and associate the skeet in a nested subcollection.
					err := db.SaveLocationEntity(firestoreClient, entity.Name, first.Post.URI, entity, newSkeet)
					if err != nil {
						log.Printf("Error saving location entity for %s: %v", entity.Name, err)
					} else {
						fmt.Printf("Saved location '%s' with skeet ID '%s'\n", entity.Name, first.Post.URI)
					}
				}
				//  Print additional details about the entity.
				if len(entity.Metadata) > 0 {
					fmt.Println("  Metadata:")
					for k, v := range entity.Metadata {
						fmt.Printf("    %s: %s\n", k, v)
					}
				}
				if len(entity.Mentions) > 0 {
					fmt.Println("  Mentions:")
					for _, m := range entity.Mentions {
						fmt.Printf("    Content: %s, BeginOffset: %d, Probability: %.3f\n", m.Content, m.BeginOffset, m.Probability)
					}
				}
				fmt.Println("")
			}
		}

	}

	c.JSON(http.StatusOK, out)
}
