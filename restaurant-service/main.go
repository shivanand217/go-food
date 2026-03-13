package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	initDB()

	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up", "service": "restaurant-service"})
	})

	api := r.Group("/api/restaurants")
	{
		api.POST("/", createRestaurant)
		api.GET("/", listRestaurants)
		api.GET("/:id", getRestaurant)
		api.POST("/:id/menu", addMenuItem)
		api.GET("/:id/menu", getMenu)
	}

	port := getEnv("PORT", "8082")
	log.Printf("Restaurant Service starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
