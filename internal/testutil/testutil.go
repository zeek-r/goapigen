package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/require"
)

// CreateTempFile creates a temporary file with the given content for testing
func CreateTempFile(t *testing.T, filename, content string) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "goapigen-test")
	require.NoError(t, err)

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	filePath := filepath.Join(tempDir, filename)
	err = os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	return filePath
}

// SimpleOpenAPISpec returns a basic OpenAPI spec for testing
func SimpleOpenAPISpec() string {
	return `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
components:
  schemas:
    User:
      type: object
      required: [name]
      properties:
        id:
          type: string
          format: uuid
          description: User ID
        name:
          type: string
          minLength: 1
          maxLength: 100
          description: User name
        email:
          type: string
          format: email
          description: User email
        age:
          type: integer
          minimum: 0
          maximum: 150
          description: User age
        created_at:
          type: string
          format: date-time
          description: Creation timestamp
        status:
          type: string
          enum: [active, inactive, pending]
          description: User status
paths:
  /users:
    get:
      operationId: listUsers
      tags: [User]
      summary: List users
      responses:
        '200':
          description: List of users
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
    post:
      operationId: createUser
      tags: [User]
      summary: Create user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      responses:
        '201':
          description: User created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
  /users/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: string
          format: uuid
    get:
      operationId: getUser
      tags: [User]
      summary: Get user by ID
      responses:
        '200':
          description: User found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: User not found
    put:
      operationId: updateUser
      tags: [User]
      summary: Update user
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      responses:
        '200':
          description: User updated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/User'
        '404':
          description: User not found
    delete:
      operationId: deleteUser
      tags: [User]
      summary: Delete user
      responses:
        '204':
          description: User deleted
        '404':
          description: User not found
`
}

// ComplexOpenAPISpec returns a more complex OpenAPI spec for testing
func ComplexOpenAPISpec() string {
	return `
openapi: 3.0.0
info:
  title: Complex Test API
  version: 2.0.0
components:
  schemas:
    User:
      type: object
      required: [name, email]
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
          minLength: 1
          maxLength: 100
        email:
          type: string
          format: email
        profile:
          $ref: '#/components/schemas/UserProfile'
        tags:
          type: array
          items:
            type: string
        metadata:
          type: object
          additionalProperties:
            type: string
        created_at:
          type: string
          format: date-time
    UserProfile:
      type: object
      properties:
        bio:
          type: string
          maxLength: 500
        avatar_url:
          type: string
          format: uri
        age:
          type: integer
          minimum: 13
          maximum: 120
        preferences:
          type: object
          properties:
            theme:
              type: string
              enum: [light, dark, auto]
            notifications:
              type: boolean
    Product:
      type: object
      required: [name, price]
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
          minLength: 1
          maxLength: 200
        description:
          type: string
          maxLength: 1000
        price:
          type: number
          minimum: 0
        category:
          type: string
          enum: [electronics, books, clothing, home]
        in_stock:
          type: boolean
        created_at:
          type: string
          format: date-time
paths:
  /users:
    get:
      operationId: listUsers
      tags: [User]
      responses:
        '200':
          description: Users list
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
    post:
      operationId: createUser
      tags: [User]
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/User'
      responses:
        '201':
          description: User created
  /products:
    get:
      operationId: listProducts
      tags: [Product]
      responses:
        '200':
          description: Products list
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Product'
`
}

// MockSchema creates a mock OpenAPI schema for testing
func MockSchema(schemaType string, props map[string]*openapi3.Schema) *openapi3.Schema {
	schemaRefs := make(map[string]*openapi3.SchemaRef)
	for name, schema := range props {
		schemaRefs[name] = &openapi3.SchemaRef{Value: schema}
	}

	return &openapi3.Schema{
		Type:       schemaType,
		Properties: schemaRefs,
	}
}

// AssertContainsAll checks that the haystack contains all needles
func AssertContainsAll(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		require.Contains(t, haystack, needle, "Expected %q to contain %q", haystack, needle)
	}
}

// AssertNotContainsAny checks that the haystack doesn't contain any needles
func AssertNotContainsAny(t *testing.T, haystack string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		require.NotContains(t, haystack, needle, "Expected %q to not contain %q", haystack, needle)
	}
}
