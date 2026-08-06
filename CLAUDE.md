# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the **Backend Portal** service for Paper Indonesia's payment gateway system. It serves as the core backend managing all business logic for payments, disbursements, merchants, users, and financial operations. The service integrates with Orchestrator, Core Processors (SNAP Core and Credit Card), and supports OpenAPI and MerchantPortal.

## Development Commands

### Application Execution
```bash
# Run HTTP server (main application)
make run-http

# Run message queue consumer
make run-consumer  

# Run specific cron job manually
make run-cron command=<cron-job-name> date="YYYY-mm-dd HH:MM:SS"

# Run console command
make run-console command=<console-name>
```

### Database Management
```bash
# Apply pending migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check migration status
make migrate-status

# Create new migration file
make new-migration MIGRATION_NAME=<descriptive_name>
```

### Testing & Quality
```bash
# Run all tests with coverage report
make test

# Run only unit tests (fast)
make run-test

# Run only integration tests
make run-integration-test

# Run both unit and integration tests
make run-all-test

# Generate API documentation
make swagger

# Lint code with custom rules
make lint
```

### Code Generation
```bash
# Generate all mocks (repository, service, controller layers)
make gen-mocks

# Build Docker image
make build
```

## Architecture Overview

This service follows **Clean Architecture** patterns with clear separation of concerns:

### Layer Structure
- **`cmd/`** - CLI commands and application entry points (serveHttp, serveConsumer, serveCron, serveConsole)
- **`internal/model/`** - Domain models, DTOs, requests/responses for each business domain
- **`internal/repository/`** - Data access layer with MySQL, Redis, and MongoDB implementations  
- **`internal/service/`** - Business logic layer (contains both v1 and v2 implementations)
- **`port/`** - Application interfaces (HTTP controllers, message consumers, cron jobs)
- **`pkg/`** - Shared utilities and external integrations

### Key Business Domains
- **Payment Processing** - Payments, refunds, payment methods
- **Financial Operations** - Disbursements, transfers, ledger, settlements
- **Merchant Management** - Merchant onboarding, fees, authentication
- **User & Access Control** - Users, roles, permissions, JWT auth
- **Integration Services** - QRIS, virtual accounts, credit cards, webhooks

## Technology Stack

- **Go 1.24** with Chi router for HTTP handling
- **MySQL 8.0** as primary database with SQLX for queries
- **Redis** for caching and rate limiting  
- **RabbitMQ** for message queuing
- **Goose** for database migrations
- **OpenTelemetry/New Relic** for observability
- **Consul** for feature flags
- **Testcontainers** for integration testing

## Development Setup

### Prerequisites
- Go 1.24+
- Docker & Docker Compose 
- Optional: Nix 2.18.5+ with direnv

### Quick Start
```bash
# 1. Copy configuration files
cp .example.config.yaml .config.yaml
cp .example.secret.yaml .secret.yaml

# 2. Download dependencies  
go mod download -x

# 3. Start infrastructure services
docker-compose up -d

# 4. Run database migrations
make migrate-up

# 5. Start HTTP server
make run-http

# Access points:
# - API: http://localhost:3000
# - Health Check: http://localhost:3000/api/v1/health-check
# - Swagger Docs: http://localhost:3000/swagger/index.html
```

### Alternative: Nix Setup
```bash
# 1. Clone repository and navigate to directory
# 2. Allow direnv configuration
direnv allow

# 3. Copy configuration files (same as above)
# 4. Download dependencies (same as above)
# 5. Start infrastructure services with Nix
nix develop --impure -c up

# 6. Start HTTP server (in new terminal)
make run-http
```

## Important Notes

### Configuration
- Application uses dual config system: `.config.yaml` (app settings) and `.secret.yaml` (sensitive data)
- Configuration is loaded via Viper with environment variable override support
- Timezone is hardcoded to UTC in main.go

### Database
- Uses table partitioning for high-volume tables (account_transactions, payments, etc.)
- All timestamp columns use Go's time.Time (not nullable)
- Foreign key constraints have been removed for performance (handled in application layer)

### Testing
- Integration tests require `INTEGRATION=1` environment variable
- Uses testcontainers for real database testing
- Mocks are generated via mockery for all layers

### Code Quality  
- Custom golangci-lint configuration in `scanner-config/`
- Git hooks in `githooks/` for pre-commit validation
- Swagger documentation auto-generated from code annotations

### Deployment
- Multi-stage Docker builds with Alpine runtime
- Helm charts available in `deployments/helm/` for Kubernetes
- Observability stack includes Prometheus, Grafana, Jaeger

### Service Integration
- Heavy use of repository pattern for data access
- Service layer implements business logic with proper error handling
- Uses Paper Indonesia's internal PDK (Platform Development Kit) libraries
- Conductor SDK integration for workflow management
# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.