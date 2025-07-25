package generator

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/zeek-r/goapigen/internal/config"
	"github.com/zeek-r/goapigen/internal/parser"
)

// MainGenerator generates the main.go file for the application
type MainGenerator struct {
	parser       *parser.OpenAPIParser
	templateFS   embed.FS
	importPath   string
	mongoURI     string
	dbName       string
	defaultPort  string
	shutdownTime int
}

// MainResourceData holds data for each API resource in main.go
type MainResourceData struct {
	Name           string // Resource name (e.g., Pet)
	VarName        string // Variable name (e.g., pet)
	CollectionName string // MongoDB collection name (e.g., pets)
	APIPath        string // API path (e.g., /pets)
}

// MainTemplateData holds data for the main.go template
type MainTemplateData struct {
	ImportPath         string             // Import path for packages
	UseGeneratedRoutes bool               // Whether to use generated routes
	Resources          []MainResourceData // Resources to be included in the router
	DefaultPort        string             // Default port for the server
	ShutdownTimeout    int                // Shutdown timeout in seconds
	MongoURI           string             // MongoDB URI
	DBName             string             // MongoDB database name
}

// NewMainGenerator creates a new MainGenerator
func NewMainGenerator(
	parser *parser.OpenAPIParser,
	importPath string,
	templateFS embed.FS,
) (*MainGenerator, error) {
	return &MainGenerator{
		parser:       parser,
		templateFS:   templateFS,
		importPath:   importPath,
		mongoURI:     "mongodb://localhost:27017",
		dbName:       "api",
		defaultPort:  "8080",
		shutdownTime: 10,
	}, nil
}

// SetMongoURI sets the MongoDB URI
func (g *MainGenerator) SetMongoURI(uri string) {
	g.mongoURI = uri
}

// SetDBName sets the database name
func (g *MainGenerator) SetDBName(name string) {
	g.dbName = name
}

// SetDefaultPort sets the default port
func (g *MainGenerator) SetDefaultPort(port string) {
	g.defaultPort = port
}

// SetShutdownTimeout sets the shutdown timeout in seconds
func (g *MainGenerator) SetShutdownTimeout(seconds int) {
	g.shutdownTime = seconds
}

// GenerateMain generates the main.go file
func (g *MainGenerator) GenerateMain() (string, error) {
	// Load template
	tmpl, err := template.ParseFS(g.templateFS, config.MainTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse main template: %w", err)
	}

	// Build resource data from schemas
	resources := make([]MainResourceData, 0)
	schemas := g.parser.GetSchemas()

	for name := range schemas {
		varName := strings.ToLower(name)
		collectionName := varName + "s" // Simple pluralization
		apiPath := varName + "s"        // Simple pluralization for API path

		resources = append(resources, MainResourceData{
			Name:           name,
			VarName:        varName,
			CollectionName: collectionName,
			APIPath:        apiPath,
		})
	}

	// Create template data
	data := MainTemplateData{
		ImportPath:         g.importPath,
		UseGeneratedRoutes: true,
		Resources:          resources,
		DefaultPort:        g.defaultPort,
		ShutdownTimeout:    g.shutdownTime,
		MongoURI:           g.mongoURI,
		DBName:             g.dbName,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute main template: %w", err)
	}

	return buf.String(), nil
}
