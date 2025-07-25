package generator

import (
	"bytes"
	"embed"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/zeek-r/goapigen/internal/config"
	"github.com/zeek-r/goapigen/internal/parser"
)

// TypeGenerator generates Go type definitions from OpenAPI schemas
type TypeGenerator struct {
	parser      *parser.OpenAPIParser
	packageName string
	templateFS  embed.FS
}

// TypeField represents a field in a struct type
type TypeField struct {
	Name    string
	Type    string
	Tags    string
	Comment string
}

// TypeDefinition represents a Go type definition
type TypeDefinition struct {
	Name   string
	Fields []TypeField
}

// TypeTemplateData represents the data needed for the type template
type TypeTemplateData struct {
	HasTimeType       bool
	AdditionalImports []string
	Types             []TypeDefinition
}

// NewTypeGenerator creates a new type generator with the given parser
func NewTypeGenerator(parser *parser.OpenAPIParser, packageName string, templateFS embed.FS) *TypeGenerator {
	return &TypeGenerator{
		parser:      parser,
		packageName: packageName,
		templateFS:  templateFS,
	}
}

// GenerateTypes generates Go type definitions for all schemas in the OpenAPI spec
func (g *TypeGenerator) GenerateTypes() (string, error) {
	schemas := g.parser.GetSchemas()

	// Sort schema names for consistent output
	schemaNames := make([]string, 0, len(schemas))
	for name := range schemas {
		schemaNames = append(schemaNames, name)
	}
	sort.Strings(schemaNames)

	// Collect imports
	imports := g.collectImports(schemas)
	hasTimeType := false
	additionalImports := []string{}

	for _, imp := range imports {
		if imp == "time" {
			hasTimeType = true
		} else {
			additionalImports = append(additionalImports, fmt.Sprintf("%q", imp))
		}
	}

	// Build type definitions
	typeDefinitions := make([]TypeDefinition, 0, len(schemaNames))
	for _, name := range schemaNames {
		schema := schemas[name]
		typeDef, err := g.buildTypeDefinition(name, schema)
		if err != nil {
			return "", fmt.Errorf("failed to generate type for %s: %w", name, err)
		}
		typeDefinitions = append(typeDefinitions, typeDef)
	}

	// Template data
	data := TypeTemplateData{
		HasTimeType:       hasTimeType,
		AdditionalImports: additionalImports,
		Types:             typeDefinitions,
	}

	// Load and execute template
	tmpl, err := template.ParseFS(g.templateFS, config.DomainTypesTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse domain types template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute domain types template: %w", err)
	}

	return buf.String(), nil
}

// collectImports determines the necessary import statements
func (g *TypeGenerator) collectImports(schemas map[string]*openapi3.Schema) []string {
	importMap := make(map[string]bool)

	for _, schema := range schemas {
		for _, prop := range schema.Properties {
			if prop.Value == nil {
				continue
			}

			// Check for time.Time
			if prop.Value.Type == "string" && prop.Value.Format == "date-time" {
				importMap["time"] = true
			}

			// Add other imports as needed
		}
	}

	imports := make([]string, 0, len(importMap))
	for imp := range importMap {
		imports = append(imports, imp)
	}
	sort.Strings(imports)
	return imports
}

// buildTypeDefinition builds a TypeDefinition from an OpenAPI schema
func (g *TypeGenerator) buildTypeDefinition(name string, schema *openapi3.Schema) (TypeDefinition, error) {
	var typeDef TypeDefinition
	typeDef.Name = name

	// Build struct fields
	if schema.Type == "object" && schema.Properties != nil {
		// Get required fields
		requiredFields := make(map[string]bool)
		for _, req := range schema.Required {
			requiredFields[req] = true
		}

		// Sort property names for consistent output
		propNames := make([]string, 0, len(schema.Properties))
		for propName := range schema.Properties {
			propNames = append(propNames, propName)
		}
		sort.Strings(propNames)

		// Process each property
		for _, propName := range propNames {
			prop := schema.Properties[propName]
			if prop.Value == nil {
				continue
			}

			// Get Go type for property
			goType, err := g.mapToGoType(prop.Value, true)
			if err != nil {
				return typeDef, fmt.Errorf("failed to map property %s to Go type: %w", propName, err)
			}

			// Format field name properly
			fieldName := formatFieldName(propName)

			// Build tags
			tags := g.generateFieldTags(propName, prop.Value, schema.Required)

			// Add comment if available
			comment := "//"
			if prop.Value.Description != "" {
				comment = "// " + prop.Value.Description
			}

			// Add field to type definition
			typeDef.Fields = append(typeDef.Fields, TypeField{
				Name:    fieldName,
				Type:    goType,
				Tags:    tags,
				Comment: comment,
			})
		}
	}

	return typeDef, nil
}

// generateTypeDefinition generates a Go type definition for a single schema
func (g *TypeGenerator) generateTypeDefinition(name string, schema *openapi3.Schema, isNested bool) (string, error) {
	switch schema.Type {
	case "object":
		return g.generateStructDefinition(name, schema, isNested)
	case "array":
		if schema.Items != nil && schema.Items.Value != nil {
			itemSchema := schema.Items.Value
			itemType, err := g.mapToGoType(itemSchema, false)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("[]%s", itemType), nil
		}
		return "[]interface{}", nil
	default:
		goType, err := g.mapToGoType(schema, isNested)
		if err != nil {
			return "", err
		}
		return goType, nil
	}
}

// generateStructDefinition generates a Go struct definition
func (g *TypeGenerator) generateStructDefinition(name string, schema *openapi3.Schema, isNested bool) (string, error) {
	var buf bytes.Buffer

	if !isNested {
		// Comment with description if available
		if schema.Description != "" {
			buf.WriteString(fmt.Sprintf("// %s %s\n", name, schema.Description))
		} else {
			buf.WriteString(fmt.Sprintf("// %s represents a %s object\n", name, name))
		}

		// Struct declaration
		buf.WriteString(fmt.Sprintf("type %s struct {\n", name))
	} else {
		// Nested struct doesn't need a name
		buf.WriteString("struct {\n")
	}

	// Sort property names for consistent output
	propNames := make([]string, 0, len(schema.Properties))
	for propName := range schema.Properties {
		propNames = append(propNames, propName)
	}
	sort.Strings(propNames)

	for _, propName := range propNames {
		prop := schema.Properties[propName]
		if prop == nil || prop.Value == nil {
			continue
		}

		fieldName := ToGoFieldName(propName)
		fieldType, err := g.mapToGoType(prop.Value, true)
		if err != nil {
			return "", err
		}

		// Field comment if available
		if prop.Value.Description != "" {
			buf.WriteString(fmt.Sprintf("\t// %s\n", prop.Value.Description))
		}

		// Field declaration with tags
		tags := g.generateFieldTags(propName, prop.Value, schema.Required)
		buf.WriteString(fmt.Sprintf("\t%s %s %s\n", fieldName, fieldType, tags))
	}

	buf.WriteString("}")
	return buf.String(), nil
}

// mapToGoType maps an OpenAPI type to a Go type
// This is a legacy function that uses the shared MapSchemaToGoType
// for backward compatibility
func (g *TypeGenerator) mapToGoType(schema *openapi3.Schema, allowNested bool) (string, error) {
	if schema.Type == "object" && allowNested {
		nestedStruct, err := g.generateStructDefinition("", schema, true)
		if err != nil {
			return "", err
		}
		return nestedStruct, nil
	}

	return MapSchemaToGoType(schema)
}

// generateFieldTags generates struct field tags
func (g *TypeGenerator) generateFieldTags(name string, schema *openapi3.Schema, requiredFields []string) string {
	tags := make(map[string]string)

	// JSON tag
	tags["json"] = name

	// BSON tag for MongoDB
	tags["bson"] = name

	// Validation tags
	valTags := []string{}

	// Required validation
	if Contains(requiredFields, name) {
		valTags = append(valTags, "required")
	}

	// String validations
	if schema.Type == "string" {
		if schema.MinLength > 0 {
			valTags = append(valTags, fmt.Sprintf("min=%d", schema.MinLength))
		}
		if schema.MaxLength != nil {
			valTags = append(valTags, fmt.Sprintf("max=%d", *schema.MaxLength))
		}
		if schema.Pattern != "" {
			valTags = append(valTags, fmt.Sprintf("regexp=%s", schema.Pattern))
		}
		if schema.Format != "" {
			valTags = append(valTags, fmt.Sprintf("format=%s", schema.Format))
		}
		if len(schema.Enum) > 0 {
			enumVals := make([]string, len(schema.Enum))
			for i, e := range schema.Enum {
				enumVals[i] = fmt.Sprintf("%v", e)
			}
			valTags = append(valTags, fmt.Sprintf("enum=%s", strings.Join(enumVals, " ")))
		}
	}

	// Number validations
	if schema.Type == "number" || schema.Type == "integer" {
		if schema.Min != nil {
			valTags = append(valTags, fmt.Sprintf("min=%v", *schema.Min))
		}
		if schema.Max != nil {
			valTags = append(valTags, fmt.Sprintf("max=%v", *schema.Max))
		}
	}

	if len(valTags) > 0 {
		tags["validate"] = strings.Join(valTags, ",")
	}

	// Build the tag string
	var tagStrs []string
	for k, v := range tags {
		tagStrs = append(tagStrs, fmt.Sprintf("%s:\"%s\"", k, v))
	}
	sort.Strings(tagStrs)

	if len(tagStrs) > 0 {
		return strings.Join(tagStrs, " ")
	}
	return ""
}

// formatFieldName formats a property name to Go style (PascalCase)
func formatFieldName(name string) string {
	// Special case for ID fields
	if strings.ToLower(name) == "id" {
		return "ID"
	}

	parts := strings.Split(name, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}
