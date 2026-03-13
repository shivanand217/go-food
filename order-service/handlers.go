package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func createOrder(c *gin.Context) {
	var input struct {
		UserID       uint   `json:"user_id" binding:"required"`
		RestaurantID string `json:"restaurant_id" binding:"required"`
		DeliveryAddr string `json:"delivery_address" binding:"required"`
		Items        []struct {
			MenuItemID string  `json:"menu_item_id" binding:"required"`
			Quantity   int     `json:"quantity" binding:"required"`
			Price      float64 `json:"price" binding:"required"`
		} `json:"items" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var totalAmount float64
	var orderItems []OrderItem
	for _, item := range input.Items {
		totalAmount += item.Price * float64(item.Quantity)
		orderItems = append(orderItems, OrderItem{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
			Price:      item.Price,
		})
	}

	order := Order{
		UserID:       input.UserID,
		RestaurantID: input.RestaurantID,
		TotalAmount:  totalAmount,
		Status:       StatusCreated,
		DeliveryAddr: input.DeliveryAddr,
		Items:        orderItems,
	}

	if err := DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Publish Kafka Event
	if err := publishOrderCreatedEvent(order); err != nil {
		log.Printf("Failed to publish order event: %v", err)
		// We still return 201 because the order was saved in DB
		// In a real system, you'd use the Outbox pattern
	}

	c.JSON(http.StatusCreated, order)
}

func getOrder(c *gin.Context) {
	id := c.Param("id")
	var order Order
	if err := DB.Preload("Items").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func getUserOrders(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseUint(userIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var orders []Order
	if err := DB.Preload("Items").Where("user_id = ?", uint(userId)).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user orders"})
		return
	}
	
	// Be nice and return empty array instead of null
	if orders == nil {
		orders = []Order{}
	}

	c.JSON(http.StatusOK, orders)
}
