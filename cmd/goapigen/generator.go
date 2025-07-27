package cli

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/zeek-r/goapigen/internal/config"
	"github.com/zeek-r/goapigen/internal/generator"
	"github.com/zeek-r/goapigen/internal/parser"
)

// GenerationConfig holds all configuration for code generation
type GenerationConfig struct {
	SpecFile    string
	OutputDir   string
	PackageName string
	HTTPPackage string
	SchemaName  string

	// Generation flags
	GenTypes    bool
	GenServices bool
	GenMongo    bool
	GenHTTP     bool
	InitProject bool
	Overwrite   bool
}

// GenerationPipeline handles the complete code generation process
type GenerationPipeline struct {
	config       *GenerationConfig
	parser       *parser.OpenAPIParser
	templateFS   embed.FS
	targetModule string
	importPath   string
}

// NewGenerationPipeline creates a new generation pipeline
func NewGenerationPipeline(config *GenerationConfig, templateFS embed.FS) (*GenerationPipeline, error) {
	// Parse OpenAPI spec
	apiParser, err := parser.NewOpenAPIParser(config.SpecFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing OpenAPI spec: %w", err)
	}

	// Create output directory
	if err := os.MkdirAll(config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("error creating output directory: %w", err)
	}

	// Initialize or detect Go module
	targetModule, err := initializeGoModule(config.OutputDir)
	if err != nil {
		return nil, fmt.Errorf("error with Go module: %w", err)
	}

	return &GenerationPipeline{
		config:       config,
		parser:       apiParser,
		templateFS:   templateFS,
		targetModule: targetModule,
		importPath:   targetModule,
	}, nil
}

// Execute runs the complete generation pipeline
func (p *GenerationPipeline) Execute() error {
	// Get schemas to generate
	schemaNames, err := p.getSchemaNames()
	if err != nil {
		return err
	}

	// Initialize project structure if requested
	if p.config.InitProject {
		if err := p.initializeProject(); err != nil {
			return fmt.Errorf("error initializing project: %w", err)
		}
	}

	// Generate types
	if p.config.GenTypes {
		if err := p.generateTypes(); err != nil {
			return fmt.Errorf("error generating types: %w", err)
		}
	}

	// Generate services
	if p.config.GenServices || p.config.GenHTTP {
		if err := p.generateServices(schemaNames); err != nil {
			return fmt.Errorf("error generating services: %w", err)
		}
	}

	// Generate MongoDB repositories
	if p.config.GenMongo {
		if err := p.generateMongoRepositories(schemaNames); err != nil {
			return fmt.Errorf("error generating repositories: %w", err)
		}
	}

	// Generate HTTP handlers
	if p.config.GenHTTP {
		if err := p.generateHTTPHandlers(); err != nil {
			return fmt.Errorf("error generating HTTP handlers: %w", err)
		}
	}

	// Regenerate routes if needed
	if p.config.GenServices || p.config.GenMongo || p.config.GenHTTP || p.config.InitProject {
		if err := p.regenerateRoutes(); err != nil {
			return fmt.Errorf("error regenerating routes: %w", err)
		}
	}

	// Generate environment file
	if p.config.InitProject {
		if err := p.generateEnvFile(); err != nil {
			return fmt.Errorf("error generating .env file: %w", err)
		}
	}

	return nil
}

// initializeGoModule initializes or detects the Go module
func initializeGoModule(outputDir string) (string, error) {
	goModPath := filepath.Join(outputDir, config.GoModFile)

	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		// No go.mod found, run go mod init
		fmt.Printf("No go.mod found in output directory. Running go mod init...\n")

		moduleName := filepath.Base(outputDir)
		cmd := exec.Command("go", "mod", "init", moduleName)
		cmd.Dir = outputDir
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("error running go mod init: %w", err)
		}
	}

	// Read module name from go.mod
	return getModuleNameFromPath(outputDir)
}

// getSchemaNames returns the schemas to generate code for
func (p *GenerationPipeline) getSchemaNames() ([]string, error) {
	if p.config.SchemaName != "" {
		if _, exists := p.parser.GetSchemaByName(p.config.SchemaName); exists {
			return []string{p.config.SchemaName}, nil
		}
		return nil, fmt.Errorf("schema %q not found in spec", p.config.SchemaName)
	}

	// Return all schemas
	schemas := p.parser.GetSchemas()
	schemaNames := make([]string, 0, len(schemas))
	for name := range schemas {
		schemaNames = append(schemaNames, name)
	}
	return schemaNames, nil
}

// initializeProject creates the basic project structure and main files
func (p *GenerationPipeline) initializeProject() error {
	fmt.Println("Initializing project structure...")

	// Create directory structure
	dirsToCreate := []string{
		p.config.OutputDir,
		filepath.Join(p.config.OutputDir, config.InternalDir),
		filepath.Join(p.config.OutputDir, config.PkgDir),
		filepath.Join(p.config.OutputDir, config.DomainDir),
		filepath.Join(p.config.OutputDir, config.ServicesDir),
		filepath.Join(p.config.OutputDir, config.AdaptersDir),
		filepath.Join(p.config.OutputDir, config.HttpAdaptersDir),
		filepath.Join(p.config.OutputDir, config.MongoAdaptersDir),
		filepath.Join(p.config.OutputDir, config.HttpUtilDir),
	}

	for _, dir := range dirsToCreate {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating directory %s: %w", dir, err)
		}
	}

	// Generate main application files
	return p.generateMainFiles()
}

// generateMainFiles generates main.go, routes.go, and database.go
func (p *GenerationPipeline) generateMainFiles() error {
	mainGen, err := generator.NewMainGenerator(p.parser, p.importPath, p.templateFS)
	if err != nil {
		return fmt.Errorf("error creating main generator: %w", err)
	}

	// Configure generator
	mainGen.SetMongoURI("mongodb://localhost:27017")
	mainGen.SetDBName(filepath.Base(p.targetModule))
	mainGen.SetDefaultPort("8080")

	// Create cmd directory
	cmdDir := filepath.Join(p.config.OutputDir, "cmd", filepath.Base(p.targetModule))
	if err := os.MkdirAll(cmdDir, 0755); err != nil {
		return fmt.Errorf("error creating cmd directory: %w", err)
	}

	// Generate files with current features
	hasServices := p.config.GenServices || p.config.GenHTTP
	files, err := mainGen.GenerateWithFeatures(p.config.GenMongo, p.config.GenMongo, hasServices, p.config.GenHTTP)
	if err != nil {
		return fmt.Errorf("error generating main files: %w", err)
	}

	// Write files
	return p.writeMainFiles(cmdDir, files)
}

// writeMainFiles writes the generated main files to disk
func (p *GenerationPipeline) writeMainFiles(cmdDir string, files map[string]string) error {
	for filename, content := range files {
		filePath := filepath.Join(cmdDir, filename)

		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("error creating directory for %s: %w", filename, err)
		}

		// For main.go, only write if it doesn't exist (stable file)
		// For routes.go and database.go, always overwrite (regenerated files)
		shouldWrite := p.config.Overwrite
		if strings.HasSuffix(filename, "routes.go") || strings.HasSuffix(filename, "database.go") {
			shouldWrite = true
		} else if _, err := os.Stat(filePath); os.IsNotExist(err) {
			shouldWrite = true
		}

		if shouldWrite {
			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("error writing %s: %w", filename, err)
			}
			fmt.Printf("Generated %s\n", filePath)
		} else {
			fmt.Printf("%s already exists. Skipping (use --overwrite to force overwrite)\n", filename)
		}
	}

	// Add required dependencies
	return p.addDependencies()
}

// addDependencies adds required Go module dependencies
func (p *GenerationPipeline) addDependencies() error {
	fmt.Println("Adding required dependencies...")
	deps := []string{
		"github.com/go-chi/chi/v5",
		"github.com/go-chi/cors",
		"github.com/joho/godotenv",
		"go.mongodb.org/mongo-driver/mongo",
	}

	for _, dep := range deps {
		cmd := exec.Command("go", "get", dep)
		cmd.Dir = p.config.OutputDir
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error adding dependency %s: %v\n", dep, err)
		}
	}
	return nil
}

// generateTypes generates domain types from OpenAPI schemas
func (p *GenerationPipeline) generateTypes() error {
	// Implementation extracted from main.go - for now, return nil
	// TODO: Extract types generation logic from main.go
	return nil
}

// generateServices generates service layer for schemas
func (p *GenerationPipeline) generateServices(schemaNames []string) error {
	// Implementation extracted from main.go - for now, return nil
	// TODO: Extract services generation logic from main.go
	return nil
}

// generateMongoRepositories generates MongoDB repositories
func (p *GenerationPipeline) generateMongoRepositories(schemaNames []string) error {
	// Implementation extracted from main.go - for now, return nil
	// TODO: Extract mongo repositories generation logic from main.go
	return nil
}

// generateHTTPHandlers generates HTTP handlers
func (p *GenerationPipeline) generateHTTPHandlers() error {
	// Implementation extracted from main.go - for now, return nil
	// TODO: Extract HTTP handlers generation logic from main.go
	return nil
}

// regenerateRoutes regenerates routes.go when components change
func (p *GenerationPipeline) regenerateRoutes() error {
	// Implementation extracted from main.go - for now, return nil
	// TODO: Extract routes regeneration logic from main.go
	return nil
}

// generateEnvFile generates the .env configuration file
func (p *GenerationPipeline) generateEnvFile() error {
	// Implementation extracted from main.go - for now, return nil
	// TODO: Extract env file generation logic from main.go
	return nil
}
