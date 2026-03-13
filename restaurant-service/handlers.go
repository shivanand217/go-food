package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func createRestaurant(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Address     string `json:"address" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	restaurant := Restaurant{
		Name:        input.Name,
		Description: input.Description,
		Address:     input.Address,
		Rating:      0.0,
		Menu:        []MenuItem{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := RestaurantCollection.InsertOne(ctx, restaurant)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create restaurant"})
		return
	}

	restaurant.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, restaurant)
}

func listRestaurants(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := RestaurantCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch restaurants"})
		return
	}
	defer cursor.Close(ctx)

	var restaurants []Restaurant
	if err := cursor.All(ctx, &restaurants); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse restaurants"})
		return
	}

	// Be nice and return empty array instead of null
	if restaurants == nil {
		restaurants = []Restaurant{}
	}

	c.JSON(http.StatusOK, restaurants)
}

func getRestaurant(c *gin.Context) {
	idParam := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restaurant ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var restaurant Restaurant
	err = RestaurantCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&restaurant)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	c.JSON(http.StatusOK, restaurant)
}

func addMenuItem(c *gin.Context) {
	idParam := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restaurant ID"})
		return
	}

	var item MenuItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item.ID = primitive.NewObjectID()
	if !item.IsAvailable && !item.IsVeg && item.Price == 0 {
		// Just some default setup if omitted
		item.IsAvailable = true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{
		"$push": bson.M{"menu": item},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	result, err := RestaurantCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil || result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found or failed to update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Menu item added successfully", "item": item})
}

func getMenu(c *gin.Context) {
	idParam := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restaurant ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var restaurant Restaurant
	err = RestaurantCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&restaurant)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	if restaurant.Menu == nil {
		restaurant.Menu = []MenuItem{}
	}

	c.JSON(http.StatusOK, restaurant.Menu)
}
