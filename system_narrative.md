# The Story of Go-Food: A Complete System Deep Dive

Welcome to the inner workings of **Go-Food**. Building a microservices backend isn't just about writing code; it's about choreographing a complex dance between completely independent programs. 

This document is your guided tour. We will walk step-by-step through how the system boots up, how the services talk to each other, and exactly what happens behind the scenes from the moment a hungry user opens the app to the moment a delivery agent is assigned.

---

## 1. The Boot Sequence: Bringing the Ecosystem to Life

When you run `docker-compose up`, you are not starting one program; you are starting **ten separate servers** on the same virtual network. 

1. **The Infrastructure Layer Wakes Up:**
   Before any of our Go code runs, the fundamental building blocks start:
   - **Zookeeper & Kafka**: The message broker spins up. Kafka needs Zookeeper to manage its clustering and topics.
   - **Databases**: Three PostgreSQL containers and one MongoDB container start up. 
   - **Redis**: A caching layer boots up (reserved for future rate-limiting).

2. **The Microservices Start & Connect:**
   Once the databases are alive, Docker begins compiling our Go code. 
   - The **User**, **Order**, and **Delivery** services boot, and each one independently attempts to connect to its own dedicated Postgres database via `GORM`. If the DB isn't ready yet, they automatically retry every 2 seconds until they secure a connection.
   - The **Restaurant** service boots and connects to MongoDB.
   - Finally, the **API Gateway** boots on port `8080`. It doesn't connect to any database; its sole job is routing traffic.

At this point, you have an entire corporate-scale datacenter running inside your laptop.

---

## 2. Communication Rules: REST vs. Events

In Go-Food, microservices communicate in two fundamentally different ways:

### A. Synchronous HTTP (REST)
When a user wants an immediate answer ("What is the menu?"), the system uses HTTP.
* **The API Gateway** takes the incoming request from the mobile app (e.g., `GET /api/restaurants`), looks at its path, forwards that HTTP request down to the `Restaurant Service`, waits for a response, and returns it to the user.
* This is tight coupling; if the Restaurant service goes down, the API Gateway fails to return the menu.

### B. Asynchronous Event-Driven (Kafka)
When a process doesn't need an immediate user response but must trigger business logic elsewhere ("The order is paid, find a driver"), the system uses **Apache Kafka**.
* Instead of calling the Delivery Service directly, the Order Service packages the order details into an "Event Payload", drops it on the Kafka `orders` topic, and forgets about it. 
* The Delivery Service constantly polls that Kafka topic. When it sees the event, it processes it.
* This is loose coupling; if the Delivery Service crashes, the Order Service doesn't care. Kafka holds the message until the Delivery Service comes back online.

---

## 3. The Flow: Journey of a Pizza Order

Let's follow a complete, end-to-end user journey through the system.

### Step 1: The User Arrives
* The user hits `POST /api/users/register`.
* The **API Gateway** identifies the `/api/users` route and proxies the JSON payload to the **User Service** (running on port `8081`).
* The **User Service** hashes the password via `bcrypt`, inserts the user record into its PostgreSQL database, and returns a success response back through the gateway.

### Step 2: Surfing the Menu
* The user wants to see what's available and hits `GET /api/restaurants`.
* The **API Gateway** proxies this to the **Restaurant Service** (`8082`). 
* Because restaurant menus contain highly fluid, deeply nested array data (categories, items, variants, toppings), the **Restaurant Service** uses **MongoDB** (a NoSQL document database) rather than rigid SQL tables. It fetches the JSON document and returns it.

### Step 3: Placing the Order (The Critical Path)
The user selects a Margherita pizza and clicks "Checkout". This triggers the most complex operation in the system.

1. **Intake**: 
   The app calls `POST /api/orders/` with the array of items. The **API Gateway** sends this to the **Order Service** (`8083`).
2. **Pricing & Persistence**: 
   The Order Service calculates the total cart value. It opens a transaction with its dedicated PostgreSQL database, creates the [Order](file:///Users/shivprakash/.gemini/antigravity/scratch/swiggy-clone/order-service/models.go#21-30) record, and inserts the `OrderItems`.
3. **The Microservice Handoff (Kafka Producer)**:
   The Order Service is done with its user-facing job. It returns a `201 Created` HTTP response to the user so the mobile app can show "Order Successful!". 
   But the backend work is just beginning. The Order Service constructs an [OrderCreatedEvent](file:///Users/shivprakash/.gemini/antigravity/scratch/swiggy-clone/delivery-service/consumer.go#11-18) struct, serializes it to a JSON string, and fires it into the **Apache Kafka** `orders` topic. 

### Step 4: The Delivery Assignment (The Kafka Consumer)
Meanwhile, in a completely separate container, the **Delivery Service** (`8084`) is running an infinite looping Goroutine in the background, listening to the Kafka `orders` topic.

1. **The Catch**: 
   The Delivery Service intercepts the [OrderCreatedEvent](file:///Users/shivprakash/.gemini/antigravity/scratch/swiggy-clone/delivery-service/consumer.go#11-18) message from Kafka.
2. **Agent Query**: 
   It reaches into its own PostgreSQL database (which only it has access to) and executes a query: `SELECT * FROM delivery_agents WHERE status = 'AVAILABLE' LIMIT 1`.
3. **The Assignment**: 
   It finds our hero, "Agent 1". It creates a [DeliveryTask](file:///Users/shivprakash/.gemini/antigravity/scratch/swiggy-clone/delivery-service/models.go#30-37) linking Agent 1 to the incoming Order ID. It then runs an `UPDATE` query, switching Agent 1's status from `AVAILABLE` to `BUSY`. 

If all agents are busy, the Delivery Service would normally leave the message in a "Retry Queue" (or Dead Letter Queue) and try again in 5 minutes. 

---

## The Beauty of Microservices

Why execute the transaction this way instead of writing one massive monolith?

1. **Scalability**: If Go-Food goes viral and order volume spikes 1000x, we can launch 50 copies of the **Order Service** container to handle the web traffic, while leaving the **User Service** running as just a single container.
2. **Resiliency**: If there's a bug in the code that controls Delivery Agent geo-tracking, the **Delivery Service** might crash. But the mobile app won't break! Users can still open the menu, view their cart, and checkout seamlessly because the **Order Service** is still alive, dropping messages onto Kafka to be processed whenever the Delivery service restarts.
3. **Data Security**: It is physically impossible for the Order Service to accidentally corrupt the Delivery Database, because the Order Service doesn't even have the password to that database. 

You now have full visibility into the anatomy of the Go-Food ecosystem!
