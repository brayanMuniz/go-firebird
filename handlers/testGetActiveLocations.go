package handlers

import (
	"cloud.google.com/go/firestore"
	"github.com/gin-gonic/gin"
)

// TODO: make db request that will return the 10 highest location. use latestsentiment last entry to get skeetsAmount
// Will first have to modify the schema to house a new field
func TestGetActiveLocation(c *gin.Context, firestoreClient *firestore.Client) {

}
