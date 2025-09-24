# Banking Microservices Architecture

This architecture is designed to support a bank with 30,000+ customers performing daily transactions.

## Services Overview

1. **API Gateway** - Entry point for all client requests
2. **Authentication Service** - Handles user authentication and authorization
3. **Account Service** - Manages customer accounts and balances
4. **Transaction Service** - Processes financial transactions
5. **Notification Service** - Handles customer notifications (email, SMS)

## Technology Stack

- Go for high-performance services
- PostgreSQL for persistent data storage
- Redis for caching
- Kafka for event streaming
- Docker for containerization
- Kubernetes for orchestration

## Architecture Diagram

```
┌─────────────┐     ┌─────────────┐
│  Web/Mobile │     │    Admin    │
│   Clients   │     │  Dashboard  │
└──────┬──────┘     └──────┬──────┘
       │                   │
       └─────────┬─────────┘
                 │
         ┌───────▼───────┐
         │  API Gateway  │
         └───────┬───────┘
                 │
     ┌───────────┼───────────┐
     │           │           │
┌────▼─────┐┌────▼─────┐┌────▼─────┐
│   Auth   ││  Account ││Transaction│
│ Service  ││ Service  ││ Service  │
└────┬─────┘└────┬─────┘└────┬─────┘
     │           │           │
     └───────────┼───────────┘
                 │
         ┌───────▼───────┐
         │ Notification │
         │   Service    │
         └───────────────┘
```

## Deployment

All services are containerized using Docker and can be deployed using the provided docker-compose.yml file.