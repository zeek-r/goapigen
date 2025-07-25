package parser

import (
	"fmt"
	"path/filepath"

	"github.com/getkin/kin-openapi/openapi3"
)

// OpenAPIParser represents a parser for OpenAPI specifications
type OpenAPIParser struct {
	Doc *openapi3.T // Exported for testing
}

// NewOpenAPIParser creates a new OpenAPI parser from the specified file path
func NewOpenAPIParser(filePath string) (*OpenAPIParser, error) {
	ext := filepath.Ext(filePath)
	loader := openapi3.NewLoader()
	
	var doc *openapi3.T
	var err error
	
	// Handle both JSON and YAML formats
	switch ext {
	case ".json", ".yaml", ".yml":
		doc, err = loader.LoadFromFile(filePath)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}
	
	if err := doc.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("invalid OpenAPI specification: %w", err)
	}
	
	return &OpenAPIParser{Doc: doc}, nil
}

// GetSchemas returns all schemas defined in the OpenAPI spec
func (p *OpenAPIParser) GetSchemas() map[string]*openapi3.Schema {
	result := make(map[string]*openapi3.Schema)
	
	if p.Doc.Components == nil || p.Doc.Components.Schemas == nil {
		return result
	}
	
	for name, schemaRef := range p.Doc.Components.Schemas {
		if schemaRef != nil && schemaRef.Value != nil {
			result[name] = schemaRef.Value
		}
	}
	
	return result
}

// GetPaths returns all paths defined in the OpenAPI spec
func (p *OpenAPIParser) GetPaths() map[string]*openapi3.PathItem {
	if p.Doc.Paths == nil {
		return map[string]*openapi3.PathItem{}
	}
	return p.Doc.Paths.Map()
}

// GetOperations returns all operations defined in the OpenAPI spec
// The returned map is keyed by operationId
func (p *OpenAPIParser) GetOperations() map[string]*openapi3.Operation {
	result := make(map[string]*openapi3.Operation)
	
	if p.Doc.Paths == nil {
		return result
	}
	
	for _, pathItem := range p.Doc.Paths.Map() {
		if pathItem.Get != nil && pathItem.Get.OperationID != "" {
			result[pathItem.Get.OperationID] = pathItem.Get
		}
		if pathItem.Post != nil && pathItem.Post.OperationID != "" {
			result[pathItem.Post.OperationID] = pathItem.Post
		}
		if pathItem.Put != nil && pathItem.Put.OperationID != "" {
			result[pathItem.Put.OperationID] = pathItem.Put
		}
		if pathItem.Delete != nil && pathItem.Delete.OperationID != "" {
			result[pathItem.Delete.OperationID] = pathItem.Delete
		}
		if pathItem.Options != nil && pathItem.Options.OperationID != "" {
			result[pathItem.Options.OperationID] = pathItem.Options
		}
		if pathItem.Head != nil && pathItem.Head.OperationID != "" {
			result[pathItem.Head.OperationID] = pathItem.Head
		}
		if pathItem.Patch != nil && pathItem.Patch.OperationID != "" {
			result[pathItem.Patch.OperationID] = pathItem.Patch
		}
	}
	
	return result
}

// GetCrudOperationsForSchema identifies CRUD operations for a given schema
// Returns a map of CRUD type to operation ID
func (p *OpenAPIParser) GetCrudOperationsForSchema(schemaName string) map[string]string {
	result := make(map[string]string)
	operations := p.GetOperations()
	
	for opID, operation := range operations {
		// Check if operation has tags matching the schema name
		for _, tag := range operation.Tags {
			if tag == schemaName {
				// Determine CRUD type based on operation ID or path/method
				if p.Doc.Paths != nil {
					// Infer method from operation by checking which method field it is in the path item
					for _, pathItem := range p.Doc.Paths.Map() {
						if pathItem.Get == operation {
							if p.isListOperation(operation) {
								result["list"] = opID
							} else {
								result["get"] = opID
							}
						} else if pathItem.Post == operation {
							result["create"] = opID
						} else if pathItem.Put == operation || pathItem.Patch == operation {
							result["update"] = opID
						} else if pathItem.Delete == operation {
							result["delete"] = opID
						}
					}
				}
			}
		}
	}
	
	return result
}

// isListOperation determines if an operation returns a list of items
// by checking response schemas
func (p *OpenAPIParser) isListOperation(operation *openapi3.Operation) bool {
	// Check for 200 response with array type
	if operation.Responses != nil {
		response := operation.Responses.Value("200")
		if response != nil && response.Value != nil && response.Value.Content != nil {
			for _, mediaType := range response.Value.Content {
				if mediaType.Schema != nil && mediaType.Schema.Value != nil {
					if mediaType.Schema.Value.Type == "array" {
						return true
					}
				}
			}
		}
	}
	return false
}

// GetInfo returns the API info from the OpenAPI spec
func (p *OpenAPIParser) GetInfo() *openapi3.Info {
	return p.Doc.Info
}

// GetSchemaByName returns a specific schema by name
func (p *OpenAPIParser) GetSchemaByName(name string) (*openapi3.Schema, bool) {
	if p.Doc.Components == nil || p.Doc.Components.Schemas == nil {
		return nil, false
	}

	schemaRef, exists := p.Doc.Components.Schemas[name]
	if !exists || schemaRef == nil || schemaRef.Value == nil {
		return nil, false
	}

	return schemaRef.Value, true
}

// GetOperationByID returns a specific operation by ID
func (p *OpenAPIParser) GetOperationByID(operationID string) (*openapi3.Operation, bool) {
	operations := p.GetOperations()
	operation, exists := operations[operationID]
	return operation, exists
}

// GetOperationsByTag returns all operations with a specific tag
func (p *OpenAPIParser) GetOperationsByTag(tag string) []*openapi3.Operation {
	result := make([]*openapi3.Operation, 0)
	
	if p.Doc.Paths != nil {
		for _, pathItem := range p.Doc.Paths.Map() {
			operations := []*openapi3.Operation{
				pathItem.Get,
				pathItem.Post,
				pathItem.Put,
				pathItem.Delete,
				pathItem.Options,
				pathItem.Head,
				pathItem.Patch,
			}
			
			for _, op := range operations {
				if op == nil {
					continue
				}
				
				for _, opTag := range op.Tags {
					if opTag == tag {
						result = append(result, op)
						break
					}
				}
			}
		}
	}
	
	return result
}