package main

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Order status constants
const (
	StatusCreated   = "CREATED"
	StatusConfirmed = "CONFIRMED"
	StatusPreparing = "PREPARING"
	StatusOutForDel = "OUT_FOR_DELIVERY"
	StatusDelivered = "DELIVERED"
	StatusCancelled = "CANCELLED"
)

type Order struct {
	gorm.Model
	UserID       uint        `json:"user_id"`
	RestaurantID string      `json:"restaurant_id"`
	TotalAmount  float64     `json:"total_amount"`
	Status       string      `json:"status" gorm:"default:'CREATED'"`
	DeliveryAddr string      `json:"delivery_address"`
	Items        []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
	gorm.Model
	OrderID    uint    `json:"order_id"`
	MenuItemID string  `json:"menu_item_id"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
}

var DB *gorm.DB

func initDB() {
	dsn := getEnv("DATABASE_URL", "host=localhost user=gofood password=password dbname=gofood_db port=5432 sslmode=disable")
	var err error
	
	// Retry connection since DB might not be ready yet
	for i := 0; i < 5; i++ {
		DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to postgres, retrying in 2 seconds... (%v)", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Failed to connect to postgres after retries: %v", err)
	}

	// Migrate the schema
	DB.AutoMigrate(&Order{}, &OrderItem{})
	log.Println("Database connection established and models migrated")
}
