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

Generated projects include comprehensive tests:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific service tests
go test ./internal/services/pet/
```

## 🌟 Current Status

### ✅ **Completed Features**
- ✅ OpenAPI 3.0 parsing and validation
- ✅ Go type generation from schemas
- ✅ MongoDB repository generation
- ✅ HTTP handler generation with Chi router
- ✅ Complete project scaffolding
- ✅ Environment configuration
- ✅ Unit test generation
- ✅ Clean compilation without warnings
- ✅ Proper error handling and domain errors

### 🚧 **In Development**
- 🔄 Request/response validation utilities
- 🔄 Middleware support
- 🔄 Authentication and authorization
- 🔄 Database migrations
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

# Run tests
go test ./...

# Test code generation
go run cmd/goapigen/main.go --spec examples/petstore/openapi.yaml --output test --init --types --mongo --http
```

## 📚 Examples

Check out the `examples/` directory for:
- **Petstore API** - Complete REST API example
- **Advanced schemas** - Complex data models
- **Custom configurations** - Environment setups

## 🛠️ Built With

- **Go 1.21+** - Core language
- **Chi Router** - HTTP routing
- **MongoDB Driver** - Database integration
- **OpenAPI 3.0** - API specifications
- **Testify** - Testing framework

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