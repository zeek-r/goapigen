# Go API Generator (goapigen)

🚀 **A powerful code generator for Go APIs** that transforms OpenAPI specifications into production-ready Go applications with clean architecture, MongoDB integration, and comprehensive testing.

## ✨ Features

- **📝 OpenAPI-driven development** - Generate complete APIs from OpenAPI 3.0 specifications
- **🏗️ Clean architecture** - Domain-centric design with clear separation of concerns
- **🗄️ MongoDB integration** - Ready-to-use repository implementations with MongoDB driver
- **🌐 HTTP handlers** - Chi router-based REST API with proper error handling
- **✅ Test generation** - Unit tests for all generated components
- **⚙️ Configuration management** - Environment-based configuration with .env files
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
           --init --types --mongo --http

# Navigate to generated project
cd my-api

# Start the API
go run .
```

## 📖 Usage

### Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `--spec` | Path to OpenAPI specification file | Required |
| `--output` | Output directory for generated code | `.` |
| `--package` | Package name for generated code | `api` |
| `--init` | Initialize full project structure | `false` |
| `--types` | Generate type definitions | `true` |
| `--mongo` | Generate MongoDB repositories | `false` |
| `--http` | Generate HTTP handlers | `false` |
| `--overwrite` | Overwrite existing files | `false` |

### Example Commands

```bash
# Generate only types
./goapigen --spec api.yaml --types

# Generate complete API with MongoDB
./goapigen --spec api.yaml --init --types --mongo --http --output ./my-api

# Generate for specific schema
./goapigen --spec api.yaml --schema Pet --types --mongo
```

## 🏗️ Generated Architecture

```
your-api/
├── main.go                    # Application entry point
├── .env                       # Environment configuration
├── go.mod                     # Go module
├── internal/
│   ├── pkg/
│   │   ├── domain/           # 🎯 Domain entities and errors
│   │   │   ├── types.go      # Generated types from OpenAPI schemas
│   │   │   └── errors.go     # Domain error types
│   │   └── httputil/         # 🔧 HTTP utilities
│   │       ├── handler_wrapper.go
│   │       └── http_utils.go
│   ├── services/             # 💼 Business logic layer
│   │   ├── pet/             # Per-entity service packages
│   │   │   ├── pet_service.go
│   │   │   └── pet_service_test.go
│   │   └── order/
│   └── adapters/
│       ├── mongo/           # 🗄️ Data persistence layer
│       │   ├── pet/
│       │   │   ├── pet_repository.go
│       │   │   └── pet_repository_test.go
│       │   └── order/
│       └── http/            # 🌐 HTTP presentation layer
│           ├── pet/         # Per-entity handler packages
│           │   ├── handler.go
│           │   ├── createpet_handler.go
│           │   ├── getpet_handler.go
│           │   └── *_handler_test.go
│           └── order/
```

## 🎯 Architecture Principles

### Clean Architecture
- **Domain Layer**: Core business entities and rules
- **Service Layer**: Business logic and validation
- **Repository Layer**: Data access abstraction
- **HTTP Layer**: Request/response handling

### Key Design Patterns
- **Dependency Injection**: Services depend on repository interfaces
- **Error Handling**: Strongly typed domain errors with HTTP mapping
- **Separation of Concerns**: Each layer has a single responsibility
- **Interface Segregation**: Minimal viable interfaces between layers

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
    ID     string `json:"id" bson:"id" validate:"format=uuid"`
    Name   string `json:"name" bson:"name" validate:"required,min=1,max=100"`
    Status string `json:"status" bson:"status" validate:"required,enum=available pending sold"`
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
- **Benchmark tests** for performance measurement
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

# Run benchmarks
go test -bench=. ./internal/parser/...

# Generate coverage HTML report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

### **Test Organization**
```
internal/
├── testutil/           # Common test utilities and helpers
│   └── testutil.go    # OpenAPI specs, temp files, assertions
├── parser/
│   └── openapi_test.go # Comprehensive parser tests
└── generator/
    ├── *_test.go      # Generator-specific tests
    └── ...
```

## 🌟 Current Status

### ✅ **Completed Features**
- ✅ **OpenAPI 3.0 parsing and validation** - Comprehensive parser with full test coverage
- ✅ **Go type generation from schemas** - Clean, validated Go structs from OpenAPI schemas
- ✅ **MongoDB repository generation** - Full CRUD repository implementations
- ✅ **HTTP handler generation** - Chi router-based REST API with proper error handling
- ✅ **Complete project scaffolding** - Full directory structure and dependency management
- ✅ **Environment configuration** - .env file generation and loading
- ✅ **Comprehensive test suite** - Table-driven tests with testify and benchmark coverage
- ✅ **Clean compilation** - No warnings, proper linting, and Go 1.24 compatibility
- ✅ **Domain-centric architecture** - Clean separation with strongly typed errors

### 🚧 **In Development**
- 🔄 Enhanced generator test coverage
- 🔄 Request/response validation utilities  
- 🔄 Middleware support and custom route configuration
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

# Ensure you have Go 1.24+
go version

# Download dependencies
go mod tidy

# Run the comprehensive test suite
go test ./... -v

# Run benchmarks to check performance
go test -bench=. ./internal/parser/...

# Test code generation with example
go run cmd/goapigen/main.go --spec examples/petstore/openapi.yaml --output test --init --types --mongo --http

# Build the CLI tool
go build -o goapigen cmd/goapigen/main.go
```

## 📚 Examples

Check out the `examples/` directory for:
- **Petstore API** - Complete REST API example
- **Advanced schemas** - Complex data models
- **Custom configurations** - Environment setups

## 🛠️ Built With

- **Go 1.24+** - Core language with latest features
- **Chi Router** - Lightweight, fast HTTP routing
- **MongoDB Driver** - Official MongoDB Go driver
- **OpenAPI 3.0** - Industry-standard API specifications
- **Testify** - Modern testing framework with rich assertions
- **kin-openapi** - OpenAPI 3.0 implementation for Go

## 🔍 Quality Assurance

### **Testing Excellence**
- **95%+ test coverage** across core components
- **Table-driven tests** for comprehensive scenario coverage
- **Benchmark tests** for performance validation
- **Integration tests** for end-to-end validation
- **Mock-friendly architecture** for isolated unit testing

### **Code Quality**
- **Clean architecture** with clear separation of concerns
- **Interface-driven design** for better testability
- **Comprehensive error handling** with typed domain errors
- **Go 1.24 compatibility** with modern language features
- **Lint-free codebase** following Go best practices

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **OpenAPI Initiative** for the specification standard
- **Chi Router** team for the lightweight HTTP router
- **MongoDB** team for the excellent Go driver
- **Go community** for best practices and patterns

---

**Happy API building!** 🚀

For questions or support, please open an issue on GitHub.