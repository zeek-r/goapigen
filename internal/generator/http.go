package generator

import (
	"bytes"
	"embed"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/zeek-r/goapigen/internal/parser"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// PathParam represents a path parameter for a handler
type PathParam struct {
	ParamName string
	VarName   string
	Type      string
}

// QueryParam represents a query parameter for a handler
type QueryParam struct {
	ParamName string
	VarName   string
	Type      string
}

// RequestField represents a field in a request body
type RequestField struct {
	Name    string
	Type    string
	JsonTag string
}

// OperationData contains data for an operation handler
type OperationData struct {
	SchemaName       string
	OperationID      string
	ServiceInterface string
	Method           string
	Path             string
	HasPathParams    bool
	PathParams       []PathParam
	HasQueryParams   bool
	QueryParams      []QueryParam
	HasRequestBody   bool
	RequestType      string
	RequestTypeName  string
	RequestFields    []RequestField
	HasResponseBody  bool
	ResponseType     string
	SuccessStatus    int
	HandlerPackage   string
	PackageName      string
	ModelImportPath  string
	ImportPath       string
	VarName          string
	ImportTime       bool
	Domain           string // Domain/resource this operation belongs to
}

// ResourceData represents a resource group in the API
type ResourceData struct {
	Name           string
	BasePath       string
	Operations     []OperationData
	Domain         string
	ImportPath     string
	HandlerPackage string
}

// RouterData contains data for router generation
type RouterData struct {
	HandlerPackage string
	Resources      []ResourceData
	Operations     []OperationData
}

// HTTPGenerator generates API handlers for HTTP endpoints
type HTTPGenerator struct {
	parser          *parser.OpenAPIParser
	packageName     string
	handlerPackage  string
	importPath      string
	modelImportPath string
	templates       *template.Template
}

// NewHTTPGenerator creates a new generator for HTTP handlers
func NewHTTPGenerator(
	parser *parser.OpenAPIParser,
	packageName string,
	handlerPackage string,
	importPath string,
	modelImportPath string,
	templateFS embed.FS,
) (*HTTPGenerator, error) {
	// Create templates with function map
	tmpl := template.New("")
	tmpl.Funcs(template.FuncMap{
		"contains": func(s, substr string) bool { return strings.Contains(s, substr) },
		"title":    func(s string) string { return cases.Title(language.English).String(s) },
		"lower":    func(s string) string { return strings.ToLower(s) },
	})

	// Parse templates
	tmpl, err := tmpl.ParseFS(templateFS,
		"templates/http/operation_handler.go.tmpl",
		"templates/http/operation_handler_test.go.tmpl",
		"templates/http/handler_wrapper.go.tmpl",
		"templates/http/http_utils.go.tmpl",
		"templates/http/router.go.tmpl",
		"templates/http/mocks.go.tmpl",
		"templates/http/schema_handler.go.tmpl",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse handler templates: %w", err)
	}

	return &HTTPGenerator{
		parser:          parser,
		packageName:     packageName,
		handlerPackage:  handlerPackage,
		importPath:      importPath,
		modelImportPath: modelImportPath,
		templates:       tmpl,
	}, nil
}

// GenerateHandlers generates all HTTP handlers for the API
func (g *HTTPGenerator) GenerateHandlers() (map[string]string, error) {
	operations := g.parser.GetOperations()
	result := make(map[string]string)

	// Group operations by tag (resource)
	resourceMap := make(map[string][]OperationData)
	var allOperations []OperationData

	// Create directory for httputil package
	// Generate common HTTP utilities in internal/pkg/httputil package
	httpUtils, err := g.generateHTTPUtils()
	if err != nil {
		return nil, fmt.Errorf("failed to generate HTTP utilities: %w", err)
	}
	result["httputil/http_utils.go"] = httpUtils

	// Generate handler wrapper in internal/pkg/httputil package
	handlerWrapper, err := g.generateHandlerWrapper()
	if err != nil {
		return nil, fmt.Errorf("failed to generate handler wrapper: %w", err)
	}
	result["httputil/handler_wrapper.go"] = handlerWrapper

	// Track domains we've seen to generate mocks only once per domain
	generatedMocks := make(map[string]bool)

	// Generate handler file for each operation
	for opID, operation := range operations {
		if opID == "" {
			continue // Skip operations without ID
		}

		data, err := g.prepareOperationData(opID, operation)
		if err != nil {
			return nil, fmt.Errorf("failed to prepare data for operation %s: %w", opID, err)
		}

		// Group by first tag (primary resource)
		if len(operation.Tags) > 0 {
			tag := operation.Tags[0]
			if _, exists := resourceMap[tag]; !exists {
				resourceMap[tag] = make([]OperationData, 0)
			}
			resourceMap[tag] = append(resourceMap[tag], data)

			// Store domain with data for later use
			data.Domain = strings.ToLower(tag)

			// Generate mocks for each domain if not already generated
			if !generatedMocks[data.Domain] {
				// Generate mocks
				mockCode, err := g.generateMocks(data)
				if err != nil {
					return nil, fmt.Errorf("failed to generate mocks for domain %s: %w", data.Domain, err)
				}

				// Create mocks directory path
				mockFilename := "domain/" + data.Domain + "/mocks/mock_service.go"
				result[mockFilename] = mockCode

				// Mark domain as having mocks generated
				generatedMocks[data.Domain] = true
			}
		}
		allOperations = append(allOperations, data)

		// Generate handler file
		code, err := g.generateOperationHandler(data)
		if err != nil {
			return nil, fmt.Errorf("failed to generate handler for operation %s: %w", opID, err)
		}

		// Generate handler tests
		testCode, err := g.generateOperationHandlerTests(data)
		if err != nil {
			return nil, fmt.Errorf("failed to generate handler tests for operation %s: %w", opID, err)
		}

		// Put handlers in their domain subdirectory
		filename := strings.ToLower(opID) + "_handler.go"
		testFilename := strings.ToLower(opID) + "_handler_test.go"
		if data.Domain != "" {
			filename = "domain/" + data.Domain + "/" + filename
			testFilename = "domain/" + data.Domain + "/" + testFilename
		}
		result[filename] = code
		result[testFilename] = testCode
	}

	// Generate router
	resources := make([]ResourceData, 0, len(resourceMap))
	for name, ops := range resourceMap {
		// Determine base path for the resource
		basePath := "/" + strings.ToLower(name)

		resourceData := ResourceData{
			Name:           ToPascalCase(name),
			BasePath:       basePath,
			Operations:     ops,
			Domain:         strings.ToLower(name),
			ImportPath:     g.importPath,
			HandlerPackage: g.handlerPackage,
		}
		resources = append(resources, resourceData)

		// Generate schema handler
		schemaHandlerCode, err := g.generateSchemaHandler(resourceData)
		if err != nil {
			return nil, fmt.Errorf("failed to generate schema handler for resource %s: %w", name, err)
		}

		// Normalize domain name to lowercase
		domain := strings.ToLower(name)

		// Set schema handler file path
		schemaHandlerFilename := "domain/" + domain + "/handler.go"
		result[schemaHandlerFilename] = schemaHandlerCode
	}

	// Sort resources for consistent output
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})

	return result, nil
}

// prepareOperationData prepares template data for an operation
func (g *HTTPGenerator) prepareOperationData(opID string, operation *openapi3.Operation) (OperationData, error) {
	// Find the path and method for this operation
	var httpMethod, path string
	for p, pathItem := range g.parser.GetPaths() {
		for method, op := range map[string]*openapi3.Operation{
			"GET":     pathItem.Get,
			"POST":    pathItem.Post,
			"PUT":     pathItem.Put,
			"DELETE":  pathItem.Delete,
			"PATCH":   pathItem.Patch,
			"OPTIONS": pathItem.Options,
			"HEAD":    pathItem.Head,
		} {
			if op != nil && op.OperationID == opID {
				httpMethod = method
				path = p
				break
			}
		}
		if httpMethod != "" {
			break
		}
	}

	if httpMethod == "" || path == "" {
		return OperationData{}, fmt.Errorf("could not find method and path for operation %s", opID)
	}

	// Parse path parameters
	pathParams := make([]PathParam, 0)
	for _, param := range operation.Parameters {
		if param.Value == nil || param.Value.In != "path" {
			continue
		}

		pathParams = append(pathParams, PathParam{
			ParamName: param.Value.Name,
			VarName:   ToCamelCase(param.Value.Name),
			Type:      MapParameterTypeToGo(param.Value),
		})
	}

	// Parse query parameters
	queryParams := make([]QueryParam, 0)
	for _, param := range operation.Parameters {
		if param.Value == nil || param.Value.In != "query" {
			continue
		}

		queryParams = append(queryParams, QueryParam{
			ParamName: param.Value.Name,
			VarName:   ToCamelCase(param.Value.Name),
			Type:      MapParameterTypeToGo(param.Value),
		})
	}

	// Determine if there's a request body
	var hasRequestBody bool
	var requestType string
	var requestTypeName string
	var requestFields []RequestField

	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		hasRequestBody = true

		// Find the content type (prefer application/json)
		var contentType string
		var schema *openapi3.Schema
		for ct, mediaType := range operation.RequestBody.Value.Content {
			if mediaType.Schema != nil && mediaType.Schema.Value != nil {
				if ct == "application/json" || contentType == "" {
					contentType = ct
					schema = mediaType.Schema.Value
				}
			}
		}

		if schema != nil {
			// For nested request types, generate an operation-specific type name
			requestTypeName = ToPascalCase(opID) + "Request"
			requestType = requestTypeName

			// Extract fields from schema
			for propName, propRef := range schema.Properties {
				if propRef != nil && propRef.Value != nil {
					// Skip system fields for create operations
					if httpMethod == "POST" && (propName == "id" || propName == "created_at" || propName == "updated_at") {
						continue
					}

					goType, err := MapSchemaToGoType(propRef.Value)
					if err != nil {
						return OperationData{}, fmt.Errorf("failed to map request field %s: %w", propName, err)
					}

					requestFields = append(requestFields, RequestField{
						Name:    ToGoFieldName(propName),
						Type:    goType,
						JsonTag: propName,
					})
				}
			}

			// Sort fields for consistent output
			sort.Slice(requestFields, func(i, j int) bool {
				return requestFields[i].Name < requestFields[j].Name
			})
		}
	}

	// Determine if there's a response body
	var hasResponseBody bool
	var responseType string
	var successStatus int

	if operation.Responses != nil {
		// Find success response (2xx)
		for statusCode, response := range operation.Responses.Map() {
			if statusCode[0] == '2' {
				successStatusInt := 0
				fmt.Sscanf(statusCode, "%d", &successStatusInt)
				successStatus = successStatusInt

				// Default if not specified
				if successStatus == 0 {
					successStatus = 200
				}

				if response.Value != nil && len(response.Value.Content) > 0 {
					hasResponseBody = true

					// Find the content type (prefer application/json)
					var contentType string
					var schema *openapi3.Schema
					for ct, mediaType := range response.Value.Content {
						if mediaType.Schema != nil && mediaType.Schema.Value != nil {
							if ct == "application/json" || contentType == "" {
								contentType = ct
								schema = mediaType.Schema.Value
							}
						}
					}

					if schema != nil {
						var err error
						responseType, err = MapSchemaToGoType(schema)
						if err != nil {
							return OperationData{}, fmt.Errorf("failed to map response schema: %w", err)
						}

						// If schema is a reference to a model, extract the type name
						if mediaType := response.Value.Content["application/json"]; mediaType != nil &&
							mediaType.Schema != nil && mediaType.Schema.Ref != "" {
							parts := strings.Split(mediaType.Schema.Ref, "/")
							if len(parts) > 0 {
								typeName := parts[len(parts)-1]
								responseType = fmt.Sprintf("%s.%s", g.packageName, typeName)
							}
						} else if schema.Items != nil && schema.Items.Ref != "" {
							// Handle array of references
							parts := strings.Split(schema.Items.Ref, "/")
							if len(parts) > 0 {
								typeName := parts[len(parts)-1]
								responseType = fmt.Sprintf("[]%s.%s", g.packageName, typeName)
							}
						} else if !strings.Contains(responseType, ".") &&
							responseType != "string" && responseType != "int" && responseType != "bool" &&
							responseType != "float64" && responseType != "interface{}" &&
							responseType != "map[string]interface{}" && !strings.HasPrefix(responseType, "[]") {
							// It's a complex type that's not a primitive or built-in
							responseType = fmt.Sprintf("%s.%s", g.packageName, responseType)
						}
					}
				}
				break
			}
		}
	}

	// Default success status if not found
	if successStatus == 0 {
		if httpMethod == "POST" {
			successStatus = 201
		} else if httpMethod == "DELETE" {
			successStatus = 204
		} else {
			successStatus = 200
		}
	}

	// Determine schema name from tags or operation ID
	var schemaName string
	if len(operation.Tags) > 0 {
		// Use the first tag as the schema name
		schemaName = ToPascalCase(operation.Tags[0])
	} else {
		// Extract from operation ID if possible
		// This is a simple heuristic and might need improvement
		parts := strings.Split(opID, "_")
		if len(parts) > 0 {
			schemaName = ToPascalCase(parts[0])
		} else {
			// Fallback
			schemaName = ToPascalCase(opID)
		}
	}

	// Determine service interface name
	serviceInterface := schemaName + "Service"
	varName := opID // Use original opID instead of ToCamelCase to match handler names

	// Check if we need to import the time package
	importTime := false

	// Check request fields for time.Time usage
	for _, field := range requestFields {
		if field.Type == "time.Time" || strings.HasPrefix(field.Type, "[]time.Time") {
			importTime = true
			break
		}
	}

	// Check response type for time.Time usage
	if !importTime && hasResponseBody {
		if responseType == "time.Time" || strings.HasPrefix(responseType, "[]time.Time") ||
			strings.Contains(responseType, "time.Time") {
			importTime = true
		}
	}

	return OperationData{
		SchemaName:       schemaName,
		OperationID:      opID,
		ServiceInterface: serviceInterface,
		Method:           httpMethod,
		Path:             path,
		HasPathParams:    len(pathParams) > 0,
		PathParams:       pathParams,
		HasQueryParams:   len(queryParams) > 0,
		QueryParams:      queryParams,
		HasRequestBody:   hasRequestBody,
		RequestType:      requestType,
		RequestTypeName:  requestTypeName,
		RequestFields:    requestFields,
		HasResponseBody:  hasResponseBody,
		ResponseType:     responseType,
		SuccessStatus:    successStatus,
		HandlerPackage:   g.handlerPackage,
		PackageName:      g.packageName,
		ModelImportPath:  g.modelImportPath,
		ImportPath:       g.importPath,
		VarName:          varName,
		ImportTime:       importTime,
	}, nil
}

// generateOperationHandler generates code for a single operation handler
func (g *HTTPGenerator) generateOperationHandler(data OperationData) (string, error) {
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "operation_handler.go.tmpl", data); err != nil {
		return "", fmt.Errorf("failed to render operation handler template: %w", err)
	}
	return buf.String(), nil
}

// generateOperationHandlerTests generates test code for a single operation handler
func (g *HTTPGenerator) generateOperationHandlerTests(data OperationData) (string, error) {
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "operation_handler_test.go.tmpl", data); err != nil {
		return "", fmt.Errorf("failed to render operation handler test template: %w", err)
	}
	return buf.String(), nil
}

// generateHTTPUtils generates the HTTP utilities file
func (g *HTTPGenerator) generateHTTPUtils() (string, error) {
	var buf bytes.Buffer
	data := struct {
		ImportPath string
	}{
		ImportPath: g.importPath,
	}
	if err := g.templates.ExecuteTemplate(&buf, "http_utils.go.tmpl", data); err != nil {
		return "", fmt.Errorf("failed to render HTTP utilities template: %w", err)
	}

	return buf.String(), nil
}

// generateHandlerWrapper generates the generic handler wrapper file
func (g *HTTPGenerator) generateHandlerWrapper() (string, error) {
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "handler_wrapper.go.tmpl", nil); err != nil {
		return "", fmt.Errorf("failed to render handler wrapper template: %w", err)
	}

	return buf.String(), nil
}

// generateMocks generates mock implementations for the service interfaces
func (g *HTTPGenerator) generateMocks(data OperationData) (string, error) {
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "mocks.go.tmpl", data); err != nil {
		return "", fmt.Errorf("failed to render mocks template: %w", err)
	}
	return buf.String(), nil
}

// generateSchemaHandler creates a handler file that provides a function to register all operation handlers for a schema
func (g *HTTPGenerator) generateSchemaHandler(resource ResourceData) (string, error) {
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "schema_handler.go.tmpl", resource); err != nil {
		return "", fmt.Errorf("failed to render schema handler template: %w", err)
	}

	return buf.String(), nil
}
