package handlers

import (
	"cloud.google.com/go/firestore"
	language "cloud.google.com/go/language/apiv2"
	"context"
	"github.com/bluesky-social/indigo/xrpc"
	"github.com/gin-gonic/gin"
	"go-firebird/skeetprocessor"
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

func FetchBlueskyHandler(c *gin.Context, firestoreClient *firestore.Client, nlpClient *language.Client) {
	// Determine which feed to use based on query parameter "feed"
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

	// Read other query parameters.
	limit := c.DefaultQuery("limit", "10")
	cursor := c.Query("cursor")
	if cursor != "" {
		cursor = strings.ReplaceAll(cursor, " ", "+")
	}

	// Determine if we're using test (mock) data.
	testParam := c.DefaultQuery("test", "false")
	var out types.FeedResponse

	if testParam == "true" {
		// Use mock data instead of calling the live API.
		out = types.FeedResponse{
			Feed: []types.FeedEntry{
				{
					Post: types.Post{
						URI: "mock:uri:7",
						Author: types.Author{
							Avatar:      "https://example.com/avatar.png",
							Handle:      "testhandle",
							DisplayName: "Test User",
						},
						Record: types.Record{
							Text:      "Bushfire Cowlands Rd, Greenwald Status: bad",
							CreatedAt: "2025-03-14T00:00:00Z",
						},
					},
				},
			},
		}
		log.Println("Using mock test data")
	} else {
		// Prepare API parameters.
		params := map[string]interface{}{
			"feed":   feedAtURI,
			"limit":  limit,
			"cursor": cursor,
		}

		// Set up the xrpc client.
		client := &xrpc.Client{
			Client:    &http.Client{Timeout: 10 * time.Second},
			Host:      "https://public.api.bsky.app",
			UserAgent: nil,
		}

		// Call the Bluesky API.
		err := client.Do(context.Background(), xrpc.Query, "json", feedMethod, params, nil, &out)
		if err != nil {
			log.Printf("Error fetching feed via xrpc: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		log.Printf("Fetched feed from /api/firebird/blusky using feed: %s", feedAtURI)
	}

	resultsList := skeetprocessor.SaveFeed(out, firestoreClient, nlpClient)

	c.JSON(http.StatusOK, resultsList)
}
