package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type OrderCreatedEvent struct {
	OrderID      uint    `json:"order_id"`
	UserID       uint    `json:"user_id"`
	RestaurantID string  `json:"restaurant_id"`
	DeliveryAddr string  `json:"delivery_address"`
	TotalAmount  float64 `json:"total_amount"`
}

func initKafkaConsumer() {
	broker := getEnv("KAFKA_BROKERS", "localhost:29092")
	groupID := getEnv("KAFKA_CONSUMER_GROUP", "delivery-service-group")
	topic := getEnv("KAFKA_ORDER_TOPIC", "orders")

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	c.SubscribeTopics([]string{topic}, nil)
	log.Printf("Listening for events on topic %s", topic)

	for {
		msg, err := c.ReadMessage(time.Second)
		if err == nil {
			log.Printf("Received msg on %s: %s\n", msg.TopicPartition, string(msg.Value))
			handleOrderCreated(msg.Value)
		} else if !err.(kafka.Error).IsTimeout() {
			log.Printf("Consumer error: %v (%v)\n", err, msg)
		}
	}
}

func handleOrderCreated(value []byte) {
	var event OrderCreatedEvent
	if err := json.Unmarshal(value, &event); err != nil {
		log.Printf("Error unmarshalling event: %v\n", err)
		return
	}

	// Logic to assign an agent
	var agent DeliveryAgent
	if err := DB.Where("status = ?", AgentStatusAvailable).First(&agent).Error; err != nil {
		log.Printf("No available agents for order %d\n", event.OrderID)
		// Usually you'd put the order in a queue or re-try later
		return
	}

	// Create task
	task := DeliveryTask{
		OrderID:         event.OrderID,
		DeliveryAgentID: agent.ID,
		Status:          TaskStatusAssigned,
		DeliveryAddress: event.DeliveryAddr,
	}

	if err := DB.Create(&task).Error; err != nil {
		log.Printf("Error creating delivery task: %v\n", err)
		return
	}

	// Update agent status
	DB.Model(&agent).Update("status", AgentStatusBusy)

	log.Printf("Assigned order %d to agent %d", event.OrderID, agent.ID)
}
