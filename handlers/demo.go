package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func GetDemoData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": "Demo disaster dataset"})
}

func TestClientHook(c *gin.Context) {
	fmt.Println("Reading example.json")

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

	// Tweet hook endpoint
	url := "http://localhost:3000/api/tweetHook"

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
