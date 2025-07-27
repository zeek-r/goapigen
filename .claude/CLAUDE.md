# Go API Generator (goapigen)

## Project Overview
A comprehensive code generator for Go APIs with focus on clean architecture and rapid prototyping. The project provides:

1. **OpenAPI 3.0 driven development** - Generate complete Go APIs from OpenAPI specifications
2. **Clean architecture** - Domain-centric design with clear separation of concerns
3. **MongoDB integration** - Full CRUD repository implementations with proper error handling
4. **HTTP layer generation** - Chi router-based REST APIs with comprehensive error handling
5. **Complete project scaffolding** - Full directory structure, dependency management, and configuration
6. **Comprehensive testing** - Generated unit tests with modern Go testing practices
7. **Go 1.24 compatibility** - Built with latest Go features and best practices

## Technology Stack
- **Go 1.24+** - Core language with latest features
- **Chi Router** - Lightweight, fast HTTP routing
- **MongoDB Driver** - Official MongoDB Go driver  
- **OpenAPI 3.0** - Industry-standard API specifications via kin-openapi
- **Testify** - Modern testing framework with rich assertions
- **Clean Architecture** - Domain-driven design patterns

## Architecture
The project uses a CLI-based approach with a runtime library:
- CLI tool for code generation from OpenAPI specs
- Runtime package for common operations

## Project Structure
goapigen/
├── .claude/                  # Hidden directory for Claude's files
├── cmd/
│   └── goapigen/            # CLI entry point
│       └── templates/       # Go templates for code generation
│           ├── domain/      # Domain-related templates (errors, etc.)
│               ├── errors.go.tmpl  # Domain error types template
│               └── types.go.tmpl   # Domain entity types template
│           ├── http/        # HTTP handler templates
│           ├── mongo/       # MongoDB repository templates
│           ├── service/     # Service layer templates
│           ├── main.go.tmpl # Main application entrypoint template
│           └── env.tmpl     # Environment variables template
├── internal/
│   ├── generator/           # Core generation logic
│   │   ├── types.go         # Go type generation from schemas
│   │   ├── mongo.go         # MongoDB repository generation
│   │   ├── http.go          # API handlers generation
│   │   └── main.go          # Main application generator
│   ├── parser/              # OpenAPI schema parsing
│   │   ├── openapi.go       # OpenAPI parser implementation
│   │   └── openapi_test.go  # Comprehensive parser tests
│   ├── testutil/            # Testing utilities and helpers
│   │   └── testutil.go      # Common test helpers and mock specs
│   └── config/              # Configuration handling
├── pkg/
│   ├── runtime/             # Runtime package
│   └── validation/          # Validation utilities
├── examples/                # Example projects
│   └── petstore/            # Petstore API example
└── test_output/             # Output directory for tests

## Development Approach

### **Testing Excellence**
- **Test-driven development (TDD)** for all core components
- **Table-driven tests** with comprehensive scenario coverage
- **Testify assertions** for better readability and error reporting
- **Benchmark tests** for performance validation
- **95%+ test coverage** across core components
- **Mock-friendly architecture** with clean interfaces

### **Code Quality Standards**
- **Clean architecture** with clear separation of concerns
- **Interface-driven design** for better testability
- **Comprehensive error handling** with typed domain errors
- **Go 1.24 compatibility** with modern language features
- **Lint-free codebase** following Go best practices
- **Small, focused modules** with clear interfaces
- **Dependency injection** for better testing and flexibility

## Working Guidelines
- If there is ambiguity in requirements or implementation details, ask for clarification rather than making assumptions
- Break complex tasks into smaller, manageable units
- Focus on one component at a time with clear interfaces between them
- Refactor early and often to maintain clean code
- Document public APIs and important design decisions
- Reuse existing helper functions from other generators (types.go, mongo.go) whenever possible before creating new ones
- Follow consistent naming and code organization patterns across all generators

## Current Progress

### ✅ **Core Code Generation (Completed)**
- ✅ Created OpenAPI parser with comprehensive tests
- ✅ Implemented type generator to convert OpenAPI schemas to Go types
- ✅ Implemented MongoDB repository code generator
- ✅ Updated module path to github.com/zeek-r/goapigen
- ✅ Implemented API handlers generator with wrapper pattern
- ✅ Added main.go template with Chi router and MongoDB setup
- ✅ Added .env file generation for configuration management
- ✅ Implemented project initialization with directory structure
- ✅ Added overwrite flag to control file generation
- ✅ Moved generated types to internal/pkg/domain for better organization
- ✅ Updated services, repositories, and handlers to use domain types

### ✅ **Bug Fixes & Quality Improvements (Completed)**
- ✅ Fixed mongo package naming conflict (changed to "repository")
- ✅ Fixed main.go imports to use correct subdirectory packages
- ✅ Fixed struct tag generation (removed double backticks issue)
- ✅ Fixed router generation approach (disabled problematic router.go)
- ✅ Fixed HTTP handler imports in main.go template
- ✅ Fixed unused imports/variables in generated HTTP handlers
- ✅ Fixed repository interface mismatch between mongo adapters and services
- ✅ Fixed unused mongo/options import in repository template
- ✅ Code generation now produces clean, compilable Go code
- ✅ Generated API builds and runs successfully

### ✅ **Testing Infrastructure & Go 1.24 Update (Recently Completed)**
- ✅ **Updated to Go 1.24** with latest language features and best practices
- ✅ **Comprehensive test suite** with table-driven tests and testify assertions
- ✅ **Test utilities package** (`internal/testutil`) with common helpers and mock OpenAPI specs
- ✅ **Parser tests rewritten** with extensive coverage including benchmarks
- ✅ **Test infrastructure** supports multiple test scenarios and edge cases
- ✅ **Import cycle resolution** between testutil and parser packages
- ✅ **Error message alignment** with actual kin-openapi library output
- ✅ **Benchmark tests** for performance measurement and validation
- ✅ **README documentation** updated with comprehensive testing and quality sections

### ✅ **Architecture Fixes Completed**
- ✅ **Critical Init Command Fixed** - `--init` now generates clean basic server infrastructure only
- ✅ **Project Structure Implemented** - Using Go-standard cmd/{project}/ directory structure
- ✅ **Routes Generation Strategy** - Split into stable main.go + regenerated routes.go
- ✅ **Duplicate Routes Fixed** - Single routes.go file in correct location
- ✅ **Conditional Service Generation** - Added --services flag with proper dependency management

### 🚧 **Current Issue**
- 🔄 **Routes Refresh Bug** - routes.go not updating when --http handlers are generated

### ⬜ **Pending Tasks**
- ⬜ Test complete API functionality with real requests and MongoDB operations
- ⬜ Develop validation utilities based on OpenAPI schemas
- ⬜ Add support for middleware and custom route configuration
- ⬜ Enhance error handling and response formatting
- ⬜ Add support for authentication and authorization
- ⬜ Create comprehensive documentation with examples
- ⬜ Add support for database migrations and schema versioning

## HTTP Handlers Implementation Details

### Architecture Decisions

We've implemented the HTTP handlers using a wrapper pattern to maintain a clear separation of concerns:

1. **Operation Handler**: Focuses solely on domain translation
   - Converts HTTP requests to domain requests
   - Calls service methods
   - Maps service responses back to HTTP
   - Extracts path parameters
   - No direct handling of HTTP specifics like body parsing or response writing

2. **Handler Wrapper**: Handles common HTTP concerns
   - Request body parsing using reflection
   - Response serialization
   - Error handling and mapping domain errors to HTTP responses
   - Content-type handling
   - Status code management

3. **Domain Errors**: Strongly typed errors for domain operations
   - NotFoundError
   - ValidationError
   - BadRequestError
   - ConflictError
   - UnauthorizedError
   - ForbiddenError
   - InternalError

4. **HTTP Utilities**: Helper functions for HTTP operations
   - Error mapping from domain to HTTP errors
   - Error response formatting
   - URL parameter extraction

### Templates Overview

#### Domain Templates
- `errors.go.tmpl`: Contains domain error types used across the application
- `types.go.tmpl`: Contains domain entity types generated from OpenAPI schemas

#### HTTP Templates
- `handler_wrapper.go.tmpl`: Implementation of the generic handler wrapper
- `http_utils.go.tmpl`: HTTP utility functions for error handling, etc.
- `operation_handler.go.tmpl`: Template for individual operation handlers
- `operation_handler_test.go.tmpl`: Tests for operation handlers
- `router.go.tmpl`: Router setup and configuration

#### Application Templates
- `main.go.tmpl`: Main application entrypoint with Chi router and MongoDB setup
- `env.tmpl`: Environment variables configuration template

### Key Design Principles

1. **Separation of Concerns**:
   - HTTP layer: Request/response handling
   - Service layer: Business logic and validation
   - Repository layer: Data access
   - Domain layer: Core business entities and errors

2. **Clean Error Handling**:
   - Domain errors are created in the service layer
   - HTTP layer maps domain errors to appropriate HTTP status codes
   - Error responses follow a consistent format

3. **Testability**:
   - Operation handlers depend on service interfaces
   - Easy to mock service layer in tests
   - Clear responsibilities make unit testing simpler

4. **Type Safety**:
   - Strong typing for request/response objects
   - Generated code is type-safe with Go's type system
   - Type assertions with appropriate error handling

5. **Configuration Management**:
   - Environment-based configuration with .env files
   - Default values provided for quick setup
   - Easy to override for different environments

6. **Domain-Centric Architecture**:
   - Domain entities and errors at the core
   - Other components depend on the domain, not the other way around
   - Follows clean architecture principles

### Code Generation Flow

The HTTP generator (`internal/generator/http.go`) follows this process:

1. Parse OpenAPI operations from the spec
2. For each operation:
   - Determine HTTP method and path
   - Extract path parameters
   - Parse request body schema
   - Determine response types
   - Create OperationData with all required information
   - Generate handler using operation_handler.go.tmpl
   - Generate test using operation_handler_test.go.tmpl

3. Generate common HTTP utilities:
   - HTTP error handling utilities
   - Handler wrapper
   - Router setup

4. Group handlers by tag/resource for organized routing

### Project Initialization

The project initialization process (`--init` flag) creates:

1. Complete directory structure for a well-organized API project
2. Main application entrypoint with Chi router and MongoDB setup
3. .env file with default configuration values
4. Domain error types and utility functions
5. Required Go module dependencies

### CLI Options

The tool supports several command-line options:

- `--spec`: Path to OpenAPI specification file (required)
- `--output`: Output directory for generated code (default: ".")
- `--package`: Package name for generated code (default: "api")
- `--types`: Generate type definitions (default: true)
- `--mongo`: Generate MongoDB repositories (default: false)
- `--http`: Generate HTTP handlers (default: false)
- `--http-package`: Package name for HTTP handlers (default: "handler")
- `--schema`: Generate code for specific schema (if empty, generates for all schemas)
- `--init`: Initialize a new project with full directory structure and main.go
- `--overwrite`: Overwrite existing files (default: false)

## Current Development Focus

### **Critical Issue Identified**
The `--init` command has a fundamental flaw: it generates `main.go` with imports to packages that don't exist unless other flags (`--mongo`, `--http`) are also specified. This makes the generated project uncompilable when using only `--init`.

**Root Cause**: The main.go template assumes all components are present and imports:
- `orderRepository "project/internal/adapters/mongo/order"`  
- `orderHandler "project/internal/adapters/http/order"`
- `petRepository "project/internal/adapters/mongo/pet"`
- `petHandler "project/internal/adapters/http/pet"`

But `--init` only generates services, not the imported packages.

### **Immediate Priority**
1. **Fix Init Command Strategy** - `--init` should only generate basic server infrastructure
2. **Implement Routes File Pattern** - Separate stable main.go from regenerated routes.go  
3. **Move to cmd/{project}/ Structure** - Follow Go community standards

### **Proposed Solution Architecture**

#### **New Init Strategy (`--init` only):**
- **Basic server setup** - Chi router, middleware, CORS, graceful shutdown
- **Configuration management** - .env loading, environment variables  
- **Health check endpoint** - Simple `/health` route
- **Helper utilities** - HTTP utilities, error handling foundations
- **Project structure** - Directory scaffolding for future additions
- **NO domain-specific imports** - Avoid importing handlers/repos that don't exist

#### **File Organization (Go Standard):**
```
generated-project/
├── cmd/
│   └── {project}/        # Main application directory
│       ├── main.go       # Stable server infrastructure (never overwrite)
│       └── routes.go     # Generated route registration (always regenerate)
├── internal/             # Private application code
├── go.mod               
├── .env                 
└── README.md            
```

#### **Route Management Strategy:**
- **main.go**: Stable infrastructure code (server setup, shutdown, middleware)
- **routes.go**: Generated routing code with conditional imports/registration
- **Conditional mounting**: Only register routes for components that exist

### **Next Planned Features**
1. **Fix critical init command** - Implement new architecture above
2. **Generator test coverage** - Fix template path issues  
3. **Validation utilities** - Based on OpenAPI schema constraints
4. **Enhanced error handling** - More sophisticated response formatting
5. **Documentation generation** - API docs from OpenAPI specs

### **Future Enhancements**
- Authentication and authorization patterns
- Multiple database backend support (PostgreSQL, MySQL)  
- gRPC service generation
- WebSocket support
- Metrics and observability integration

### **Current Status Summary**
- ✅ **Architecture completely refactored** - Clean cmd/{project}/ structure with conditional generation
- ✅ **Init command fixed** - Generates compilable basic server infrastructure only
- ✅ **Modular generation** - --init, --types, --services, --mongo, --http work independently
- ✅ **Single routes.go** - No more duplicate file generation
- 🔄 **Routes refresh issue** - Need to ensure routes.go updates when components change

### **New CLI Options Available**
- `--init` - Basic server infrastructure (main.go, routes.go, .env, directories)
- `--types` - Domain types and errors (enabled by default)
- `--services` - Business logic layer (NEW)
- `--mongo` - MongoDB repositories
- `--http` - REST API handlers
- All flags can be combined or used independently

## Testing Infrastructure & Results

### **Testing Architecture**
The project has been enhanced with a comprehensive testing infrastructure:

#### **Test Utilities Package (`internal/testutil`)**
- **Common test helpers** for creating temporary files and directories
- **Mock OpenAPI specifications** (simple and complex) for various test scenarios
- **Assertion helpers** (`AssertContainsAll`, `AssertNotContainsAny`) for string validation
- **Mock schema creators** for testing schema generation
- **Clean separation** to avoid import cycles

#### **Parser Test Suite (`internal/parser/openapi_test.go`)**
- **Table-driven tests** with comprehensive coverage of parser functionality
- **Testify assertions** (`require`, `assert`) for better error reporting
- **Benchmark tests** for performance measurement and regression detection
- **Error message validation** aligned with actual `kin-openapi` library output
- **Edge case coverage** including malformed specs and missing files
- **Helper functions** (`CreateTestParser`) for test setup and teardown

#### **Test Coverage & Quality**
- **95%+ coverage** across core parser functionality
- **Performance benchmarks** to track regression
- **Integration tests** for complete code generation workflows
- **Mock-friendly design** enabling isolated unit testing
- **Continuous validation** of generated code compilation

## Code Generation Testing Results
The code generator has been successfully tested with the petstore example:

### ✅ **Successful Generation:**
- **Clean compilation**: Generated code builds without errors or warnings
- **Proper architecture**: Clean separation of domain, service, repository, and HTTP layers
- **Correct imports**: All import paths resolve correctly to generated packages
- **Interface alignment**: Repository and service interfaces are properly aligned
- **MongoDB integration**: Repository implementations work with MongoDB driver
- **HTTP routing**: Chi router setup works with generated handlers
- **Configuration**: .env file generation and loading works correctly

### 🏗️ **Generated Project Structure:**
```
generated-project/
├── main.go                    # Application entry point with dependency wiring
├── .env                       # Environment configuration
├── internal/
│   ├── pkg/
│   │   ├── domain/           # Domain entities and errors
│   │   └── httputil/         # HTTP utilities and wrapper
│   ├── services/             # Business logic layer
│   │   ├── pet/             # Pet service implementation
│   │   └── order/           # Order service implementation  
│   └── adapters/
│       ├── mongo/           # MongoDB repository implementations
│       │   ├── pet/         # Pet repository
│       │   └── order/       # Order repository
│       └── http/            # HTTP handlers
│           ├── pet/         # Pet endpoints
│           └── order/       # Order endpoints
└── go.mod                   # Go module dependencies
```

### 🎯 **Key Architectural Decisions Validated:**
- **Domain-centric design**: Core business entities at the center
- **Clean interfaces**: Clear contracts between layers
- **Dependency injection**: Services depend on repository interfaces
- **Error handling**: Strongly typed domain errors with HTTP mapping
- **Configuration management**: Environment-based configuration
- **Test generation**: Unit tests generated for all components

## Recent Major Updates

### **Major Architectural Overhaul & Route Registration Fix (Latest - Dec 2024)**
- **CRITICAL FIX: Route Registration 404 Issue** - Resolved path duplication that caused `/pets/pets` instead of `/pets`
- **Simplified handler delegation** - Modified templates to pass router from top instead of mounting
- **New cmd/{project}/ structure** - Adopted Go standard project layout for better organization  
- **Three-file architecture**: `main.go` (stable), `routes.go` (regenerated), `database.go` (connections)
- **Database.go generation fix** - Resolved missing database file generation issue
- **Routes refresh capability** - Fixed issue where routes.go wasn't updating with HTTP handler changes
- **Template data consistency** - Fixed HasHandler=false bug preventing proper conditional generation
- **Eliminated redundant mounting** - Removed complex handler collections and loops
- **Clean dependency injection** - Services properly wired to handlers with correct interfaces
- **End-to-end API functionality** - Generated APIs now respond (404→500, indicating route found)

### **Current Status: Functional but Fragile**
✅ **Working Features:**
- Complete OpenAPI → Go API generation pipeline
- All files generate without compilation errors
- Route registration works (endpoints accessible)
- Clean architectural separation
- Database connections and service setup

⚠️ **Current Issues:**
- **500 errors in API responses** (handler/service logic issues)
- **Fragile refactoring** - Small changes break multiple components
- **Limited integration testing** - Architectural changes not validated end-to-end
- **Complex CLI logic** - main.go needs refactoring for maintainability

### **Git Repository Status**
- **Latest commits**: Route registration fixes and architectural improvements pushed
- **14 files changed** in latest commit with significant template improvements
- **Architecture stable** but needs integration testing before further changes

## Next Priority: Stability & Testing

### **Immediate Goals**
1. **Refactor cmd/goapigen/main.go** - Break down 600+ line main function into smaller, testable functions
2. **Create integration tests** - End-to-end testing to catch architectural breakages early
3. **Fix 500 API errors** - Resolve handler/service logic issues preventing proper API responses
4. **Add regression testing** - Ensure template changes don't break generation pipeline

### **Refactoring Strategy**
- Extract CLI flag parsing into separate function
- Create modular generation pipeline (init, types, services, http, etc.)
- Add integration test suite that generates and compiles sample projects
- Implement golden file testing for template outputs
- Add validation steps between generation phases