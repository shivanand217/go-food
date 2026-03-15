# Go-Food Backend Clone (Golang Microservices)

A production-level backend ecosystem for a food delivery platform (Go-Food/UberEats clone) built completely in Go using Microservices and Event-Driven Architecture.

## System Architecture
This ecosystem comprises 5 distinct microservices orchestrated together.
- **API Gateway** (Port `8080`)
- **User Service** (Port `8081`): Postgres
- **Restaurant Service** (Port `8082`): MongoDB
- **Order Service** (Port `8083`): Postgres + Kafka Producer
- **Delivery Service** (Port `8084`): Postgres + Kafka Consumer
- **Infrastructure**: Zookeeper, Apache Kafka, PostgreSQL, MongoDB, and Redis.

---

## 🚀 Step 1: Clone and Run
To run the entire ecosystem locally, all you need is Docker and Docker Compose. No local Go installation is required because the images use multi-stage builds.

1. **Clone the repository**
   ```bash
   git clone <your-repo-url>
   cd go-food
   ```

2. **Start the cluster**
   ```bash
   docker-compose up --build -d
   ```
   *Note: This will download the Postgres, Mongo, Kafka, Zookeeper, and Redis images, and then compile the Go binaries for all 5 microservices. It may take a few minutes on the first run.*

3. **Verify the cluster is running**
   ```bash
   docker-compose ps
   ```
   You can also check the API Gateway Health endpoint:
   ```bash
   curl -s http://localhost:8080/health
   ```
   *(Should return `{"service":"api-gateway","status":"up"}`)*

---

## 🧪 Step 2: Test the Ecosystem (End-to-End)

We will use the API Gateway on `localhost:8080` to interact with our microservices.

### 1. Register a User (User Service)
Create an account on the platform.
```bash
curl -X POST http://localhost:8080/api/users/register \
-H "Content-Type: application/json" \
-d '{
  "name": "Shiv",
  "email": "shiv@example.com",
  "password": "password123",
  "phone": "12345"
}'
```

### 2. Create a Restaurant (Restaurant Service)
Onboard a new restaurant.
```bash
curl -X POST http://localhost:8080/api/restaurants/ \
-H "Content-Type: application/json" \
-d '{
  "name": "Pizza Hut",
  "description": "Best Pizzas",
  "address": "123 Main St"
}'
```
> **IMPORTANT:** Copy the `id` from the response (e.g., `69b50...`). You will use it in the next steps. Let's refer to it as `<restaurant_id>`.

### 3. Add a Menu Item (Restaurant Service)
Add a pizza to your restaurant's menu.
```bash
curl -X POST http://localhost:8080/api/restaurants/<restaurant_id>/menu \
-H "Content-Type: application/json" \
-d '{
  "name": "Margherita",
  "price": 10.99,
  "is_veg": true
}'
```

---

## ⚡ Step 3: Trigger the Event-Driven Flow (Order & Delivery Services)

The real magic happens when you place an order.
1. The **Order Service** validates and saves the order to Postgres.
2. The **Order Service** asynchronously drops an `OrderCreated` payload onto the Apache Kafka message broker.
3. The **Delivery Service** constantly listens to Kafka. It picks up the event, searches for the nearest `AVAILABLE` delivery agent in its dedicated Postgres database, and assigns them the task, changing their status to `BUSY`.

### 1. Check Available Agents
By default, the Delivery service seeds 2 available agents into its database on startup. Let's check them:
```bash
curl -s http://localhost:8080/api/delivery/agents
```
*(You should see an array of two available agents).*

### 2. Place the Order!
```bash
curl -X POST http://localhost:8080/api/orders/ \
-H "Content-Type: application/json" \
-d '{
  "user_id": 1,
  "restaurant_id": "<restaurant_id>",
  "delivery_address": "456 Side St",
  "items": [
    {
      "menu_item_id": "menu123",
      "quantity": 2,
      "price": 10.99
    }
  ]
}'
```

### 3. Verify Agent Auto-Assignment
Query the available agents endpoint one more time:
```bash
curl -s http://localhost:8080/api/delivery/agents
```
You will notice only **one** agent remains available. The missing agent has been automatically converted from `AVAILABLE` to `BUSY` via our asynchronous Kafka event successfully!

---

## Shutting Down
To wipe the cluster and remove all containers:
```bash
docker-compose down -v
```
