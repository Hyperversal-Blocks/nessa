package docs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Swagger/OpenAPI compatible structures

func TestGetAllJsonTags(t *testing.T) {
	t.Parallel()

	t.Run("challenger", func(t *testing.T) {
		dir, _ := os.Getwd()
		structs, err := ExtractAPIStructs(filepath.Join(dir, "../ephemeral.go"))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Clean struct tags by removing backticks and escape characters
		for i := range structs {
			for j := range structs[i].Fields {
				// Remove backticks and clean up escaped quotes
				if structs[i].Fields[j].Tags != "" {
					tags := structs[i].Fields[j].Tags
					// Remove backticks
					tags = strings.Trim(tags, "`")
					// Remove all quotes (both escaped and regular)
					tags = strings.ReplaceAll(tags, "\\\"", "")
					tags = strings.ReplaceAll(tags, "\"", "")
					structs[i].Fields[j].Tags = tags
				}
			}
		}

		// Convert to OpenAPI components format
		openAPISpec := convertToSwaggerDefinitions(structs)

		// Convert to JSON
		jsonData, err := json.MarshalIndent(openAPISpec, "", "  ")
		if err != nil {
			fmt.Printf("Error marshaling to JSON: %v\n", err)
			return
		}

		// Create the file in the correct location
		dst, err := os.Create(filepath.Join(dir, "swagger-definitions.json"))
		if err != nil {
			t.Fail()
		}

		defer dst.Close()

		if _, err = io.Copy(dst, bytes.NewReader(jsonData)); err != nil {
			t.Fail()
		}
	})
}

// Convert our struct definitions to OpenAPI components format
func convertToSwaggerDefinitions(structs []StructDef) APISpec {
	spec := APISpec{
		Components: make(map[string]Schema),
	}

	for _, s := range structs {
		schema := Schema{
			Type:       "object",
			Properties: make(map[string]Property),
		}

		for _, field := range s.Fields {
			propName := field.Name

			// Use tag name if available
			if field.Tags != "" {
				// Extract the JSON tag name
				tagParts := strings.Split(field.Tags, ",")
				if strings.HasPrefix(tagParts[0], "json:") {
					jsonName := strings.TrimPrefix(tagParts[0], "json:")
					if jsonName != "-" && jsonName != "" {
						propName = jsonName
					}
				}
			}

			prop := Property{}

			// Map Go types to OpenAPI types
			switch {
			case strings.HasPrefix(field.Type, "string"):
				prop.Type = "string"
			case strings.HasPrefix(field.Type, "int"), strings.HasPrefix(field.Type, "uint"):
				prop.Type = "integer"
				if strings.Contains(field.Type, "64") {
					prop.Format = "int64"
				} else if strings.Contains(field.Type, "32") {
					prop.Format = "int32"
				}
			case strings.HasPrefix(field.Type, "float"):
				prop.Type = "number"
				if strings.Contains(field.Type, "64") {
					prop.Format = "double"
				} else if strings.Contains(field.Type, "32") {
					prop.Format = "float"
				}
			case strings.HasPrefix(field.Type, "bool"):
				prop.Type = "boolean"
			case strings.HasPrefix(field.Type, "time.Time"):
				prop.Type = "string"
				prop.Format = "date-time"
			case strings.HasPrefix(field.Type, "[]"):
				prop.Type = "array"
				itemType := strings.TrimPrefix(field.Type, "[]")
				prop.Items = &PropertyRef{Type: mapGoTypeToOpenAPIType(itemType)}
			case strings.HasPrefix(field.Type, "map["):
				prop.Type = "object"
				// Additional properties could be handled here
			case strings.HasPrefix(field.Type, "*"):
				// Handle pointer types
				baseType := strings.TrimPrefix(field.Type, "*")
				mappedType := mapGoTypeToOpenAPIType(baseType)
				prop.Type = mappedType
			default:
				// Could be a reference to another schema
				if isUpperCase(field.Type) {
					prop.Ref = "#/components/schemas/" + field.Type
				} else {
					prop.Type = "object"
				}
			}

			// Add description from comment if available
			if field.Comment != "" {
				prop.Description = field.Comment
			}

			schema.Properties[propName] = prop
		}

		spec.Components[s.Name] = schema
	}

	return spec
}

// // Helper function to map Go types to OpenAPI types
// func mapGoTypeToOpenAPIType(goType string) string {
// 	switch {
// 	case strings.HasPrefix(goType, "string"):
// 		return "string"
// 	case strings.HasPrefix(goType, "int"), strings.HasPrefix(goType, "uint"):
// 		return "integer"
// 	case strings.HasPrefix(goType, "float"):
// 		return "number"
// 	case strings.HasPrefix(goType, "bool"):
// 		return "boolean"
// 	default:
// 		return "object"
// 	}
// }
//
// // Helper to check if first character is uppercase (exported)
// func isUpperCase(s string) bool {
// 	if len(s) == 0 {
// 		return false
// 	}
// 	firstChar := s[0]
// 	return firstChar >= 'A' && firstChar <= 'Z'
// }

type SwaggerComponents struct {
	Definitions struct {
		ChallengeRespondPayload struct {
			Items struct {
				Title string `json:"title"`
				Type  string `json:"type"`
			} `json:"items"`
		} `json:"ChallengeRespondPayload"`
	} `json:"definitions"`
}
