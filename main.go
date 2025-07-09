package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/user/libreplex/database" // Import the database package
)

func main() {
	// Initialize database connection using the default production path
	database.ConnectDefaultProductionDB()
	// The global DB instance can now be accessed via database.GetDB() if needed by handlers
	log.Println("Default production database initialized successfully.")

	r := gin.Default()

	// API v1 group
	apiV1 := r.Group("/api")
	{
		apiV1.GET("/health", func(c *gin.Context) {
			// You could add a DB ping here if desired, using database.GetDB()
			c.JSON(http.StatusOK, gin.H{
				"status":  "UP",
				"message": "ShelfStone is running",
			})
		})
		// Example of using the DB in a handler (not part of original health check)
		// apiV1.GET("/users", func(c *gin.Context) {
		// 	var users []database.User
		//  // Use database.GetDB() to access the initialized GORM instance
		// 	result := database.GetDB().Find(&users)
		// 	if result.Error != nil {
		// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		// 		return
		// 	}
		// 	c.JSON(http.StatusOK, users)
		// })
	}

	r.Run(":8080") // Listen and serve on 0.0.0.0:8080
}
