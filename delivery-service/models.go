package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	AgentStatusAvailable = "AVAILABLE"
	AgentStatusBusy      = "BUSY"
	AgentStatusOffline   = "OFFLINE"
)

const (
	TaskStatusAssigned  = "ASSIGNED"
	TaskStatusPickedUp  = "PICKED_UP"
	TaskStatusDelivered = "DELIVERED"
)

type DeliveryAgent struct {
	gorm.Model
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Status  string `json:"status" gorm:"default:'AVAILABLE'"`
}

type DeliveryTask struct {
	gorm.Model
	OrderID         uint   `json:"order_id"`
	DeliveryAgentID uint   `json:"agent_id"`
	Status          string `json:"status" gorm:"default:'ASSIGNED'"`
	DeliveryAddress string `json:"delivery_address"`
}

var DB *gorm.DB

func initDB() {
	dsn := getEnv("DATABASE_URL", "host=localhost user=swiggy password=password dbname=swiggy_db port=5432 sslmode=disable")
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}

	// Migrate the schema
	DB.AutoMigrate(&DeliveryAgent{}, &DeliveryTask{})
	log.Println("Database connection established and models migrated")
	
	// Seed some agents if none exist
	var count int64
	DB.Model(&DeliveryAgent{}).Count(&count)
	if count == 0 {
		agents := []DeliveryAgent{
			{Name: "Agent 1", Phone: "1234567890", Status: AgentStatusAvailable},
			{Name: "Agent 2", Phone: "0987654321", Status: AgentStatusAvailable},
		}
		DB.Create(&agents)
	}
}
