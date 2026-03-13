package main

import (
	"log"
	"encoding/json"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var Producer *kafka.Producer

type OrderCreatedEvent struct {
	OrderID      uint    `json:"order_id"`
	UserID       uint    `json:"user_id"`
	RestaurantID string  `json:"restaurant_id"`
	DeliveryAddr string  `json:"delivery_address"`
	TotalAmount  float64 `json:"total_amount"`
}

func initKafkaProducer() {
	broker := getEnv("KAFKA_BROKERS", "localhost:29092")
	var err error

	Producer, err = kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}

	// Delivery report handler for produced messages
	go func() {
		for e := range Producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v\n", ev.TopicPartition)
				} else {
					log.Printf("Delivered message to %v\n", ev.TopicPartition)
				}
			}
		}
	}()
	
	log.Println("Kafka Producer initialized")
}

func publishOrderCreatedEvent(order Order) error {
	topic := getEnv("KAFKA_ORDER_TOPIC", "orders")

	event := OrderCreatedEvent{
		OrderID:      order.ID,
		UserID:       order.UserID,
		RestaurantID: order.RestaurantID,
		DeliveryAddr: order.DeliveryAddr,
		TotalAmount:  order.TotalAmount,
	}

	value, err := json.Marshal(event)
	if err != nil {
		return err
	}

	err = Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          value,
	}, nil)

	return err
}
