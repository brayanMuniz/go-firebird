package handlers

import (
	"cloud.google.com/go/firestore"
	"context"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/gin-gonic/gin"
	"go-firebird/db"
	"go-firebird/types"
	"log"
	"net/http"
	"strings"
	"time"
)

// documentation: https://github.com/bluesky-social/indigo/blob/2503553ea604ea7f0bfa6a021b284b371ac2ac96/xrpc/xrpc.go#L114
const (
	feedMethod = "app.bsky.feed.getFeed"
)

// FetchBlueskyHandler fetches a hydrated feed using the Bluesky API.
func FetchBlueskyHandler(c *gin.Context, firestoreClient *firestore.Client) {
	// Initialize the xrpc client to use the public API endpoint.
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

	// Call the Bluesky API using the xrpc client.
	err := client.Do(context.Background(), xrpc.Query, "json", feedMethod, params, nil, &out)
	if err != nil {
		log.Printf("Error fetching feed via xrpc: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Fetched feed from /api/firebird/blusky")

	// Save the first skeet from the feed into the database
	first := out.Feed[0]
	if first.Post.URI != "" {
		newSkeet := db.Skeet{
			Author:  first.Post.Author.DisplayName,
			Content: first.Post.Record.Text,
		}
		db.WriteSkeet(firestoreClient, newSkeet.Author, newSkeet.Content)

	}

	c.JSON(http.StatusOK, out)
}
