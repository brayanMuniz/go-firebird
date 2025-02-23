package handlers

import (
	"context"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"time"
)

// documentation: https://github.com/bluesky-social/indigo/blob/2503553ea604ea7f0bfa6a021b284b371ac2ac96/xrpc/xrpc.go#L114

// FeedResponse represents the response from app.bsky.feed.getFeed.
type FeedResponse struct {
	Cursor string        `json:"cursor"`
	Feed   []interface{} `json:"feed"`
}

const (
	feedMethod = "app.bsky.feed.getFeed"
)

// FetchBlueskyHandler fetches a hydrated feed using the Bluesky API.
func FetchBlueskyHandler(c *gin.Context) {
	// Initialize the xrpc client to use the public API endpoint.
	client := &xrpc.Client{
		Client:    &http.Client{Timeout: 10 * time.Second},
		Host:      "https://public.api.bsky.app", // Use the public endpoint for unauthenticated requests.
		UserAgent: nil,                           // can set a custom agent
	}

	// This is the fire URI
	feedAtURI := "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejsyozb6iq"

	// Read query parameters
	limit := c.DefaultQuery("limit", "30") // Default to 30 if not specified
	cursor := c.Query("cursor")            // Cursor for pagination, empty if not provided

	// Prepare parameters. The limit can be adjusted (min 1, max 100, default 50).
	params := map[string]interface{}{
		"feed":  feedAtURI,
		"limit": limit,
	}

	if cursor != "" {
		params["cursor"] = cursor
	}

	log.Printf("Fetching feed with params: %+v", params)

	var out FeedResponse

	// Call the Bluesky API using the xrpc client.
	err := client.Do(context.Background(), xrpc.Query, "json", feedMethod, params, nil, &out)
	if err != nil {
		log.Printf("Error fetching feed via xrpc: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Fetched feed via xrpc: %+v", out)
	c.JSON(http.StatusOK, out)
}
