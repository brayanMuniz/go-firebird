package handlers

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/bluesky-social/indigo/xrpc"
	"github.com/gin-gonic/gin"

	"go-firebird/nlp"
	"go-firebird/types"

	language "cloud.google.com/go/language/apiv2"
)

func TestSentiment(c *gin.Context, nlpClient *language.Client) {
	mockParam := c.Query("mock")

	// JSON response struct
	type SkeetSentimentResponse struct {
		Author    string        `json:"author"`
		Content   string        `json:"content"`
		Sentiment nlp.Sentiment `json:"sentiment"`
	}

	// If mockParam == "t", use mock data instead of fetching from Bluesky.
	if mockParam == "t" {
		mockAuthor := "Mock Tester"
		mockContent := "This fire is pretty bad boys!"

		responseData := SkeetSentimentResponse{
			Author:  mockAuthor,
			Content: mockContent,
		}

		// Analyze entities using the mock text.
		sentiment, err := nlp.AnalyzeSentiment(nlpClient, mockContent)
		if err != nil {
			log.Printf("Error analyzing entities (mock mode): %v", err)
			c.JSON(http.StatusOK, gin.H{
				"author":  responseData.Author,
				"content": responseData.Content,
				"error":   "Failed to analyze entities in mock mode",
			})
			return
		}
		responseData.Sentiment = sentiment

		// Return JSON response with mock data and entities.
		c.JSON(http.StatusOK, responseData)
		return
	}

	client := &xrpc.Client{
		Client:    &http.Client{Timeout: 10 * time.Second},
		Host:      "https://public.api.bsky.app",
		UserAgent: nil,
	}

	const feedMethod = "app.bsky.feed.getFeed"

	// fire uri
	feedAtURI := "at://did:plc:qiknc4t5rq7yngvz7g4aezq7/app.bsky.feed.generator/aaaejsyozb6iq"

	// Read query parameters
	cursor := c.Query("cursor")
	if cursor != "" {
		cursor = strings.ReplaceAll(cursor, " ", "+")
	}

	// Prepare request params
	params := map[string]interface{}{
		"feed":   feedAtURI,
		"limit":  1,
		"cursor": cursor,
	}

	var out types.FeedResponse

	// Fetch the feed from Bluesky
	if err := client.Do(context.Background(), xrpc.Query, "json", feedMethod, params, nil, &out); err != nil {
		log.Printf("Error fetching feed via xrpc: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If there's no post in the feed :(
	if len(out.Feed) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No feed items returned"})
		return
	}

	// Extract the first post
	first := out.Feed[0]
	if first.Post.URI == "" {
		c.JSON(http.StatusOK, gin.H{"message": "First feed item has no valid URI"})
		return
	}
	responseData := SkeetSentimentResponse{
		Author:  first.Post.Author.DisplayName,
		Content: first.Post.Record.Text,
	}

	// Analyze entities using the mock text.
	sentiment, err := nlp.AnalyzeSentiment(nlpClient, responseData.Content)
	if err != nil {
		log.Printf("Error analyzing entities (mock mode): %v", err)
		c.JSON(http.StatusOK, gin.H{
			"author":  responseData.Author,
			"content": responseData.Content,
			"error":   "Failed to analyze entities in mock mode",
		})
		return
	}
	responseData.Sentiment = sentiment

	// Return JSON
	c.JSON(http.StatusOK, responseData)
}
