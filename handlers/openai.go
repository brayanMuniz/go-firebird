package handlers

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
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

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an assistant specializing in analyzing the sentiment of tweets.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Analyze the sentiment of the following tweet: " + request.Input,
				},
			},
			MaxTokens: 100,
		},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"classification": resp.Choices[0].Message.Content,
	})
}
