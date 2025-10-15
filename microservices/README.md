# Microservices Architecture

This folder contains the microservice-ready version of the application. Each service is designed to be independently deployable and scalable.

## Services Overview

### 1. Auth Service (`auth-service/`)
- Handles user registration and login
- JWT token generation and validation
- User authentication middleware

### 2. User Service (`user-service/`)
- User profile management
- User data operations
- Depends on Auth Service for authentication

### 3. Admin Service (`admin-service/`)
- Administrative operations
- User management (list, delete, update roles)
- Depends on Auth Service for authentication

## Architecture Benefits

- **Independent Scaling**: Scale services based on demand
- **Technology Flexibility**: Use different languages/frameworks per service
- **Team Autonomy**: Different teams can work on different services
- **Fault Isolation**: Service failures are contained
- **Easier Testing**: Test services in isolation

## Getting Started

Each service has its own `main.go` and can be run independently:

```bash
cd microservices/auth-service
go run main.go

cd microservices/user-service
go run main.go

cd microservices/admin-service
go run main.go
```

## Future Migration

To migrate from monolithic to microservices:
1. Deploy each service independently
2. Set up service discovery (Consul, Kubernetes)
3. Implement inter-service communication (gRPC/HTTP)
4. Add API Gateway for routing
5. Configure shared databases or implement event-driven architecture
