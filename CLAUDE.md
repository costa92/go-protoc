# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

A production-ready Go microservice framework built on Kratos v2 with Protocol Buffers as the single source of truth. Uses Clean Architecture with Wire dependency injection, supporting HTTP/gRPC APIs with comprehensive error handling, i18n, authentication, and observability.

## Quick Development Commands

### Essential Daily Commands
- `make help` - View all available commands
- `make run-api` - Start development server with hot reload
- `make build` - Build optimized binary to ./bin/apiserver
- `make test` - Run all tests with verbose output
- `make fmt` - Format and import order correction
- `make tidy` - Clean up go.mod dependencies

### Development Environment
- `make dev-setup` - Complete development environment setup (tools + services + code generation)
- `make dev-clean` - Clean development environment and stop services
- `make dev-quick` - Quick development start (skip tool installation)
- `make dev-watch` - File watching with auto-rebuild
- `make dev-test` - Full test pipeline with coverage
- `make dev-bench` - Run benchmarks

### Code Generation
- `make generate` - Generate protobuf/gRPC/HTTP code with buf
- `make wire` - Regenerate dependency injection code (run after structural changes)
- `buf generate` - Direct protobuf generation (if buf.yaml changed)

### Tool Installation
- `make install-tools` - Install CI tools only
- `make install-tools A=1` - Install all development tools
- `make tools.install.<tool>` - Install specific tool (wire, golangci-lint, buf, etc.)

### Infrastructure Services
- `make run-redis` - Start Redis with docker-compose
- `make stop-redis` - Stop Redis service
- `make run-jaeger` - Start Jaeger tracing
- `make stop-jaeger` - Stop Jaeger service  
- `make run-kafka` - Start Kafka service
- `make stop-kafka` - Stop Kafka service
- `make start-all` - Start all dependent services
- `make stop-all` - Stop all services

### Documentation
- `make docs.dev` - Generate development documentation to docs/DEVELOPMENT.md
- `make docs.api` - Generate API documentation from protobuf
- `make docs.serve` - Serve documentation locally

## Architecture Overview

### Clean Architecture Layers
```
cmd/apiserver/          → Entry points & orchestration
internal/apiserver/     → Core business logic
├── handler/           → HTTP/gRPC handlers (delivery)
├── biz/              → Use cases (application)
├── store/            → Data access (infrastructure)
└── config/           → Internal configuration

pkg/                  → Reusable packages for import
├── api/              → Protobuf definitions & generated code
├── errorsx/          → Context-aware error system with i18n
├── authn/           → JWT authentication utilities
├── db/              → Database abstractions
├── server/          → HTTP/gRPC server configurations
└── options/         → Component configuration schemas
```

### Dependency Injection (Wire)
- **Centralized in** `internal/apiserver/wire.go`
- **Generated factories** in `internal/apiserver/wire_gen.go`
- **Auto-discovery** - run `make wire` after adding new dependencies

### Request Flow
```
HTTP Request → gRPC-Gateway → Handlers → Biz → Store → Database
     ↓                                          ↓
  OpenAPI docs (auto-generated)      GORM + context transactions
```

## Configuration & Environment

### Local Setup
1. Copy and edit: `cp configs/apiserver.yaml configs/apiserver_local.yaml`
2. Configure database in local config
3. Start dependencies: `make run-redis` (or `docker-compose -f deployments/redis/docker-compose.yml up`)
4. Run: `go run cmd/apiserver/main.go -c configs/apiserver_local.yaml`

### Alternative Quick Setup
```bash
make dev-setup      # Install tools + start services + generate code
make run-api        # Start the API server
```

### Development Dependencies
- **Database**: MySQL 8.0+ or PostgreSQL 12+
- **Cache**: Redis 6.2+ (docker-compose provided)
- **Observability**: Jaeger, Prometheus (docker-compose provided)

### Configuration Files
- `configs/apiserver.yaml` - Default configuration
- `configs/apiserver_v1.yaml` - Alternative configuration template
- Environment-specific configs use format: `apiserver_<env>.yaml`

## Testing Commands

### Standard Testing
```bash
go test ./...                    # Run all tests
go test -v ./pkg/errorsx/       # Run package tests verbose
go test -run TestSpecific       # Run specific test
go test -bench=. ./...          # Run benchmarks
```

### Integration Testing
```bash
docker-compose -f deployments/redis/docker-compose.yml up -d
go test -tags=integration ./...   # Run integration tests
```

## API Development

### Adding New Endpoints
1. **API Definition**: Edit `pkg/api/apiserver/v1/apiserver.proto`
2. **Generate Code**: `make generate`
3. **Implement Handler**: Create in `internal/apiserver/handler/`
4. **Wire Dependency**: Run `make wire`
5. **Documentation**: Auto-generated at `api/openapi/apiserver/v1/`

### Key Development Conventions
- **Error codes**: Define in protobuf, auto-generated with `protoc-gen-go-errors-code`
- **Validation**: Use protoc-gen-validate annotations
- **i18n**: Use `pkg/i18n/` with context-based locale detection
- **Logging**: Structured logging via context middleware
- **Testing**: Test names follow `Test<Level><Description>` pattern

## Package Navigation Guide

### Starting Points
- **Server startup**: `cmd/apiserver/app/server.go:Start()`
- **Handler examples**: `internal/apiserver/handler/user.go`
- **Wire setup**: `internal/apiserver/wire_gen.go:InitializeWebServer()`
- **Error handling**: `pkg/errorsx/builder.go:NewCode()`
- **Database setup**: `pkg/db/mysql.go:NewMySQL()`

### Configuration Patterns
- **Feature flags**: `pkg/feature/` - Toggle features per environment
- **Options pattern**: `pkg/options/` - Consistent component configuration
- **Environment overrides**: Viper handles env var to struct mapping automatically

## Infrastructure Templates

### Docker Services Ready to Use
- **Redis**: `make run-redis` or `docker-compose -f deployments/redis/docker-compose.yml up`
- **Jaeger**: `make run-jaeger` or `docker-compose -f deployments/jaeger/docker-compose.yml up`
- **Kafka**: `make run-kafka` or `docker-compose -f deployments/kafka/docker-compose.yml up`
- **All services**: `make start-all` (starts Redis, Jaeger, Kafka simultaneously)

### Monitoring Endpoints
- **Health**: `GET /health`
- **Metrics**: `GET /metrics` (Prometheus)
- **Tracing**: Integrated with OpenTelemetry (configurable Jaeger endpoint)
- **Debug**: `GET /debug/pprof` (when enabled)

## Module Information
- **Go Version**: 1.24.0+ required
- **Module Path**: `github.com/costa92/go-protoc/v2`
- **Branch**: Currently on `v2` branch
- **Protobuf**: Uses buf.build for dependency management

## Project-Specific Tools
- **rename-project**: Change module path with `make rename-project OLD_PATH=X NEW_PATH=Y`
- **githooks**: Git hooks installed automatically: githooks/{pre-commit,commit-msg,pre-push}

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.