package generator

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/getkin/kin-openapi/openapi3"
)

type BaseStruct struct{}

func (b *BaseStruct) Lower(s string) string {
	return strings.ToLower(s)
}

// Common case conversion functions
// ToSnakeCase converts a string to snake_case
func ToSnakeCase(s string) string {
	var result strings.Builder

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	return strings.ToLower(result.String())
}

// ToCamelCase converts a string to camelCase
func ToCamelCase(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	// Split by non-alphanumeric characters
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})

	// Convert to camelCase
	for i, part := range parts {
		if i == 0 {
			parts[i] = strings.ToLower(part)
		} else {
			parts[i] = strings.Title(strings.ToLower(part))
		}
	}

	return strings.Join(parts, "")
}

// ToPascalCase converts a string to PascalCase
func ToPascalCase(s string) string {
	s = ToCamelCase(s)
	if s == "" {
		return s
	}
	return strings.Title(s)
}

// ToGoFieldName converts a JSON property name to a Go field name
func ToGoFieldName(name string) string {
	// Handle special cases
	if name == "id" {
		return "ID"
	}

	words := SplitWords(name)
	for i, word := range words {
		if i == 0 {
			words[i] = strings.Title(word)
		} else if len(word) > 0 {
			words[i] = strings.Title(word)
		}
	}

	return strings.Join(words, "")
}

// SplitWords splits a string into words based on common delimiters and casing
func SplitWords(s string) []string {
	var words []string
	var lastPos int

	// Split by delimiters like underscore, dash, etc.
	if strings.ContainsAny(s, "_-") {
		words = append(words, strings.FieldsFunc(s, func(r rune) bool {
			return r == '_' || r == '-'
		})...)
		return words
	}

	// Split by camelCase or snake_case
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			words = append(words, s[lastPos:i])
			lastPos = i
		}
	}
	if lastPos < len(s) {
		words = append(words, s[lastPos:])
	}

	return words
}

// Contains checks if a string slice contains a string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Type mapping functions

// MapSchemaToGoType maps an OpenAPI schema to a Go type
func MapSchemaToGoType(schema *openapi3.Schema) (string, error) {
	switch schema.Type {
	case "string":
		switch schema.Format {
		case "date-time":
			return "time.Time", nil
		case "binary":
			return "[]byte", nil
		case "uuid":
			return "string", nil
		default:
			return "string", nil
		}
	case "number":
		switch schema.Format {
		case "float":
			return "float32", nil
		default:
			return "float64", nil
		}
	case "integer":
		switch schema.Format {
		case "int32":
			return "int32", nil
		case "int64":
			return "int64", nil
		default:
			return "int", nil
		}
	case "boolean":
		return "bool", nil
	case "array":
		if schema.Items != nil && schema.Items.Value != nil {
			itemType, err := MapSchemaToGoType(schema.Items.Value)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("[]%s", itemType), nil
		}
		return "[]interface{}", nil
	case "object":
		// Check for additional properties (maps)
		if schema.AdditionalProperties.Schema != nil && schema.AdditionalProperties.Schema.Value != nil {
			valueType, err := MapSchemaToGoType(schema.AdditionalProperties.Schema.Value)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("map[string]%s", valueType), nil
		}

		// Check if this is a reference to a model
		// In kin-openapi, this is actually in SchemaRef, not Schema
		// This should be handled at the caller level

		return "map[string]interface{}", nil
	default:
		return "interface{}", nil
	}
}

// MapParameterTypeToGo maps an OpenAPI parameter type to a Go type
func MapParameterTypeToGo(param *openapi3.Parameter) string {
	if param.Schema == nil || param.Schema.Value == nil {
		return "string"
	}

	schema := param.Schema.Value

	switch schema.Type {
	case "string":
		return "string"
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	default:
		return "string"
	}
}

// GetTestValueForProperty generates appropriate test values for different property types
func GetTestValueForProperty(schema *openapi3.Schema) string {
	switch schema.Type {
	case "string":
		switch schema.Format {
		case "date-time":
			return "time.Now()"
		case "uuid":
			return "\"00000000-0000-0000-0000-000000000000\""
		case "email":
			return "\"test@example.com\""
		case "uri":
			return "\"https://example.com\""
		default:
			if len(schema.Enum) > 0 {
				// Use first enum value
				return fmt.Sprintf("%q", schema.Enum[0])
			}
			return "\"test-string\""
		}
	case "integer":
		return "42"
	case "number":
		if schema.Format == "float" {
			return "42.5"
		}
		return "42.0"
	case "boolean":
		return "true"
	case "array":
		if schema.Items != nil && schema.Items.Value != nil {
			itemValue := GetTestValueForProperty(schema.Items.Value)
			itemType, _ := MapSchemaToGoType(schema.Items.Value)
			return fmt.Sprintf("[]%s{%s}", itemType, itemValue)
		}
		return "[]interface{}{}"
	case "object":
		return "map[string]interface{}{\"test\": \"value\"}"
	default:
		return "nil"
	}
}
