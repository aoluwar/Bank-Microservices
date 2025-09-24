# Banking Microservices Documentation

## Overview
This banking microservices system is designed to handle 30,000+ customers performing daily transactions. The architecture follows microservices principles for scalability, resilience, and maintainability.

## Architecture Components

### 1. API Gateway
- **Purpose**: Single entry point for all client requests
- **Features**: Request routing, authentication middleware, load balancing
- **Port**: 8000
- **Endpoints**:
  - `/auth/*` → Authentication Service
  - `/accounts/*` → Account Service
  - `/transactions/*` → Transaction Service

### 2. Authentication Service
- **Purpose**: User authentication and authorization
- **Port**: 8082
- **Key Endpoints**:
  - `POST /auth/register` - Register new user
  - `POST /auth/login` - Authenticate user and issue JWT
  - `GET /auth/validate` - Validate JWT token
  - `GET /auth/users/{id}` - Get user details
  - `PUT /auth/users/{id}` - Update user details
  - `PUT /auth/users/{id}/password` - Change password

### 3. Account Service
- **Purpose**: Manage customer accounts
- **Port**: 8080
- **Key Endpoints**:
  - `GET /accounts` - List all accounts
  - `GET /accounts/{id}` - Get account details
  - `POST /accounts` - Create new account
  - `PUT /accounts/{id}` - Update account details
  - `GET /accounts/{id}/balance` - Get account balance
  - `POST /accounts/{id}/deposit` - Deposit funds
  - `POST /accounts/{id}/withdraw` - Withdraw funds

### 4. Transaction Service
- **Purpose**: Process and record financial transactions
- **Port**: 8081
- **Key Endpoints**:
  - `GET /transactions` - List all transactions
  - `GET /transactions/{id}` - Get transaction details
  - `POST /transactions` - Create new transaction
  - `GET /accounts/{id}/transactions` - Get account transactions
  - `POST /transactions/transfer` - Transfer funds between accounts

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Accounts Table
```sql
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    account_number VARCHAR(20) UNIQUE NOT NULL,
    account_type VARCHAR(20) NOT NULL,
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(10) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Transactions Table
```sql
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    transaction_type VARCHAR(20) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    source_account_id INTEGER REFERENCES accounts(id),
    destination_account_id INTEGER REFERENCES accounts(id),
    status VARCHAR(10) NOT NULL DEFAULT 'completed',
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## Deployment

### Prerequisites
- Docker and Docker Compose
- Go 1.19+
- PostgreSQL 13+
- Redis 6+

### Setup and Deployment
1. Clone the repository
2. Create necessary Go module files in each service directory:
   ```bash
   cd account-service
   go mod init bank/account-service
   go mod tidy
   # Repeat for other services
   ```
3. Start the services:
   ```bash
   cd microservices
   docker-compose up -d
   ```

### Scaling Considerations
- Each service can be horizontally scaled independently
- Use Kubernetes for production deployment
- Implement database sharding for high transaction volumes
- Add Redis caching for frequently accessed data

## Security Considerations
- JWT tokens for authentication
- HTTPS for all communications
- Password hashing with bcrypt
- Environment variables for sensitive configuration
- Regular security audits

## Monitoring and Logging
- Implement Prometheus for metrics collection
- Use Grafana for visualization
- Centralized logging with ELK stack
- Health check endpoints for each service