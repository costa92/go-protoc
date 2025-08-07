
# GEMINI.md

## Project Overview

This project is a Go-based API server built with the Kratos framework. It uses Protocol Buffers for API definitions and includes features like JWT authentication, internationalization (i18n), and database integration with MySQL. The project is structured to support both gRPC and HTTP services.

### Key Technologies:

*   **Framework:** Go, Kratos
*   **API:** gRPC, HTTP, Protocol Buffers
*   **Authentication:** JWT
*   **Database:** MySQL
*   **Tooling:** Buf, Make, Cobra

## Building and Running

### Prerequisites

*   Go
*   Make
*   Buf

### Installation

Install the necessary tools using the Makefile:

```bash
make install-tools
```

### Running the API Server

To run the API server locally:

```bash
make run-api
```

### Building the Binary

To build the API server binary:

```bash
make build
```

The binary will be located in the `bin/` directory.

### Generating Code

To generate code from the Protocol Buffer definitions:

```bash
make generate
```

## Development Conventions

*   **Code Generation:** Protocol Buffer definitions are managed with Buf and generated using `make generate`.
*   **Dependency Management:** Go modules are used for dependency management. Use `go mod tidy` to keep dependencies clean.
*   **Configuration:** The application is configured using YAML files in the `configs/` directory.
*   **Testing:** (TODO: Add instructions on how to run tests if available)
*   **Linting:** (TODO: Add instructions on how to run the linter if available)
