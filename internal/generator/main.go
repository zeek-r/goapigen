package generator

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

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
	HasRepository  bool   // Whether repository is generated for this resource
	HasService     bool   // Whether service is generated for this resource
	HasHandler     bool   // Whether handler is generated for this resource
}

// MainTemplateData holds data for the main.go template
type MainTemplateData struct {
	ImportPath      string             // Import path for packages
	UseMongo        bool               // Whether MongoDB is used
	HasResources    bool               // Whether any resources are defined
	Resources       []MainResourceData // Resources to be included in the router
	DefaultPort     string             // Default port for the server
	ShutdownTimeout int                // Shutdown timeout in seconds
	MongoURI        string             // MongoDB URI
	DBName          string             // MongoDB database name
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

// GenerateMain generates both main.go and routes.go files
func (g *MainGenerator) GenerateMain() (map[string]string, error) {
	result := make(map[string]string)

	// Generate stable main.go
	mainCode, err := g.GenerateMainFile(false, false, false, false) // Basic server only
	if err != nil {
		return nil, fmt.Errorf("failed to generate main.go: %w", err)
	}
	result["main.go"] = mainCode

	// Generate routes.go with basic health check only
	routesCode, err := g.GenerateRoutesFile(false, false, false, false)
	if err != nil {
		return nil, fmt.Errorf("failed to generate routes.go: %w", err)
	}
	result["routes.go"] = routesCode

	// Generate database.go with basic setup
	databaseCode, err := g.GenerateDatabaseFile(false, false, false, false)
	if err != nil {
		return nil, fmt.Errorf("failed to generate database.go: %w", err)
	}
	result["database.go"] = databaseCode

	return result, nil
}

// GenerateMainFile generates the stable main.go file
func (g *MainGenerator) GenerateMainFile(useMongo, hasRepo, hasServices, hasHandler bool) (string, error) {
	// Load template
	tmpl, err := template.ParseFS(g.templateFS, "templates/cmd/main.go.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to parse main template: %w", err)
	}

	// Build resource data from schemas for conditional logic
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
			HasRepository:  hasRepo,
			HasService:     hasServices,
			HasHandler:     hasHandler,
		})
	}

	// Create template data
	data := MainTemplateData{
		ImportPath:      g.importPath,
		UseMongo:        useMongo,
		HasResources:    len(resources) > 0,
		Resources:       resources,
		DefaultPort:     g.defaultPort,
		ShutdownTimeout: g.shutdownTime,
		MongoURI:        g.mongoURI,
		DBName:          g.dbName,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute main template: %w", err)
	}

	return buf.String(), nil
}

// GenerateRoutesFile generates the routes.go file with conditional imports
func (g *MainGenerator) GenerateRoutesFile(useMongo, hasRepo, hasServices, hasHandler bool) (string, error) {
	// Load template
	tmpl, err := template.ParseFS(g.templateFS, "templates/cmd/routes.go.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to parse routes template: %w", err)
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
			HasRepository:  hasRepo,
			HasService:     hasServices,
			HasHandler:     hasHandler,
		})
	}

	// Create template data
	data := MainTemplateData{
		ImportPath:      g.importPath,
		UseMongo:        useMongo,
		HasResources:    len(resources) > 0,
		Resources:       resources,
		DefaultPort:     g.defaultPort,
		ShutdownTimeout: g.shutdownTime,
		MongoURI:        g.mongoURI,
		DBName:          g.dbName,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute routes template: %w", err)
	}

	return buf.String(), nil
}

// GenerateDatabaseFile generates the database.go file with dependency injection
func (g *MainGenerator) GenerateDatabaseFile(useMongo, hasRepo, hasServices, hasHandler bool) (string, error) {
	// Load template
	tmpl, err := template.ParseFS(g.templateFS, "templates/cmd/database.go.tmpl")
	if err != nil {
		return "", fmt.Errorf("failed to parse database template: %w", err)
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
			HasRepository:  hasRepo,
			HasService:     hasServices,
			HasHandler:     hasHandler,
		})
	}

	// Create template data
	data := MainTemplateData{
		ImportPath:      g.importPath,
		UseMongo:        useMongo,
		HasResources:    len(resources) > 0,
		Resources:       resources,
		DefaultPort:     g.defaultPort,
		ShutdownTimeout: g.shutdownTime,
		MongoURI:        g.mongoURI,
		DBName:          g.dbName,
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute database template: %w", err)
	}

	return buf.String(), nil
}

// GenerateWithFeatures generates both files with specific feature flags
func (g *MainGenerator) GenerateWithFeatures(useMongo, hasRepo, hasServices, hasHandler bool) (map[string]string, error) {
	result := make(map[string]string)

	// Generate main.go with appropriate features
	mainCode, err := g.GenerateMainFile(useMongo, hasRepo, hasServices, hasHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to generate main.go: %w", err)
	}
	result["main.go"] = mainCode

	// Generate routes.go with conditional imports
	routesCode, err := g.GenerateRoutesFile(useMongo, hasRepo, hasServices, hasHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to generate routes.go: %w", err)
	}
	result["routes.go"] = routesCode

	// Generate database.go with dependency injection
	databaseCode, err := g.GenerateDatabaseFile(useMongo, hasRepo, hasServices, hasHandler)
	if err != nil {
		return nil, fmt.Errorf("failed to generate database.go: %w", err)
	}
	result["database.go"] = databaseCode

	return result, nil
}
