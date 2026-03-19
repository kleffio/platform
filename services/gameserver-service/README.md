# gameserver-service

Go microservice responsible for game server lifecycle management on the Kleff platform.

## Architecture

Follows **Hexagonal Architecture** (Ports & Adapters):

```
gameserver-service/
├── cmd/
│   └── api/
│       └── main.go         # Application entry point
├── internal/
│   ├── adapters/           # Inbound/outbound adapters (HTTP handlers, DB clients, etc.)
│   ├── application/        # Use cases & application services
│   ├── bootstrap/          # Dependency wiring & server startup
│   ├── domain/             # Core domain models & business rules
│   └── ports/              # Interface definitions (driven & driving ports)
└── configs/                # Configuration files
```

### Layer Responsibilities

| Layer | Responsibility |
|---|---|
| `domain/` | Pure business entities and rules — no external dependencies |
| `ports/` | Interface contracts that the application core exposes and depends on |
| `application/` | Use case orchestration; calls domain logic, calls out through ports |
| `adapters/` | Concrete implementations of ports (HTTP, database, message queues, etc.) |
| `bootstrap/` | Wires everything together and starts the server |

## Getting Started

```bash
# Run locally
go run ./cmd/api

# Build binary
go build -o bin/api ./cmd/api
```

## Docker

Build and run from the **repository root**:

```bash
docker build -f services/gameserver-service/Dockerfile -t kleff-gameserver-service .
docker run -p 8082:8080 kleff-gameserver-service
```

## Module

```
github.com/kleff/gameserver-service
```

Go version: `1.23`
