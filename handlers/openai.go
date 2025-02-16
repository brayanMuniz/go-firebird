package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"go-firebird/types"
	"log"
	"net/http"
	"os"
)

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
	var result types.TweetAnalysis
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
					"Return the results in the EXACT JSON format that follows this schema, ensuring every field is properly structured. " +
					"Do NOT return any extra text before or after the JSON.",
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
		MaxTokens: 1500,
	})

	log.Printf("Raw OpenAI Response: %s", resp.Choices[0].Message.Content)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Unmarshal structured response
	err = schema.Unmarshal(resp.Choices[0].Message.Content, &result)
	if err != nil {
		log.Printf("Failed to parse OpenAI response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OpenAI returned invalid JSON format"})
		return
	}

	// Convert to JSON and return to frontend
	c.JSON(http.StatusOK, result)
}
