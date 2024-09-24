package main

import (
	"knowledgeMart/config"
	"knowledgeMart/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	database.ConnectDB()

	router := gin.Default()

	routes.RegisterRoutes(router)

	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}

//askj dfbz vejp fgeg
