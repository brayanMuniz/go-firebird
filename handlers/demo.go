package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetDemoData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"data": "Demo disaster dataset"})
}
