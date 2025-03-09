package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/xrpc"
	"github.com/gin-gonic/gin"

	"go-firebird/types"
)

type MLRequest struct {
	URI string `json:"uri"`
}

type MLResponse struct {
	URI []float64 `json:"uri"`
}

// fetches a post from Bluesky, sends its text to the ML model, and returns classification results.
func ClassifyLiveTest(c *gin.Context) {
	client := &xrpc.Client{
		Client:    &http.Client{Timeout: 10 * time.Second},
		Host:      "https://public.api.bsky.app",
		UserAgent: nil,
	}
	const feedMethod = "app.bsky.feed.getFeed"

	// Fire feed
	feedAtURI := "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejsyozb6iq"

	cursor := c.Query("cursor")
	if cursor != "" {
		cursor = strings.ReplaceAll(cursor, " ", "+")
	}

	params := map[string]interface{}{
		"feed":   feedAtURI,
		"limit":  1, // just fetch 1 post for demonstration
		"cursor": cursor,
	}

	var out types.FeedResponse
	if err := client.Do(context.Background(), xrpc.Query, "json", feedMethod, params, nil, &out); err != nil {
		log.Printf("Error fetching feed via xrpc: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(out.Feed) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No feed items returned"})
		return
	}

	first := out.Feed[0]
	if first.Post.URI == "" {
		c.JSON(http.StatusOK, gin.H{"message": "First feed item has no valid URI"})
		return
	}

	author := first.Post.Author.DisplayName
	uid := first.Post.URI
	text := first.Post.Record.Text

	// Build the request payload
	payload := MLRequest{URI: text}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshaling ML request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to build ML request"})
		return
	}

	mlURL := "https://firebirdmodel-165032778338.us-central1.run.app/tweets/"

	req, err := http.NewRequest(http.MethodPost, mlURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		log.Printf("Error creating request for ML model: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ML request"})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request to the ML model.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error sending request to ML model: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to contact ML model"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("ML model returned status: %d", resp.StatusCode)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ML model error"})
		return
	}

	// Parse the ML model's JSON response.
	var mlResp MLResponse
	if err := json.NewDecoder(resp.Body).Decode(&mlResp); err != nil {
		log.Printf("Error decoding ML response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse ML response"})
		return
	}

	// 3. Return JSON with classification + original data.
	c.JSON(http.StatusOK, gin.H{
		"author":         author,
		"uid":            uid,
		"content":        text,
		"classification": mlResp.URI,
	})
}
