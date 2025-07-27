package generator

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/zeek-r/goapigen/internal/parser"
)

// EnumField represents a field with enum constraints
type EnumField struct {
	Name    string
	Type    string
	JsonTag string
	Values  []string
}

// MinMaxField represents a field with min/max constraints
type MinMaxField struct {
	Name    string
	Type    string
	JsonTag string
	Min     float64
	HasMax  bool
	Max     float64
}

// ServiceTemplateData contains data for the service template
type ServiceTemplateData struct {
	SchemaName     string
	VarName        string
	PackageName    string
	ImportPath     string
	HasCreateOp    bool
	HasGetOp       bool
	HasListOp      bool
	HasUpdateOp    bool
	HasDeleteOp    bool
	HasCreatedAt   bool
	HasUpdatedAt   bool
	CreateFields   []RequestField
	UpdateFields   []RequestField
	RequiredFields []RequestField
	EnumFields     []EnumField
	MinMaxFields   []MinMaxField
	ImportTime     bool
}

// ServiceGenerator generates service implementations for API schemas
type ServiceGenerator struct {
	parser      *parser.OpenAPIParser
	packageName string
	importPath  string
	typeGen     *TypeGenerator
	templates   *template.Template
}

// NewServiceGenerator creates a new service generator
func NewServiceGenerator(parser *parser.OpenAPIParser, packageName string, importPath string, templateFS embed.FS) (*ServiceGenerator, error) {
	// Create templates with function map
	tmpl := template.New("")
	tmpl.Funcs(template.FuncMap{
		"contains": func(s, substr string) bool { return strings.Contains(s, substr) },
		"lower":    func(s string) string { return strings.ToLower(s) },
	})

	// Parse templates
	tmpl, err := tmpl.ParseFS(templateFS,
		"templates/service/service.go.tmpl",
		"templates/service/service_test.go.tmpl",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &ServiceGenerator{
		parser:      parser,
		packageName: packageName,
		importPath:  importPath,
		typeGen:     NewTypeGenerator(parser, packageName, templateFS),
		templates:   tmpl,
	}, nil
}

// GenerateService generates a service for a schema
func (g *ServiceGenerator) GenerateService(schemaName string) (string, error) {
	// Generate the template data
	data, err := g.prepareTemplateData(schemaName)
	if err != nil {
		return "", err
	}

	// Render the template
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "service.go.tmpl", data); err != nil {
		return "", fmt.Errorf("failed to render service template: %w", err)
	}

	return buf.String(), nil
}

// GenerateServiceTests generates test code for a service
func (g *ServiceGenerator) GenerateServiceTests(schemaName string) (string, error) {
	// Generate the template data
	data, err := g.prepareTemplateData(schemaName)
	if err != nil {
		return "", err
	}

	// Render the template
	var buf bytes.Buffer
	if err := g.templates.ExecuteTemplate(&buf, "service_test.go.tmpl", data); err != nil {
		return "", fmt.Errorf("failed to render service test template: %w", err)
	}

	return buf.String(), nil
}

// prepareTemplateData prepares data for the service templates
func (g *ServiceGenerator) prepareTemplateData(schemaName string) (ServiceTemplateData, error) {
	// Check if schema exists
	schema, exists := g.parser.GetSchemaByName(schemaName)
	if !exists {
		return ServiceTemplateData{}, fmt.Errorf("schema %s not found", schemaName)
	}

	// Get CRUD operations for this schema
	crudOps := g.parser.GetCrudOperationsForSchema(schemaName)

	// Prepare field data
	var createFields, updateFields, requiredFields []RequestField
	var enumFields []EnumField
	var minMaxFields []MinMaxField
	importTime := false

	for propName, propRef := range schema.Properties {
		if propRef == nil || propRef.Value == nil {
			continue
		}

		fieldName := ToGoFieldName(propName)
		fieldType, err := MapSchemaToGoType(propRef.Value)
		if err != nil {
			return ServiceTemplateData{}, fmt.Errorf("failed to map field %s: %w", propName, err)
		}

		field := RequestField{
			Name:    fieldName,
			Type:    fieldType,
			JsonTag: propName,
		}

		// Check if required
		for _, required := range schema.Required {
			if required == propName {
				requiredFields = append(requiredFields, field)
				break
			}
		}

		// Check for enum values
		if len(propRef.Value.Enum) > 0 {
			values := make([]string, 0, len(propRef.Value.Enum))
			for _, enumVal := range propRef.Value.Enum {
				if strVal, ok := enumVal.(string); ok {
					values = append(values, strVal)
				}
			}

			if len(values) > 0 {
				enumFields = append(enumFields, EnumField{
					Name:    fieldName,
					Type:    fieldType,
					JsonTag: propName,
					Values:  values,
				})
			}
		}

		// Check for min/max constraints
		hasMin := propRef.Value.Min != nil
		hasMax := propRef.Value.Max != nil

		if hasMin || hasMax {
			mmField := MinMaxField{
				Name:    fieldName,
				Type:    fieldType,
				JsonTag: propName,
			}

			if hasMin {
				mmField.Min = *propRef.Value.Min
			}

			if hasMax {
				mmField.HasMax = true
				mmField.Max = *propRef.Value.Max
			}

			minMaxFields = append(minMaxFields, mmField)
		}

		// Skip ID field for create operation
		if propName == "id" || propName == "ID" {
			updateFields = append(updateFields, field)
			continue
		}

		// Skip timestamp fields that are autogenerated
		if propName == "created_at" || propName == "createdAt" ||
			propName == "updated_at" || propName == "updatedAt" {
			continue
		}

		// Add to appropriate operation fields
		createFields = append(createFields, field)
		updateFields = append(updateFields, field)

		// Check if we need to import time package
		if fieldType == "time.Time" || fieldType == "*time.Time" {
			importTime = true
		}
	}

	// Prepare template data
	data := ServiceTemplateData{
		SchemaName:     schemaName,
		VarName:        ToCamelCase(schemaName),
		PackageName:    g.packageName,
		ImportPath:     g.importPath,
		HasCreateOp:    false,
		HasGetOp:       false,
		HasListOp:      false,
		HasUpdateOp:    false,
		HasDeleteOp:    false,
		HasCreatedAt:   false,
		HasUpdatedAt:   false,
		CreateFields:   createFields,
		UpdateFields:   updateFields,
		RequiredFields: requiredFields,
		EnumFields:     enumFields,
		MinMaxFields:   minMaxFields,
		ImportTime:     importTime,
	}

	// Set operation flags based on OpenAPI spec
	if _, ok := crudOps["create"]; ok {
		data.HasCreateOp = true
	}
	if _, ok := crudOps["get"]; ok {
		data.HasGetOp = true
	}
	if _, ok := crudOps["list"]; ok {
		data.HasListOp = true
	}
	if _, ok := crudOps["update"]; ok {
		data.HasUpdateOp = true
	}
	if _, ok := crudOps["delete"]; ok {
		data.HasDeleteOp = true
	}

	// Check for timestamp fields
	for propName, propRef := range schema.Properties {
		if propRef != nil && propRef.Value != nil {
			if propName == "created_at" || propName == "createdAt" {
				data.HasCreatedAt = true
				data.ImportTime = true
			}
			if propName == "updated_at" || propName == "updatedAt" {
				data.HasUpdatedAt = true
				data.ImportTime = true
			}
		}
	}

	return data, nil
}
