package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	r := gin.Default()

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "up",
			"service": "api-gateway",
		})
	})

	// Reverse proxy setup
	setupProxy(r, "/api/users/*path", getEnv("USER_SERVICE_URL", "http://localhost:8081"))
	setupProxy(r, "/api/restaurants/*path", getEnv("RESTAURANT_SERVICE_URL", "http://localhost:8082"))
	setupProxy(r, "/api/orders/*path", getEnv("ORDER_SERVICE_URL", "http://localhost:8083"))
	setupProxy(r, "/api/delivery/*path", getEnv("DELIVERY_SERVICE_URL", "http://localhost:8084"))

	port := getEnv("PORT", "8080")
	log.Printf("API Gateway starting on port %s", port)
	
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start API Gateway: %v", err)
	}
}

// setupProxy creates a reverse proxy route
func setupProxy(r *gin.Engine, route string, targetURL string) {
	url, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Invalid target URL for %s: %v", targetURL, err)
	}
	
	proxy := httputil.NewSingleHostReverseProxy(url)

	r.Any(route, func(c *gin.Context) {
		// Remove the prefix if necessary based on how downstream expects the path, 
		// but typically we'll keep the context path or strip it. Let's send the whole path down for now.
		proxy.ServeHTTP(c.Writer, c.Request)
	})
}

// getEnv gets an environment variable or returns a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
