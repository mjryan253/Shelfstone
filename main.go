package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/user/libreplex/database" // Import the database package
)

func main() {
	// Initialize database connection
	database.ConnectDatabase()
	log.Println("Database initialized successfully.")

	r := gin.Default()

	// API v1 group
	apiV1 := r.Group("/api")
	{
		apiV1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status": "ok",
			})
		})
	}

	r.Run(":8080") // Listen and serve on 0.0.0.0:8080
}
