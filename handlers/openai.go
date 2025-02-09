package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"log"
	"net/http"
	"os"
)

type TweetAnalysis struct {
	Tweets []struct {
		OriginalTweetData struct {
			TweetText string `json:"tweet_text" jsonschema:"default="`
			Timestamp string `json:"timestamp" jsonschema:"default="`
			User      struct {
				Username       string `json:"username" jsonschema:"default="`
				Verified       bool   `json:"verified" jsonschema:"default=false"`
				FollowersCount int    `json:"followers_count" jsonschema:"default=0"`
			} `json:"user"`
			Hashtags []string `json:"hashtags" jsonschema:"default=[]"`
		} `json:"original_tweet_data"`
		Analysis []struct {
			// Only allowed disasters
			DisasterName       string  `json:"disaster_name" jsonschema:"enum=Wildfire,Flood,Earthquake,Hurricane,Tornado,Tsunami,Volcanic Eruption,Landslide,Blizzard,Drought;default="`
			RelevantToDisaster float64 `json:"relevant_to_disaster" jsonschema:"default=0"`
			Sentiment          struct {
				Emotion  string  `json:"emotion" jsonschema:"default="`
				Tone     string  `json:"tone" jsonschema:"default="`
				Polarity float64 `json:"polarity" jsonschema:"default=0"`
				Details  struct {
					Urgency         float64 `json:"urgency" jsonschema:"default=0"`
					EmotionalImpact float64 `json:"emotional_impact" jsonschema:"default=0"`
				} `json:"details"`
			} `json:"sentiment"`
			SarcasmConfidence float64 `json:"sarcasm_confidence" jsonschema:"default=0"`
		} `json:"analysis"`
	} `json:"tweets"`
}

func CallOpenAI(c *gin.Context) {
	var request struct {
		Input string `json:"input"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))
	ctx := context.Background()

	// Generate JSON schema from Go struct
	var result TweetAnalysis
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		log.Fatalf("GenerateSchemaForType error: %v", err)
	}

	// OpenAI API Request
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleSystem,
				Content: "You are an AI assistant specializing in disaster tweet analysis. " +
					"Classify each tweet according to disaster relevance, analyze sentiment, and " +
					"estimate sarcasm confidence. " +
					"Return the results in the defined structured format. If a field is not provided, " +
					"leave it as its default value (e.g. empty strings, 0, false, or empty arrays).",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "Analyze the sentiment of the following tweet: " + request.Input,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "tweet_analysis",
				Schema: schema,
				Strict: true,
			},
		},
		MaxTokens: 500,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Unmarshal structured response
	err = schema.Unmarshal(resp.Choices[0].Message.Content, &result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse structured response"})
		return
	}

	// Convert to JSON and return to frontend
	c.JSON(http.StatusOK, result)
}
