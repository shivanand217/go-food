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
	go initKafkaConsumer()

	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "up", "service": "delivery-service"})
	})

	api := r.Group("/api/delivery")
	{
		api.POST("/agents", addAgent)
		api.GET("/agents", getAvailableAgents)
		api.PUT("/tasks/:id/status", updateTaskStatus)
	}

	port := getEnv("PORT", "8084")
	log.Printf("Delivery Service starting on port %s", port)
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
