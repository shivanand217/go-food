# Swiggy Clone Microservices Walkthrough

The backend for the Swiggy platform is now fully orchestrated and functioning natively as a loosely-coupled microservices system utilizing event-driven architecture!

## Architecture Summary
We built the following stack, orchestrated entirely through Docker Compose:
- **API Gateway**: A single entry point (port `8080`) that proxies all traffic to the respective downstream microservices.
- **User Service** *(Postgres)*: Handles customer registration and authentication.
- **Restaurant Service** *(MongoDB)*: Manages restaurant details and their dynamic menu catalogs.
- **Order Service** *(Postgres + Kafka)*: Processes incoming user orders, acts as a Kafka Producer, and emits an [OrderCreated](file:///Users/shivprakash/.gemini/antigravity/scratch/swiggy-clone/order-service/kafka.go#12-19) event to a topic.
- **Delivery Service** *(Postgres + Kafka)*: Manages delivery agents, listens as a Kafka Consumer for new order events, and automatically assigns available delivery agents to orders.
- **Infrastructure**: Zookeeper, Kafka, Postgres, MongoDB, and Redis running in parallel containers.

## Testing the Flow
With `docker-compose up -d` running, you can trace a full transaction through the system.

### 1. Create a User
```bash
curl -X POST http://localhost:8080/api/users/register \
-H "Content-Type: application/json" \
-d '{"name": "Shiv", "email": "shiv@example.com", "password": "password123", "phone": "12345"}' 
```

### 2. Create a Restaurant & Menu Item
First create the restaurant (Note the `id` returned):
```bash
curl -X POST http://localhost:8080/api/restaurants/ \
-H "Content-Type: application/json" \
-d '{"name": "Pizza Hut", "description": "Best Pizzas", "address": "123 Main St"}'
```
Using the MongoDB `id` returned, add a menu item:
```bash
curl -X POST http://localhost:8080/api/restaurants/<restaurant_id>/menu \
-H "Content-Type: application/json" \
-d '{"name": "Margherita", "price": 10.99, "is_veg": true}'
```

### 3. Place an Order (Triggers Kafka)
To trigger the automated delivery pipeline, place an order:
```bash
curl -X POST http://localhost:8080/api/orders/ \
-H "Content-Type: application/json" \
-d '{"user_id": 1, "restaurant_id": "<restaurant_id>", "delivery_address": "456 Side St", "items": [{"menu_item_id": "menu123", "quantity": 2, "price": 10.99}]}'
```

Because of our asynchronous event-driven design, the *Order Service* immediately returns `201 Created` to the user and transparently places a message onto the Kafka `orders` topic. 

The *Delivery Service* picks up that message and automatically queries its local Postgres database for an agent where `status = AVAILABLE`, assigns them to the new order, and marks the agent as `BUSY`!

You can verify the delivery agents' statuses by hitting:
```bash
curl -s http://localhost:8080/api/delivery/agents
```
*(Notice that there are originally two agents seeded on initialization; after placing one order, only one agent will remain in the array because the other has been auto-assigned.)*
