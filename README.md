# Order & Payment Microservices (gRPC + Contract-First)

## Overview

This project is a microservices-based system consisting of:

- Order Service (REST API, Gin)
- Payment Service (gRPC server)
- Notification Service (NATS + Redis worker)
- PostgreSQL for persistence
- Redis for caching and idempotency
- NATS JetStream for event-driven communication

The system demonstrates:
- Clean Architecture principles
- gRPC communication between services
- Event-driven messaging
- Idempotent order creation
- Payment processing workflow
- Retry mechanism with DLQ
- Contract-first service design

## Architecture

Order Service (REST API)
→ gRPC →
Payment Service
→ NATS event (payment.completed) →
Notification Service
→ Redis + retry logic + DLQ

## Services

### Order Service

REST API built with Gin.

Endpoints:
- POST /orders – create order
- GET /orders/:id – get order
- PATCH /orders/:id/cancel – cancel order

Features:
- Calls Payment Service via gRPC
- Idempotency key support
- Redis caching for performance
- Order states: Pending, Paid, Failed, Cancelled

---

### Payment Service

gRPC-based service responsible for payment processing.

Business rules:
- Amount > 100000 → Declined
- Otherwise → Authorized

Features:
- Stores payments in PostgreSQL
- Publishes events to NATS JetStream

---

### Notification Service

Background worker consuming events from NATS.

Features:
- Listens to payment.completed events
- Sends notifications via EmailSender interface
- Retry mechanism with exponential backoff
- Redis-based deduplication
- Dead Letter Queue (DLQ) for failed events

---

## Messaging (NATS)

Streams:
- PAYMENTS → payment.completed
- PAYMENTS_DLQ → payment.dlq

---

## Database Schema

### Orders

CREATE TABLE orders (
    id VARCHAR(36) PRIMARY KEY,
    customer_id VARCHAR(50) NOT NULL,
    customer_email VARCHAR(100) NOT NULL,
    item_name VARCHAR(100) NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    idempotency_key VARCHAR(50) UNIQUE
);

---

### Payments

CREATE TABLE payments (
    id VARCHAR(36) PRIMARY KEY,
    order_id VARCHAR(36) NOT NULL,
    transaction_id VARCHAR(36),
    customer_email VARCHAR(100) NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL
);

---

## Running the Project

Run all services using Docker Compose:

docker-compose up --build

---

## Service URLs

- Order Service → http://localhost:8080
- Payment Service (gRPC) → localhost:50051
- NATS Monitoring → http://localhost:8222
- PostgreSQL → localhost:5432
- Redis → localhost:6379

---

## Flow

1. Client creates order via REST API
2. Order Service calls Payment Service (gRPC)
3. Payment Service processes payment
4. Event is published to NATS
5. Notification Service consumes event
6. Email is sent (mock implementation)
7. Failed messages are retried or sent to DLQ

---

## Key Features

- Idempotent order creation
- Retry with exponential backoff
- Event-driven architecture (NATS JetStream)
- Clean Architecture structure
- gRPC contract-first communication
- Redis caching layer
- Dead Letter Queue (DLQ)

---

## Notes

- Email sending is simulated (mock provider)
- Payment rules are simplified for educational purposes
- Notification service is stateless and scalable

---

## License

Educational project for Advanced Programming course.
