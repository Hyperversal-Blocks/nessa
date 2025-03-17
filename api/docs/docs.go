package docs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
)

// APISpec represents a minimal structure for schema definitions
type APISpec struct {
	Components map[string]Schema `json:"definitions"`
}

// Schema represents a type definition in Swagger 2.0
type Schema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required,omitempty"`
}

// Property represents a field in a schema
type Property struct {
	Type        string       `json:"type"`
	Format      string       `json:"format,omitempty"`
	Description string       `json:"description,omitempty"`
	Ref         string       `json:"$ref,omitempty"`
	Items       *PropertyRef `json:"items,omitempty"`
}

// PropertyRef is used for array items references
type PropertyRef struct {
	Ref  string `json:"$ref,omitempty"`
	Type string `json:"type,omitempty"`
}

// SwaggerObject represents a Swagger 2.0 specification.
type SwaggerObject struct {
	Swagger             string                 `json:"swagger"`
	Info                InfoObject             `json:"info"`
	Host                string                 `json:"host"`
	BasePath            string                 `json:"basePath"`
	Schemes             []string               `json:"schemes"`
	Consumes            []string               `json:"consumes"`
	Produces            []string               `json:"produces"`
	Paths               map[string]PathItem    `json:"paths"`
	SecurityDefinitions map[string]SecurityDef `json:"securityDefinitions,omitempty"`
	Definitions         map[string]Schema      `json:"definitions,omitempty"`
}

// InfoObject represents metadata about the API.
type InfoObject struct {
	Description string  `json:"description"`
	Title       string  `json:"title"`
	Contact     Contact `json:"contact"`
	License     License `json:"license"`
	Version     string  `json:"version"`
}

// Contact represents basic contact information in the API spec.
type Contact struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// License represents licensing information in the API spec.
type License struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// SecurityDef defines a security scheme (e.g. API key, Bearer token).
type SecurityDef struct {
	Type        string `json:"type"`
	Name        string `json:"name,omitempty"`
	In          string `json:"in,omitempty"`
	Description string `json:"description,omitempty"`
}

// PathItem describes the operations available on a single path.
type PathItem map[string]Operation

// Operation describes a single API operation on a path.
type Operation struct {
	Summary     string                `json:"summary"`
	Description string                `json:"description"`
	Parameters  []Parameter           `json:"parameters,omitempty"`
	Responses   map[string]Response   `json:"responses"`
	Tags        []string              `json:"tags,omitempty"`
	Consumes    []string              `json:"consumes,omitempty"`
	Produces    []string              `json:"produces,omitempty"`
	Security    []map[string][]string `json:"security,omitempty"`
}

// Parameter describes a single parameter for an operation.
type Parameter struct {
	Name        string     `json:"name"`
	In          string     `json:"in"`
	Required    bool       `json:"required"`
	Type        string     `json:"type,omitempty"`
	Schema      *SchemaRef `json:"schema,omitempty"`
	Description string     `json:"description"`
}

// Response describes a single response from an API Operation.
type Response struct {
	Description string     `json:"description"`
	Schema      *SchemaRef `json:"schema,omitempty"`
}

// SchemaRef represents a reference to a schema definition
type SchemaRef struct {
	Ref string `json:"$ref,omitempty"`
}

func BuildOpenAPISpec(router *mux.Router) (*SwaggerObject, error) {
	paths := make(map[string]PathItem)

	// Extract API structs first so we can match them with endpoints
	dir, _ := os.Getwd()
	structs, err := ExtractAPIStructs(filepath.Join(dir, "../"))
	if err != nil {
		fmt.Printf("Warning: Could not extract API structs: %v\n", err)
	}

	// Clean struct tags
	for i := range structs {
		for j := range structs[i].Fields {
			if structs[i].Fields[j].Tags != "" {
				tags := structs[i].Fields[j].Tags
				tags = strings.Trim(tags, "`")
				tags = strings.ReplaceAll(tags, "\\\"", "")
				tags = strings.ReplaceAll(tags, "\"", "")
				structs[i].Fields[j].Tags = tags
			}
		}
	}

	// Create a map of request and response types for easy lookup
	requestTypes := make(map[string]string)
	responseTypes := make(map[string]string)

	for _, s := range structs {
		name := strings.ToLower(s.Name)
		if strings.Contains(name, "request") || strings.Contains(name, "req") {
			// Extract the base name (e.g., "createuser" from "createuserrequest")
			baseName := strings.TrimSuffix(strings.TrimSuffix(name, "request"), "req")
			requestTypes[baseName] = s.Name
		} else if strings.Contains(name, "response") || strings.Contains(name, "resp") {
			baseName := strings.TrimSuffix(strings.TrimSuffix(name, "response"), "resp")
			responseTypes[baseName] = s.Name
		}
	}

	re := regexp.MustCompile(`\{([^}]+)\}`)

	err = router.Walk(func(route *mux.Route, r *mux.Router, ancestors []*mux.Route) error {
		tpl, err := route.GetPathTemplate()
		if err != nil || tpl == "" {
			return nil
		}

		methods, err := route.GetMethods()
		if err != nil || len(methods) == 0 {
			return nil
		}

		if _, exists := paths[tpl]; !exists {
			paths[tpl] = make(PathItem)
		}

		matches := re.FindAllStringSubmatch(tpl, -1)
		var parameters []Parameter
		for _, match := range matches {
			if len(match) == 2 {
				paramName := match[1]
				parameters = append(parameters, Parameter{
					Name:        paramName,
					In:          "path",
					Required:    true,
					Type:        "string",
					Description: "Auto-detected path parameter",
				})
			}
		}

		routeName := route.GetName()

		for _, m := range methods {
			lowerMethod := strings.ToLower(m)
			op := Operation{
				Summary:     "Endpoint " + tpl,
				Description: "Automatically generated endpoint for " + tpl,
				Parameters:  parameters,
				Responses: map[string]Response{
					"200": {Description: "OK"},
					"400": {Description: "Bad Request"},
					"500": {Description: "Internal Server Error"},
				},
				Tags:     []string{},
				Consumes: []string{"application/json"},
				Produces: []string{"application/json"},
				Security: []map[string][]string{
					{"ApiKeyAuth": {}},
				},
			}

			// Only add request body for methods that typically include one
			needsRequestBody := lowerMethod == "post" || lowerMethod == "put" || lowerMethod == "patch"

			if routeName != "" {
				op.Summary = routeName
				op.Tags = []string{routeName}

				// Try to match the route name with request/response types
				routeNameLower := strings.ToLower(routeName)

				// If the method requires a request body, try to find a matching request type
				if needsRequestBody {
					foundRequestType := false

					// First try exact match by route name
					requestType, exists := requestTypes[routeNameLower]
					if exists {
						op.Parameters = append(op.Parameters, Parameter{
							Name:        "body",
							In:          "body",
							Required:    true,
							Schema:      &SchemaRef{Ref: "#/definitions/" + requestType},
							Description: requestType + " object",
						})
						foundRequestType = true
					} else {
						// Try partial matching if exact match failed
						for prefix, reqTypeName := range requestTypes {
							if strings.HasPrefix(routeNameLower, prefix) {
								op.Parameters = append(op.Parameters, Parameter{
									Name:        "body",
									In:          "body",
									Required:    true,
									Schema:      &SchemaRef{Ref: "#/definitions/" + reqTypeName},
									Description: reqTypeName + " object",
								})
								foundRequestType = true
								break
							}
						}
					}

					// If we still don't have a request type, make an educated guess based on path
					if !foundRequestType {
						// Extract last part of URL path as potential model name
						pathParts := strings.Split(strings.Trim(tpl, "/"), "/")
						if len(pathParts) > 0 {
							lastPathPart := pathParts[len(pathParts)-1]
							// Remove URL parameters if any
							lastPathPart = strings.Split(lastPathPart, "{")[0]

							// Try to find a request type that contains this path part
							for _, reqTypeName := range requestTypes {
								if strings.Contains(strings.ToLower(reqTypeName), strings.ToLower(lastPathPart)) {
									op.Parameters = append(op.Parameters, Parameter{
										Name:        "body",
										In:          "body",
										Required:    true,
										Schema:      &SchemaRef{Ref: "#/definitions/" + reqTypeName},
										Description: reqTypeName + " object",
									})
									foundRequestType = true
									break
								}
							}
						}
					}
				}

				// Try to find a matching response type
				// Check for exact match first
				responseType, exists := responseTypes[routeNameLower]
				if exists {
					op.Responses["200"] = Response{
						Description: "Successful operation",
						Schema:      &SchemaRef{Ref: "#/definitions/" + responseType},
					}
				} else {
					// Try partial matching if exact match failed
					for prefix, respTypeName := range responseTypes {
						if strings.HasPrefix(routeNameLower, prefix) {
							op.Responses["200"] = Response{
								Description: "Successful operation",
								Schema:      &SchemaRef{Ref: "#/definitions/" + respTypeName},
							}
							break
						}
					}
				}
			} else if needsRequestBody {
				// For unnamed routes that need a body, try to guess from the path
				pathParts := strings.Split(strings.Trim(tpl, "/"), "/")
				if len(pathParts) > 0 {
					lastPathPart := pathParts[len(pathParts)-1]
					// Remove URL parameters if any
					lastPathPart = strings.Split(lastPathPart, "{")[0]

					// Check if we have any request type that matches this path part
					for _, reqTypeName := range requestTypes {
						if strings.Contains(strings.ToLower(reqTypeName), strings.ToLower(lastPathPart)) {
							op.Parameters = append(op.Parameters, Parameter{
								Name:        "body",
								In:          "body",
								Required:    true,
								Schema:      &SchemaRef{Ref: "#/definitions/" + reqTypeName},
								Description: reqTypeName + " object",
							})
							break
						}
					}
				}
			}

			paths[tpl][lowerMethod] = op
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Create the Swagger spec object with all the information
	spec := &SwaggerObject{
		Swagger: "2.0",
		Info: InfoObject{
			Description: "Automatically generated API documentation.",
			Title:       "Keyless-Auth",
			Contact: Contact{
				Name: "Hackathon@Encode",
				URL:  "https://example.com",
			},
			License: License{
				Name: "MIT",
				URL:  "https://opensource.org/licenses/MIT",
			},
			Version: "1.1.1",
		},
		Host:     getEnv("API_HOST", "localhost:8081"),
		BasePath: "/",
		Schemes:  []string{"http"},
		Consumes: []string{"application/json"},
		Produces: []string{"application/json"},
		Paths:    paths,
		SecurityDefinitions: map[string]SecurityDef{
			"ApiKeyAuth": {
				Type:        "apiKey",
				Name:        "X-API-KEY",
				In:          "header",
				Description: "Use an API key to authenticate requests.",
			},
		},
		Definitions: make(map[string]Schema),
	}

	// Add schema definitions
	if err == nil {
		for _, s := range structs {
			schema := Schema{
				Type:       "object",
				Properties: make(map[string]Property),
			}

			for _, field := range s.Fields {
				propName := field.Name

				// Use tag name if available
				if field.Tags != "" {
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
				case strings.HasPrefix(field.Type, "*"):
					baseType := strings.TrimPrefix(field.Type, "*")
					mappedType := mapGoTypeToOpenAPIType(baseType)
					prop.Type = mappedType
				default:
					if isUpperCase(field.Type) {
						prop.Ref = "#/definitions/" + field.Type
					} else {
						prop.Type = "object"
					}
				}

				if field.Comment != "" {
					prop.Description = field.Comment
				}

				schema.Properties[propName] = prop
			}

			spec.Definitions[s.Name] = schema
		}
	}

	jsonSpec, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return nil, err
	}

	// Create the full path to api/docs directory
	docsDir := filepath.Join(dir, "api", "docs")

	// Ensure the directory exists
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return nil, err
	}

	// Create the file in the correct location
	dst, err := os.Create(filepath.Join(docsDir, "doc.json"))
	if err != nil {
		return nil, err
	}

	defer dst.Close()

	if _, err = io.Copy(dst, bytes.NewReader(jsonSpec)); err != nil {
		return nil, err
	}
	return spec, nil
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

type StructField struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Tags    string `json:"tags,omitempty"`
	Comment string `json:"comment,omitempty"`
}

type StructDef struct {
	Name   string        `json:"name"`
	Fields []StructField `json:"fields"`
}

func ExtractAPIStructs(pkgDir string) ([]StructDef, error) {
	var results []StructDef
	keywords := []string{"resp", "response", "req", "request"}

	// Walk through all .go files in the directory
	err := filepath.Walk(pkgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Parse the Go file
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("error parsing %s: %v", path, err)
		}

		// Inspect the AST and find struct declarations
		ast.Inspect(node, func(n ast.Node) bool {
			// Check if node is a type spec (type declaration)
			typeSpec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}

			// Check if it's a struct
			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				return true
			}

			// Check if struct name contains any of our keywords
			structName := typeSpec.Name.Name
			containsKeyword := false
			lowerName := strings.ToLower(structName)
			for _, keyword := range keywords {
				if strings.Contains(lowerName, keyword) {
					containsKeyword = true
					break
				}
			}

			if !containsKeyword {
				return true
			}

			// Extract struct information
			structDef := StructDef{
				Name: structName,
			}

			// Extract fields
			if structType.Fields != nil {
				for _, field := range structType.Fields.List {
					for _, name := range field.Names {
						fieldDef := StructField{
							Name: name.Name,
							Type: formatType(field.Type),
						}

						// Extract field tags if any
						if field.Tag != nil {
							fieldDef.Tags = field.Tag.Value
						}

						// Extract field comment if any
						if field.Comment != nil {
							fieldDef.Comment = field.Comment.Text()
						}

						structDef.Fields = append(structDef.Fields, fieldDef)
					}
				}
			}

			results = append(results, structDef)
			return true
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// Helper function to format the type of a field
func formatType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatType(t.X)
	case *ast.ArrayType:
		return "[]" + formatType(t.Elt)
	case *ast.MapType:
		return "map[" + formatType(t.Key) + "]" + formatType(t.Value)
	case *ast.SelectorExpr:
		return formatType(t.X) + "." + t.Sel.Name
	default:
		return fmt.Sprintf("%T", expr) // fallback
	}
}

// Helper function to map Go types to OpenAPI types
func mapGoTypeToOpenAPIType(goType string) string {
	switch {
	case strings.HasPrefix(goType, "string"):
		return "string"
	case strings.HasPrefix(goType, "int"), strings.HasPrefix(goType, "uint"):
		return "integer"
	case strings.HasPrefix(goType, "float"):
		return "number"
	case strings.HasPrefix(goType, "bool"):
		return "boolean"
	default:
		return "object"
	}
}

// Helper to check if first character is uppercase
func isUpperCase(s string) bool {
	if len(s) == 0 {
		return false
	}
	firstChar := s[0]
	return firstChar >= 'A' && firstChar <= 'Z'
}
