package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-firebird/types"
	"io"
	"net/http"
	"os"
	"time"
)

func GetDemoData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": "Demo disaster dataset"})
}

func SimulateDisasterTweets(c *gin.Context) {
	fmt.Println("STARTING SIMULATION")

	// Read JSON file
	data, err := os.ReadFile("./notes/exampleList.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Parse JSON into TweetAnalysis struct
	var tweetData types.TweetAnalysis
	err = json.Unmarshal(data, &tweetData)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse JSON"})
		return
	}

	// true if all tweets good uwu
	allGood := true

	// Endpoint where tweets will be sent
	baseURL := os.Getenv("CLIENT_URL")
	url := fmt.Sprintf("%s/api/tweetHook", baseURL)

	// Loop through tweets and send them one by one
	for i, tweet := range tweetData.Tweets {
		wrappedTweet := map[string]interface{}{
			"tweets": []types.Tweet{tweet}, // Wrap it in a tweets array
		}

		// Marshal tweet into JSON
		tweetJSON, err := json.Marshal(wrappedTweet)
		if err != nil {
			fmt.Println("Error marshaling tweet:", err)
			continue
		}

		fmt.Println(bytes.NewBuffer(tweetJSON))
		// Send POST request
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(tweetJSON))
		if err != nil {
			fmt.Println("Error making request:", err)
			allGood = false
			continue
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			continue
		}

		// Log response
		fmt.Printf("Tweet %d sent. Response: %s\n", i+1, string(body))
		fmt.Printf("\n")

		// Delay before sending the next tweet
		time.Sleep(1 * time.Second)
	}

	if allGood {
		c.JSON(http.StatusOK, gin.H{"message": "All tweets sent successfully"})
	} else {
		c.JSON(http.StatusPartialContent, gin.H{"message": "Not all tweets were sent well :("})
	}

}

func TestClientHook(c *gin.Context) {
	// take a look at this log if you are confused as to why the path is the way it is
	absPath, _ := os.Getwd()
	fmt.Println("Current working directory:", absPath)

	// Read JSON file
	data, err := os.ReadFile("./notes/example.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	baseURL := os.Getenv("CLIENT_URL")
	url := fmt.Sprintf("%s/api/tweetHook", baseURL)

	// Send POST request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println("Error making request:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send request"})
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// Log the response
	fmt.Println("Response:", string(body))

	// Return response to client
	c.JSON(resp.StatusCode, gin.H{
		"message":  "Request sent",
		"response": string(body),
	})
}
