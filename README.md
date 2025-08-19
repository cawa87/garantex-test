# Garantex Rate Service

gRPC service for retrieving USDT exchange rates from Garantex exchange.

## Features

- gRPC API with GetRates and HealthCheck endpoints
- PostgreSQL storage with automatic rate persistence
- Real-time rates from Garantex (BTC/USDT market)
- Prometheus metrics and structured logging
- Docker support with graceful shutdown
- OpenTelemetry tracing for request observability
- Database migrations with schema versioning
- Comprehensive code quality with golangci-lint

## Project Structure

```
.
├── cmd/app/main.go                 # Application entry point
├── gen/go/rate_service.v1/         # Generated protobuf files
├── internal/
│   ├── app/app.go                  # Application orchestration
│   ├── config/config.go            # Configuration management
│   ├── lib/logger/sl/sl.go         # Structured logging
│   ├── repository/postgres/        # Database layer
│   ├── service/exchange/           # Exchange API client
│   └── transport/grpc/             # gRPC server
├── migrations/                     # Database migrations
├── proto/rate_service.v1/          # Protobuf definitions
├── Dockerfile                      # Container build
├── docker-compose.yml              # Service orchestration
├── Makefile                        # Build automation
└── README.md                       # This file
```

## Quick Start

1. Build and run:
```bash
make build
docker compose up -d
```

2. Service endpoints:
   - gRPC: `localhost:50051`
   - Metrics: `http://localhost:9090/metrics`

3. Test the service:
```bash
go run examples/client.go
```

### Development

```bash
go mod download
make test
make lint
```

## Configuration

Environment variables:
- `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_DBNAME`
- `SERVER_GRPC_PORT`, `SERVER_METRICS_PORT`
- `EXCHANGE_BASE_URL`, `EXCHANGE_TIMEOUT`
- `LOG_LEVEL`

## API

### GetRates
Retrieves current BTC/USDT rates from Garantex.

### HealthCheck
Checks service health and dependencies.

## Commands

```bash
make build          # Build application
make test           # Run tests
make lint           # Run linter
make docker-build   # Build Docker image
make run-docker     # Start with Docker Compose
make stop-docker    # Stop services
```
