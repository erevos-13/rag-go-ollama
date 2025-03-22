package routes

import (
	"net/http"

	"github.com/erevos-13/rag-go-ollama/models"
	"github.com/gin-gonic/gin"
)

type Query struct {
	Query string `json:"query"`
}

func SearchByDocument(c *gin.Context) {
	var query Query
	err := c.ShouldBindJSON(&query)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	completion, results, err := models.SearchDocument(query.Query, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"completion": completion, "results": results})

}

func UploadDocument(c *gin.Context) {
	err := models.UpdateDocument(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Document uploaded and indexed successfully"})
}
