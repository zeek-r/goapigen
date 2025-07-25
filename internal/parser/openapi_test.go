package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zeek-r/goapigen/internal/testutil"
)

// CreateTestParser creates a parser from a spec string for testing
func CreateTestParser(t *testing.T, spec string) *OpenAPIParser {
	t.Helper()

	specFile := testutil.CreateTempFile(t, "test-spec.yaml", spec)
	parser, err := NewOpenAPIParser(specFile)
	require.NoError(t, err)

	return parser
}

func TestNewOpenAPIParser(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid simple spec",
			spec:    testutil.SimpleOpenAPISpec(),
			wantErr: false,
		},
		{
			name:    "valid complex spec",
			spec:    testutil.ComplexOpenAPISpec(),
			wantErr: false,
		},
		{
			name:    "invalid yaml",
			spec:    "invalid: yaml: content: [",
			wantErr: true,
			errMsg:  "failed to load OpenAPI spec",
		},
		{
			name: "missing openapi version",
			spec: `
info:
  title: Test
  version: 1.0.0
`,
			wantErr: true,
			errMsg:  "invalid OpenAPI specification",
		},
		{
			name: "unsupported openapi version",
			spec: `
openapi: 2.0.0
info:
  title: Test
  version: 1.0.0
`,
			wantErr: true,
			errMsg:  "invalid OpenAPI specification",
		},
		{
			name: "missing info section",
			spec: `
openapi: 3.0.0
`,
			wantErr: true,
			errMsg:  "invalid OpenAPI specification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			specFile := testutil.CreateTempFile(t, "test.yaml", tt.spec)

			parser, err := NewOpenAPIParser(specFile)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, parser)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, parser)
			}
		})
	}
}

func TestNewOpenAPIParser_FileNotFound(t *testing.T) {
	_, err := NewOpenAPIParser("nonexistent.yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load OpenAPI spec")
}

func TestOpenAPIParser_GetSchemas(t *testing.T) {
	parser := CreateTestParser(t, testutil.SimpleOpenAPISpec())

	schemas := parser.GetSchemas()

	require.NotEmpty(t, schemas)
	assert.Contains(t, schemas, "User")

	userSchema := schemas["User"]
	require.NotNil(t, userSchema)
	assert.Equal(t, "object", userSchema.Type)
	assert.Contains(t, userSchema.Properties, "id")
	assert.Contains(t, userSchema.Properties, "name")
	assert.Contains(t, userSchema.Properties, "email")
}

func TestOpenAPIParser_GetSchemaByName(t *testing.T) {
	parser := CreateTestParser(t, testutil.ComplexOpenAPISpec())

	tests := []struct {
		name       string
		schemaName string
		wantExists bool
		wantType   string
	}{
		{
			name:       "existing schema",
			schemaName: "User",
			wantExists: true,
			wantType:   "object",
		},
		{
			name:       "another existing schema",
			schemaName: "Product",
			wantExists: true,
			wantType:   "object",
		},
		{
			name:       "nested schema",
			schemaName: "UserProfile",
			wantExists: true,
			wantType:   "object",
		},
		{
			name:       "non-existing schema",
			schemaName: "NonExistent",
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, exists := parser.GetSchemaByName(tt.schemaName)

			assert.Equal(t, tt.wantExists, exists)
			if tt.wantExists {
				require.NotNil(t, schema)
				assert.Equal(t, tt.wantType, schema.Type)
			} else {
				assert.Nil(t, schema)
			}
		})
	}
}

func TestOpenAPIParser_GetPaths(t *testing.T) {
	parser := CreateTestParser(t, testutil.SimpleOpenAPISpec())

	paths := parser.GetPaths()

	require.NotEmpty(t, paths)
	assert.Contains(t, paths, "/users")
	assert.Contains(t, paths, "/users/{id}")

	usersPath := paths["/users"]
	require.NotNil(t, usersPath)
	assert.NotNil(t, usersPath.Get)
	assert.NotNil(t, usersPath.Post)

	userByIdPath := paths["/users/{id}"]
	require.NotNil(t, userByIdPath)
	assert.NotNil(t, userByIdPath.Get)
	assert.NotNil(t, userByIdPath.Put)
	assert.NotNil(t, userByIdPath.Delete)
}

func TestOpenAPIParser_GetOperations(t *testing.T) {
	parser := CreateTestParser(t, testutil.SimpleOpenAPISpec())

	operations := parser.GetOperations()

	require.NotEmpty(t, operations)

	// Check for specific operations
	assert.Contains(t, operations, "listUsers")
	assert.Contains(t, operations, "createUser")
	assert.Contains(t, operations, "getUser")
	assert.Contains(t, operations, "updateUser")
	assert.Contains(t, operations, "deleteUser")

	// Verify operation details
	listUsersOp := operations["listUsers"]
	require.NotNil(t, listUsersOp)
	assert.Equal(t, "listUsers", listUsersOp.OperationID)
}

func TestOpenAPIParser_GetOperationsByTag(t *testing.T) {
	parser := CreateTestParser(t, testutil.SimpleOpenAPISpec())

	tests := []struct {
		name      string
		tag       string
		wantCount int
		wantOps   []string
	}{
		{
			name:      "User tag operations",
			tag:       "User",
			wantCount: 5,
			wantOps:   []string{"listUsers", "createUser", "getUser", "updateUser", "deleteUser"},
		},
		{
			name:      "non-existing tag",
			tag:       "NonExistent",
			wantCount: 0,
			wantOps:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operations := parser.GetOperationsByTag(tt.tag)

			assert.Len(t, operations, tt.wantCount)

			operationIds := make([]string, len(operations))
			for i, op := range operations {
				operationIds[i] = op.OperationID
			}

			for _, expectedOp := range tt.wantOps {
				assert.Contains(t, operationIds, expectedOp)
			}
		})
	}
}

func TestOpenAPIParser_GetOperationByID(t *testing.T) {
	parser := CreateTestParser(t, testutil.SimpleOpenAPISpec())

	tests := []struct {
		name        string
		operationID string
		wantExists  bool
	}{
		{
			name:        "existing operation",
			operationID: "listUsers",
			wantExists:  true,
		},
		{
			name:        "another existing operation",
			operationID: "createUser",
			wantExists:  true,
		},
		{
			name:        "non-existing operation",
			operationID: "nonExistent",
			wantExists:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operation, exists := parser.GetOperationByID(tt.operationID)

			assert.Equal(t, tt.wantExists, exists)
			if tt.wantExists {
				require.NotNil(t, operation)
				assert.Equal(t, tt.operationID, operation.OperationID)
			} else {
				assert.Nil(t, operation)
			}
		})
	}
}

func TestOpenAPIParser_GetInfo(t *testing.T) {
	parser := CreateTestParser(t, testutil.SimpleOpenAPISpec())

	info := parser.GetInfo()

	require.NotNil(t, info)
	assert.Equal(t, "Test API", info.Title)
	assert.Equal(t, "1.0.0", info.Version)
}

func TestOpenAPIParser_GetCrudOperationsForSchema(t *testing.T) {
	parser := CreateTestParser(t, testutil.SimpleOpenAPISpec())

	crudOps := parser.GetCrudOperationsForSchema("User")

	require.NotEmpty(t, crudOps)

	// The exact mapping depends on the implementation,
	// but we should have some CRUD operations
	assert.True(t, len(crudOps) > 0, "Expected some CRUD operations for User schema")
}

// Benchmark tests
func BenchmarkNewOpenAPIParser(b *testing.B) {
	spec := testutil.SimpleOpenAPISpec()
	specFile := testutil.CreateTempFile(&testing.T{}, "bench.yaml", spec)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser, err := NewOpenAPIParser(specFile)
		if err != nil {
			b.Fatal(err)
		}
		_ = parser
	}
}

func BenchmarkGetSchemas(b *testing.B) {
	parser := CreateTestParser(&testing.T{}, testutil.ComplexOpenAPISpec())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		schemas := parser.GetSchemas()
		_ = schemas
	}
}

func BenchmarkGetOperations(b *testing.B) {
	parser := CreateTestParser(&testing.T{}, testutil.ComplexOpenAPISpec())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		operations := parser.GetOperations()
		_ = operations
	}
}
