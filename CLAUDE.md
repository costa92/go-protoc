# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go microservice framework built on Kratos v2 with Protocol Buffers, featuring error handling, internationalization, authentication, and observability. The project follows a modular architecture with clear separation of concerns.

## Development Commands

### Build and Run
- `make build` - Build the API server binary
- `make run-api` - Run the development server 
- `go run cmd/apiserver/main.go` - Alternative way to run the server

### Code Generation
- `make generate` - Generate Protocol Buffer code using buf
- `buf generate` - Direct buf command for protobuf generation
- `make wire` - Generate dependency injection code using Wire

### Code Quality
- `make fmt` - Format Go code
- `go mod tidy` - Clean up Go modules
- `make tidy` - Alias for go mod tidy

### Tools and Installation
- `make install-tools` - Install CI-related tools
- `make install-tools A=1` - Install all tools including optional ones

### Project Management
- `make rename-project OLD_PATH=<old> NEW_PATH=<new>` - Rename the project module path across all files

## Architecture

### Core Structure
- `cmd/apiserver/` - Main application entry point
- `internal/apiserver/` - Private application logic (business logic, handlers, stores)
- `pkg/` - Public packages that can be imported by other projects
- `api/` - OpenAPI/Swagger generated documentation
- `third_party/protobuf/` - Third-party protobuf definitions

### Key Packages
- `pkg/errorsx/` - Enhanced error handling with HTTP/gRPC status code mapping and i18n support
- `pkg/i18n/` - Internationalization support with context-based locale detection
- `pkg/authn/` - JWT-based authentication with configurable claims
- `pkg/server/` - HTTP and gRPC server implementations with middleware support
- `pkg/options/` - Configuration options for various components (databases, messaging, etc.)

### Protocol Buffers
- Uses buf for protobuf management and generation
- Custom error code generation via protoc-gen-go-errors-code
- API definitions in `pkg/api/apiserver/v1/`
- Generated OpenAPI documentation in `api/openapi/`

### Configuration
- YAML-based configuration in `configs/`
- Viper for configuration management
- Support for multiple environments and feature gates

### Dependencies
- Kratos v2 framework for microservices
- Wire for dependency injection
- GORM for database operations (MySQL/PostgreSQL)
- Redis for caching
- Consul/etcd for service registry
- OpenTelemetry for observability
- Prometheus for metrics

## Testing and Quality
No specific test commands are defined in the Makefile. Use standard Go testing commands:
- `go test ./...` - Run all tests
- `go test -v ./pkg/errorsx/` - Run tests for specific package

## Module Information
- Module path: `github.com/costa92/go-protoc/v2`
- Go version: 1.24.0
- Currently on branch: v2