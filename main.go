package main

import (
	database "knowledgeMart/config"
	"knowledgeMart/routes"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDB()

	router := gin.Default()

	routes.RegisterRoutes(router)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	err := router.Run(":" + port)
	if err != nil {
		panic(err)
	}
}
