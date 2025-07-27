# Go API Generator (goapigen)

🚀 **A powerful code generator for Go APIs** that transforms OpenAPI specifications into production-ready Go applications with clean architecture, MongoDB integration, context-aware logging, and comprehensive configuration management.

## ✨ Features

- **📝 OpenAPI-driven development** - Generate complete APIs from OpenAPI 3.0 specifications
- **🏗️ Clean architecture** - Domain-centric design with clear separation of concerns
- **🗄️ MongoDB integration** - Ready-to-use repository implementations with MongoDB driver
- **🌐 HTTP handlers** - Chi router-based REST API with proper error handling
- **✅ Test generation** - Unit tests for all generated components
- **⚙️ Configuration management** - Environment-based configuration with envconfig
- **📋 Context-aware logging** - Structured logging with zapctxd and field propagation
- **🔒 Type safety** - Strongly typed Go code with proper validation tags
- **📦 Project scaffolding** - Complete project structure with dependencies

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/zeek-r/goapigen.git
cd goapigen

# Build the CLI tool
go build -o goapigen cmd/goapigen/main.go
```

### Basic Usage

```bash
# Generate a complete API project
./goapigen --spec examples/petstore/openapi.yaml \
           --output ./my-api \
           --init --services --mongo --http

# Navigate to generated project
cd my-api

# Start the API
go run cmd/my-api/*.go
```

## 📖 Usage

### Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `--spec` | Path to OpenAPI specification file | Required |
| `--output` | Output directory for generated code | `.` |
| `--package` | Package name for generated code | `api` |
| `--init` | Initialize full project structure with config and logging | `false` |
| `--types` | Generate type definitions | `true` |
| `--services` | Generate service layer | `false` |
| `--mongo` | Generate MongoDB repositories | `false` |
| `--http` | Generate HTTP handlers | `false` |
| `--overwrite` | Overwrite existing files | `false` |
| `--schema` | Generate code for specific schema only | All schemas |

### Example Commands

```bash
# Generate only types
./goapigen --spec api.yaml --types

# Generate complete API with MongoDB and services
./goapigen --spec api.yaml --init --services --mongo --http --output ./my-api

# Generate for specific schema
./goapigen --spec api.yaml --schema Pet --types --mongo
```

## 🏗️ Generated Architecture

```
your-api/
├── cmd/
│   └── your-api/              # Application entry point
│       ├── main.go           # Server startup
│       ├── routes.go         # Route registration
│       └── database.go       # Database setup
├── .env                      # Environment configuration
├── go.mod                    # Go module dependencies
└── internal/
    ├── pkg/
    │   ├── config/           # 🔧 Configuration management
    │   │   ├── config.go     # envconfig-based configuration
    │   │   └── config_test.go
    │   ├── logger/           # 📋 Context-aware logging
    │   │   ├── logger.go     # zapctxd logger integration
    │   │   └── logger_test.go
    │   └── domain/           # 🎯 Domain entities and errors
    │       ├── types.go      # Generated types from OpenAPI schemas
    │       └── errors.go     # Domain error types
    ├── services/             # 💼 Business logic layer
    │   ├── pet/             # Per-entity service packages
    │   │   ├── pet_service.go
    │   │   └── pet_service_test.go
    │   └── order/
    └── adapters/
        ├── repository/      # 🗄️ Data persistence layer
        │   ├── pet/
        │   │   ├── pet_repository.go
        │   │   └── pet_repository_test.go
        │   └── order/
        └── http/            # 🌐 HTTP presentation layer
            ├── pet/         # Per-entity handler packages
            │   ├── handler.go
            │   └── *_handler_test.go
            └── order/
```

## 🎯 Architecture Principles

### Clean Architecture
- **Domain Layer**: Core business entities and rules
- **Service Layer**: Business logic and validation  
- **Repository Layer**: Data access abstraction
- **HTTP Layer**: Request/response handling
- **Configuration Layer**: Environment-based configuration management
- **Logging Layer**: Context-aware structured logging

### Key Design Patterns
- **Dependency Injection**: Services depend on repository interfaces
- **Error Handling**: Strongly typed domain errors with HTTP mapping
- **Separation of Concerns**: Each layer has a single responsibility
- **Interface Segregation**: Minimal viable interfaces between layers
- **Configuration Management**: Struct-based environment variable processing
- **Context-aware Logging**: Automatic field propagation through request contexts

## 📋 Configuration Management

The generated projects use `envconfig` for clean, struct-based configuration:

```go
type Config struct {
    Server   ServerConfig   `envconfig:"SERVER"`
    Database DatabaseConfig `envconfig:"DATABASE"`
    Logging  LoggingConfig  `envconfig:"LOGGING"`
}

type ServerConfig struct {
    Port string `envconfig:"PORT" default:"8080"`
    Host string `envconfig:"HOST" default:"localhost"`
}
```

### Environment Variables
- `PORT` - Server port (default: 8080)
- `HOST` - Server host (default: localhost)
- `MONGO_URI` - MongoDB connection string (default: mongodb://localhost:27017)
- `DB_NAME` - Database name (default: project name)
- `LOG_LEVEL` - Logging level: debug, info, warn, error (default: info)
- `LOG_DEVELOPMENT` - Development mode logging (default: false)
- `LOG_FORMAT` - Log format: json, console (default: json)

## 📋 Context-aware Logging

The generated projects use `zapctxd` for structured, context-aware logging:

```go
// Initialize logger from configuration
logger := logger.NewFromEnv()

// Add fields to context
ctx = ctxd.AddFields(ctx, "user_id", "123", "request_id", "abc")

// Log with automatic field propagation
logger.Info(ctx, "processing request") 
// Output: {"level":"info","msg":"processing request","user_id":"123","request_id":"abc"}
```

## 📋 Example OpenAPI Schema

```yaml
openapi: 3.0.0
info:
  title: Pet Store API
  version: 1.0.0

components:
  schemas:
    Pet:
      type: object
      required: [name, status]
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
          minLength: 1
          maxLength: 100
        status:
          type: string
          enum: [available, pending, sold]

paths:
  /pets:
    get:
      operationId: listPets
      responses:
        '200':
          description: List of pets
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Pet'
```

## 🔧 Generated Code Features

### Type Safety
```go
// Generated domain types with validation tags
type Pet struct {
    ID     string `json:"id" bson:"id"`
    Name   string `json:"name" bson:"name"`
    Status string `json:"status" bson:"status"`
}
```

### Clean Service Layer
```go
type PetService interface {
    Create(ctx context.Context, request PetCreateRequest) (domain.Pet, error)
    GetByID(ctx context.Context, id string) (domain.Pet, error)
    List(ctx context.Context) ([]domain.Pet, error)
}
```

### Repository Pattern
```go
type PetRepository interface {
    Create(ctx context.Context, pet *domain.Pet) error
    GetByID(ctx context.Context, id string) (*domain.Pet, error)
    List(ctx context.Context) ([]*domain.Pet, error)
}
```

## 🧪 Testing

The project includes a comprehensive test suite with modern Go testing practices:

### **Test Infrastructure**
- **Table-driven tests** with comprehensive coverage
- **Testify assertions** for better readability and error reporting
- **Test utilities** package for common testing patterns
- **Mock-friendly design** with clean interfaces

### **Running Tests**
```bash
# Run all tests with verbose output
go test ./... -v

# Run tests with coverage report
go test -cover ./...

# Run specific package tests
go test ./internal/parser/... -v

# Generate coverage HTML report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### **Test Organization**
```
internal/
├── parser/
│   └── openapi_test.go # Comprehensive parser tests
└── generator/
    ├── *_test.go      # Generator-specific tests
    └── ...
test/
└── integration/       # End-to-end integration tests
    └── generation_test.go
```

## 🌟 Current Status

### ✅ **Completed Features**
- ✅ **OpenAPI 3.0 parsing and validation** - Comprehensive parser with full test coverage
- ✅ **Go type generation from schemas** - Clean Go structs from OpenAPI schemas  
- ✅ **MongoDB repository generation** - Full CRUD repository implementations
- ✅ **HTTP handler generation** - Chi router-based REST API with proper error handling
- ✅ **Service layer generation** - Business logic layer with clean interfaces
- ✅ **Complete project scaffolding** - Full cmd/{project}/ structure and dependency management
- ✅ **Environment configuration** - envconfig-based configuration management
- ✅ **Context-aware logging** - zapctxd integration with structured logging
- ✅ **Comprehensive test generation** - Unit tests for all generated components
- ✅ **Integration test suite** - End-to-end generator validation
- ✅ **Clean compilation** - No warnings, proper linting

### 🚧 **In Development**
- 🔄 Enhanced middleware support and custom route configuration
- 🔄 Authentication and authorization patterns
- 🔄 Database migrations and schema versioning
- 🔄 API documentation generation

## 🤝 Contributing

We welcome contributions! Here's how to get started:

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** and add tests
4. **Run tests**: `go test ./...`
5. **Commit your changes**: `git commit -m "Add amazing feature"`
6. **Push to branch**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

### Development Setup

```bash
# Clone and setup
git clone https://github.com/zeek-r/goapigen.git
cd goapigen

# Ensure you have Go 1.21+
go version

# Download dependencies
go mod tidy

# Run the comprehensive test suite
go test ./... -v

# Test code generation with example
go run cmd/goapigen/main.go --spec examples/petstore/openapi.yaml --output test --init --services --mongo --http

# Build the CLI tool
go build -o goapigen cmd/goapigen/main.go
```

## 📚 Examples

Check out the `examples/` directory for:
- **Petstore API** - Complete REST API example with OpenAPI specification

## 🛠️ Built With

- **Go 1.21+** - Core language
- **Chi Router** - Lightweight, fast HTTP routing
- **MongoDB Driver** - Official MongoDB Go driver
- **OpenAPI 3.0** - Industry-standard API specifications
- **zapctxd** - Context-aware structured logging
- **envconfig** - Environment variable configuration
- **Testify** - Modern testing framework with rich assertions
- **kin-openapi** - OpenAPI 3.0 implementation for Go

## 🔍 Quality Assurance

### **Testing Excellence**
- **Comprehensive test coverage** across core components
- **Table-driven tests** for comprehensive scenario coverage
- **Integration tests** for end-to-end validation
- **Mock-friendly architecture** for isolated unit testing

### **Code Quality**
- **Clean architecture** with clear separation of concerns
- **Interface-driven design** for better testability
- **Comprehensive error handling** with typed domain errors
- **Modern Go practices** with proper dependency management
- **Lint-free codebase** following Go best practices

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **OpenAPI Initiative** for the specification standard
- **Chi Router** team for the lightweight HTTP router
- **MongoDB** team for the excellent Go driver
- **zapctxd** team for context-aware logging
- **envconfig** team for clean configuration management
- **Go community** for best practices and patterns

---

**Happy API building!** 🚀

For questions or support, please open an issue on GitHub.