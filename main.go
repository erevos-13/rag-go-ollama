package main

import (
	"github.com/erevos-13/rag-go-ollama/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	server := gin.Default()
	server.Use(gin.Logger())
	server.Use(gin.Recovery())
	server.Use(cors.Default())

	server.POST("/document", routes.UploadDocument)
	server.POST("/document/search", routes.SearchByDocument)

	server.Run("0.0.0.0:3000")
}
