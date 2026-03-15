# Go-Food Architecture Design

This document outlines the High-Level Design (HLD) and Low-Level Design (LLD) for the Go-Food backend microservices. These diagrams will help developers understand how communication flows between the different services.

## High-Level Design (HLD)

The HLD provides a bird's-eye view of the entire system architecture, showing all the deployed microservices, their dedicated databases, and the event streaming platform used for asynchronous communication.

> [!NOTE]
> Every microservice manages its own decoupled database to ensure that there are no single points of failure and to allow each service to scale independently based on load.

```mermaid
graph TD
    %% User/Client
    Client([Client App / Web]) -->|HTTP REST| APIGateway(API Gateway\nPort 8080)

    %% API Gateway Routing
    APIGateway -->|/api/users| UserService(User Service)
    APIGateway -->|/api/restaurants| RestaurantService(Restaurant Service)
    APIGateway -->|/api/orders| OrderService(Order Service)
    APIGateway -->|/api/delivery| DeliveryService(Delivery Service)

    %% Databases
    UserService -->|Read/Write| UserDB[(PostgreSQL\nUser DB)]
    RestaurantService -->|Read/Write| RestaurantDB[(MongoDB\nRestaurant/Menu DB)]
    OrderService -->|Read/Write| OrderDB[(PostgreSQL\nOrder DB)]
    DeliveryService -->|Read/Write| DeliveryDB[(PostgreSQL\nDelivery/Agent DB)]

    %% Message Broker
    subgraph Event Backbone
        Kafka[[Apache Kafka\nTopic: orders]]
        Zookeeper[Zookeeper]
        Zookeeper -.->|Manages| Kafka
    end

    %% Event Publishing / Consuming
    OrderService -->|Produces Event\n'OrderCreated'| Kafka
    Kafka -->|Consumes Event\n'OrderCreated'| DeliveryService

    classDef service fill:#4a86e8,stroke:#000,stroke-width:2px,color:#fff;
    classDef database fill:#f6b26b,stroke:#000,stroke-width:2px;
    classDef broker fill:#8e7cc3,stroke:#000,stroke-width:2px,color:#fff;
    classDef gateway fill:#e06666,stroke:#000,stroke-width:2px,color:#fff;

    class APIGateway gateway;
    class UserService,RestaurantService,OrderService,DeliveryService service;
    class UserDB,RestaurantDB,OrderDB,DeliveryDB database;
    class Kafka broker;
```

---

## Low-Level Design (LLD)

The LLD diagrams zoom into specific system flows. The most complex flow in our current architecture is the **Order Placement and Delivery Assignment Flow**.

### Order Placement and Auto-Assignment Sequence

This sequence diagram details the exact synchronous API calls and asynchronous Kafka events that occur when a user successfully checks out an order.

```mermaid
sequenceDiagram
    actor User
    participant API as API Gateway
    participant OrderSvc as Order Service
    participant OrderDB as Order DB (Postgres)
    participant Kafka as Kafka (Topic: orders)
    participant DelSvc as Delivery Service
    participant DelDB as Delivery DB (Postgres)

    Note over User, API: 1. User places a new order
    User->>API: POST /api/orders/ {restaurant_id, items...}
    API->>OrderSvc: Proxy Request
    
    Note over OrderSvc, OrderDB: 2. Save Order to Database
    OrderSvc->>OrderDB: Insert Order (Status: CREATED)
    OrderDB-->>OrderSvc: Return Order ID
    
    Note over OrderSvc, Kafka: 3. Publish Event asynchronously
    OrderSvc-)Kafka: Publish 'OrderCreatedEvent' (order_id, user_id, amount)
    
    Note over OrderSvc, User: 4. Respond to user immediately
    OrderSvc-->>API: 201 Created (Order Data)
    API-->>User: 201 Created (Order Data)
    
    Note over Kafka, DelDB: 5. Background Delivery Processing
    Kafka-->>DelSvc: Consume 'OrderCreatedEvent' message
    
    DelSvc->>DelDB: Query `DeliveryAgent` WHERE status = 'AVAILABLE' LIMIT 1
    alt Agent Found
        DelDB-->>DelSvc: Return Agent details
        DelSvc->>DelDB: Insert `DeliveryTask` (agent_id, order_id, status: ASSIGNED)
        DelSvc->>DelDB: Update `DeliveryAgent` SET status = 'BUSY'
    else No Agents Available
        DelDB-->>DelSvc: Not Found
        Note over DelSvc: Order remains in unassigned queue<br/>(Retry mechanism/DLQ needed)
    end
```

### Component Breakdown

| Microservice | Responsibility | DB | Integration Points |
|--------------|----------------|----|-------------------|
| **API Gateway** | Request routing, Rate limiting (planned), Auth parsing | N/A | Proxies to all internal services |
| **User Service** | Registration, Authentication, Profile Management | Postgres | None (sync response only) |
| **Restaurant** | Restaurant onboarding, Menu item configuration | MongoDB | None (sync response only) |
| **Order Service** | Checkout computations, Order tracking/history | Postgres | **Produces** to Kafka |
| **Delivery** | Delivery agent lifecycle, fleet tracking, geo-location | Postgres | **Consumes** from Kafka |
