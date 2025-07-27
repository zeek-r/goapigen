package main

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/zeek-r/goapigen/internal/config"
	"github.com/zeek-r/goapigen/internal/generator"
	"github.com/zeek-r/goapigen/internal/parser"
)

//go:embed templates
var templateFS embed.FS

func main() {
	// Command line flags
	var (
		specFile    = flag.String("spec", "", "Path to OpenAPI specification file")
		outputDir   = flag.String("output", ".", "Output directory for generated code")
		packageName = flag.String("package", config.DefaultAPIPackage, "Package name for generated code")
		genTypes    = flag.Bool("types", true, "Generate type definitions")
		genServices = flag.Bool("services", false, "Generate service layer")
		genMongo    = flag.Bool("mongo", false, "Generate MongoDB repositories")
		genHTTP     = flag.Bool("http", false, "Generate HTTP handlers")
		httpPackage = flag.String("http-package", config.DefaultHandlerPackage, "Package name for HTTP handlers")
		schemaName  = flag.String("schema", "", "Generate code for specific schema (if empty, generates for all schemas)")
		initProject = flag.Bool("init", false, "Initialize a new project with full directory structure and main.go")
		overwrite   = flag.Bool("overwrite", false, "Overwrite existing files (default: false)")
	)

	flag.Parse()

	// Validate inputs
	if *specFile == "" {
		fmt.Println("Error: OpenAPI specification file is required")
		flag.Usage()
		os.Exit(1)
	}

	// Parse the OpenAPI spec
	apiParser, err := parser.NewOpenAPIParser(*specFile)
	if err != nil {
		fmt.Printf("Error parsing OpenAPI spec: %v\n", err)
		os.Exit(1)
	}

	// Get all schemas or filter by name
	var schemaNames []string
	if *schemaName != "" {
		if _, exists := apiParser.GetSchemaByName(*schemaName); exists {
			schemaNames = []string{*schemaName}
		} else {
			fmt.Printf("Error: Schema %q not found in spec\n", *schemaName)
			os.Exit(1)
		}
	} else {
		schemas := apiParser.GetSchemas()
		schemaNames = make([]string, 0, len(schemas))
		for name := range schemas {
			schemaNames = append(schemaNames, name)
		}
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	// Check if go.mod exists in output directory, if not run go mod init
	goModPath := filepath.Join(*outputDir, config.GoModFile)
	var targetModuleName string

	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		// No go.mod found, run go mod init
		fmt.Printf("No go.mod found in output directory. Running go mod init...\n")

		// Use the directory name as the module name
		moduleName := filepath.Base(*outputDir)
		cmd := exec.Command("go", "mod", "init", moduleName)
		cmd.Dir = *outputDir
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error running go mod init: %v\n", err)
			os.Exit(1)
		}

		// Read the module name from the created go.mod
		detectedModule, err := getModuleNameFromPath(*outputDir)
		if err != nil {
			fmt.Printf("Warning: couldn't determine Go module name after go mod init: %v. Using fallback.\n", err)
			targetModuleName = "generated-api"
		} else {
			targetModuleName = detectedModule
		}
	} else {
		// go.mod exists, read the module name from it
		detectedModule, err := getModuleNameFromPath(*outputDir)
		if err != nil {
			fmt.Printf("Warning: couldn't determine Go module name from %s: %v. Using fallback.\n", goModPath, err)
			targetModuleName = "generated-api"
		} else {
			targetModuleName = detectedModule
		}
	}

	// Build import paths
	importPath := targetModuleName

	// Initialize full project structure if requested
	if *initProject {
		fmt.Println("Initializing project structure...")

		// Create all necessary directories
		dirsToCreate := []string{
			*outputDir,
			filepath.Join(*outputDir, config.InternalDir),
			filepath.Join(*outputDir, config.PkgDir),
			filepath.Join(*outputDir, config.DomainDir),
			filepath.Join(*outputDir, config.ServicesDir),
			filepath.Join(*outputDir, config.AdaptersDir),
			filepath.Join(*outputDir, config.HttpAdaptersDir),
			filepath.Join(*outputDir, config.MongoAdaptersDir),
			filepath.Join(*outputDir, config.HttpUtilDir),
		}

		for _, dir := range dirsToCreate {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("Error creating directory %s: %v\n", dir, err)
				os.Exit(1)
			}
		}

		// Generate main.go file
		mainGen, err := generator.NewMainGenerator(apiParser, importPath, templateFS)
		if err != nil {
			fmt.Printf("Error creating main generator: %v\n", err)
			os.Exit(1)
		}

		// Configure main generator
		mainGen.SetMongoURI("mongodb://localhost:27017")   // Default for now, will be generated in .env
		mainGen.SetDBName(filepath.Base(targetModuleName)) // Default for now, will be generated in .env
		mainGen.SetDefaultPort("8080")                     // Default for now, will be generated in .env

		// Create cmd directory
		cmdDir := filepath.Join(*outputDir, "cmd", filepath.Base(targetModuleName))
		if err := os.MkdirAll(cmdDir, 0755); err != nil {
			fmt.Printf("Error creating cmd directory: %v\n", err)
			os.Exit(1)
		}

		// Generate main.go and routes.go files using basic init (no features)
		files, err := mainGen.GenerateMain()
		if err != nil {
			fmt.Printf("Error generating main files: %v\n", err)
			os.Exit(1)
		}

		// Write files
		for filename, content := range files {
			filePath := filepath.Join(cmdDir, filename)

			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				fmt.Printf("Error creating directory for %s: %v\n", filename, err)
				continue
			}

			// For main.go, only write if it doesn't exist (stable file)
			// For routes.go, always overwrite (regenerated file)
			shouldWrite := *overwrite
			if strings.HasSuffix(filename, "routes.go") {
				shouldWrite = true // Always regenerate routes
			} else if _, err := os.Stat(filePath); os.IsNotExist(err) {
				shouldWrite = true // Write if doesn't exist
			}

			if shouldWrite {
				if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
					fmt.Printf("Error writing %s: %v\n", filename, err)
					continue
				}
				fmt.Printf("Generated %s\n", filePath)
			} else {
				fmt.Printf("%s already exists. Skipping (use --overwrite to force overwrite)\n", filename)
			}
		}

		// Add dependencies to go.mod
		fmt.Println("Adding required dependencies...")
		deps := []string{
			"github.com/go-chi/chi/v5",
			"github.com/go-chi/cors",
			"github.com/joho/godotenv",
			"go.mongodb.org/mongo-driver/mongo",
		}

		for _, dep := range deps {
			cmd := exec.Command("go", "get", dep)
			cmd.Dir = *outputDir
			if err := cmd.Run(); err != nil {
				fmt.Printf("Error adding dependency %s: %v\n", dep, err)
			}
		}
	}

	// Generate types if requested
	if *genTypes {
		// Create domain directory
		domainDir := filepath.Join(*outputDir, config.DomainDir)
		if err := os.MkdirAll(domainDir, 0755); err != nil {
			fmt.Printf("Error creating domain directory: %v\n", err)
			os.Exit(1)
		}

		typeGen := generator.NewTypeGenerator(apiParser, config.DomainPackage, templateFS)
		typesCode, err := typeGen.GenerateTypes()
		if err != nil {
			fmt.Printf("Error generating types: %v\n", err)
			os.Exit(1)
		}

		typesFilePath := filepath.Join(domainDir, config.TypesFile)
		if _, err := os.Stat(typesFilePath); os.IsNotExist(err) || *overwrite {
			if err := os.WriteFile(typesFilePath, []byte(typesCode), 0644); err != nil {
				fmt.Printf("Error writing types file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Generated types in %s\n", typesFilePath)
		} else {
			fmt.Printf("Types file already exists. Skipping (use --overwrite to force overwrite)\n")
		}
	}

	// Create internal directory structure
	internalDir := filepath.Join(*outputDir, config.InternalDir)
	pkgDir := filepath.Join(*outputDir, config.PkgDir)
	domainDir := filepath.Join(*outputDir, config.DomainDir)
	servicesDir := filepath.Join(*outputDir, config.ServicesDir)
	adaptersDir := filepath.Join(*outputDir, config.AdaptersDir)

	// Create domain error types directory
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		fmt.Printf("Error creating domain directory: %v\n", err)
		os.Exit(1)
	}

	// Generate domain errors from template
	domainErrorsTemplate, err := template.ParseFS(templateFS, config.DomainErrorsTemplate)
	if err != nil {
		fmt.Printf("Error parsing domain errors template: %v\n", err)
	} else {
		var buf bytes.Buffer
		if err := domainErrorsTemplate.Execute(&buf, nil); err != nil {
			fmt.Printf("Error executing domain errors template: %v\n", err)
		} else {
			errorsDest := filepath.Join(domainDir, config.ErrorsFile)
			// Check if file exists, don't overwrite unless explicitly requested
			if _, err := os.Stat(errorsDest); os.IsNotExist(err) || *overwrite {
				if err := os.WriteFile(errorsDest, buf.Bytes(), 0644); err != nil {
					fmt.Printf("Error writing domain errors file: %v\n", err)
				} else {
					fmt.Printf("Generated domain errors in %s\n", errorsDest)
				}
			} else {
				fmt.Printf("Domain errors file already exists. Skipping (use --overwrite to force overwrite)\n")
			}
		}
	}

	// Generate services for each schema (if explicitly requested or if HTTP handlers need them)
	if *genServices || *genHTTP {
		if err := os.MkdirAll(servicesDir, 0755); err != nil {
			fmt.Printf("Error creating services directory: %v\n", err)
			os.Exit(1)
		}

		serviceGen, err := generator.NewServiceGenerator(apiParser, *packageName, importPath, templateFS)
		if err != nil {
			fmt.Printf("Error creating service generator: %v\n", err)
		} else {
			// Generate service for each schema
			for _, name := range schemaNames {
				serviceCode, err := serviceGen.GenerateService(name)
				if err != nil {
					fmt.Printf("Error generating service for %s: %v\n", name, err)
					continue
				}

				// Create domain-specific directory for the service
				schemaServiceDir := filepath.Join(servicesDir, strings.ToLower(name))
				if err := os.MkdirAll(schemaServiceDir, 0755); err != nil {
					fmt.Printf("Error creating directory for %s: %v\n", name, err)
					continue
				}

				// Write service file
				serviceFilename := strings.ToLower(name) + "_service.go"
				serviceFilePath := filepath.Join(schemaServiceDir, serviceFilename)

				// Check if file exists, don't overwrite unless explicitly requested
				if _, err := os.Stat(serviceFilePath); os.IsNotExist(err) || *overwrite {
					if err := os.WriteFile(serviceFilePath, []byte(serviceCode), 0644); err != nil {
						fmt.Printf("Error writing service file for %s: %v\n", name, err)
						continue
					}
					fmt.Printf("Generated service for %s in %s\n", name, serviceFilePath)
				} else {
					fmt.Printf("Service file for %s already exists. Skipping (use --overwrite to force overwrite)\n", name)
				}

				// Generate service tests
				serviceTestCode, err := serviceGen.GenerateServiceTests(name)
				if err != nil {
					fmt.Printf("Error generating service tests for %s: %v\n", name, err)
				} else {
					// Write service test file
					testFilename := strings.ToLower(name) + "_service_test.go"
					testFilePath := filepath.Join(schemaServiceDir, testFilename)

					// Check if file exists, don't overwrite unless explicitly requested
					if _, err := os.Stat(testFilePath); os.IsNotExist(err) || *overwrite {
						if err := os.WriteFile(testFilePath, []byte(serviceTestCode), 0644); err != nil {
							fmt.Printf("Error writing service test file for %s: %v\n", name, err)
						} else {
							fmt.Printf("Generated service tests for %s in %s\n", name, testFilePath)
						}
					} else {
						fmt.Printf("Service test file for %s already exists. Skipping (use --overwrite to force overwrite)\n", name)
					}
				}
			}
		}
	}

	// Generate MongoDB repositories if requested
	if *genMongo {
		// Set up mongo directory
		mongoDir := filepath.Join(*outputDir, config.MongoAdaptersDir)
		if err := os.MkdirAll(mongoDir, 0755); err != nil {
			fmt.Printf("Error creating mongo directory: %v\n", err)
			os.Exit(1)
		}

		repoPackage := config.DefaultRepoPackage // Using just 'mongo' as the package name, following Go conventions
		mongoImportPath := importPath

		mongoGen, err := generator.NewMongoGenerator(apiParser, *packageName, repoPackage, mongoImportPath, templateFS)
		if err != nil {
			fmt.Printf("Error creating MongoDB generator: %v\n", err)
			os.Exit(1)
		}

		// Generate repository and tests for each schema
		for _, name := range schemaNames {
			// Generate repository
			repoCode, err := mongoGen.GenerateRepository(name)
			if err != nil {
				fmt.Printf("Error generating repository for %s: %v\n", name, err)
				continue
			}

			// Create domain-specific directory for the repository
			domainDir := filepath.Join(mongoDir, strings.ToLower(name))
			if err := os.MkdirAll(domainDir, 0755); err != nil {
				fmt.Printf("Error creating directory for %s: %v\n", name, err)
				continue
			}

			// Write repository file
			repoFilename := strings.ToLower(name) + "_repository.go"
			repoFilePath := filepath.Join(domainDir, repoFilename)

			// Check if file exists, don't overwrite unless explicitly requested
			if _, err := os.Stat(repoFilePath); os.IsNotExist(err) || *overwrite {
				if err := os.WriteFile(repoFilePath, []byte(repoCode), 0644); err != nil {
					fmt.Printf("Error writing repository file for %s: %v\n", name, err)
					continue
				}
				fmt.Printf("Generated MongoDB repository for %s in %s\n", name, repoFilePath)
			} else {
				fmt.Printf("Repository file for %s already exists. Skipping (use --overwrite to force overwrite)\n", name)
			}

			// Generate test file
			testCode, err := mongoGen.GenerateRepositoryTests(name)
			if err != nil {
				fmt.Printf("Error generating repository tests for %s: %v\n", name, err)
			} else {
				// Write test file to the same domain directory
				testFilename := strings.ToLower(name) + "_repository_test.go"
				domainDir := filepath.Join(mongoDir, strings.ToLower(name))
				testFilePath := filepath.Join(domainDir, testFilename)

				// Check if file exists, don't overwrite unless explicitly requested
				if _, err := os.Stat(testFilePath); os.IsNotExist(err) || *overwrite {
					if err := os.WriteFile(testFilePath, []byte(testCode), 0644); err != nil {
						fmt.Printf("Error writing repository test file for %s: %v\n", name, err)
					} else {
						fmt.Printf("Generated repository tests for %s in %s\n", name, testFilePath)
					}
				} else {
					fmt.Printf("Repository test file for %s already exists. Skipping (use --overwrite to force overwrite)\n", name)
				}
			}
		}
	}

	// Generate HTTP handlers if requested
	if *genHTTP {
		// Create internal directory structure if it doesn't exist
		httpDir := filepath.Join(*outputDir, config.HttpAdaptersDir)
		httpUtilDir := filepath.Join(*outputDir, config.HttpUtilDir)

		// Create all required directories
		dirs := []string{internalDir, adaptersDir, httpDir, pkgDir, httpUtilDir}
		for _, dir := range dirs {
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Printf("Error creating directory %s: %v\n", dir, err)
				os.Exit(1)
			}
		}

		// Build import paths
		modelImportPath := importPath
		httpImportPath := importPath

		httpGen, err := generator.NewHTTPGenerator(apiParser, *packageName, *httpPackage, httpImportPath, modelImportPath, templateFS)
		if err != nil {
			fmt.Printf("Error creating HTTP handler generator: %v\n", err)
			os.Exit(1)
		}

		// Generate HTTP handlers
		handlersCode, err := httpGen.GenerateHandlers()
		if err != nil {
			fmt.Printf("Error generating HTTP handlers: %v\n", err)
			os.Exit(1)
		}

		// Write handler files
		for filename, code := range handlersCode {
			var handlerFilePath string

			// Handle different file types based on prefix
			if strings.HasPrefix(filename, "httputil/") {
				// Write to internal/pkg/httputil directory
				baseFilename := filepath.Base(filename)
				handlerFilePath = filepath.Join(httpUtilDir, baseFilename)
			} else if strings.HasPrefix(filename, "domain/") {
				// Extract domain name from path like "domain/pet/handler.go"
				parts := strings.Split(filename, "/")
				if len(parts) >= 2 {
					domain := parts[1]
					// Create domain directory if it doesn't exist
					domainDir := filepath.Join(httpDir, domain)
					if err := os.MkdirAll(domainDir, 0755); err != nil {
						fmt.Printf("Error creating domain directory %s: %v\n", domainDir, err)
						continue
					}

					// Get filename without domain prefix
					domainFilename := strings.Join(parts[2:], "/")
					handlerFilePath = filepath.Join(domainDir, domainFilename)
				} else {
					// Fallback if path format is unexpected
					handlerFilePath = filepath.Join(httpDir, filename)
				}
			} else {
				// Write regular handler file to HTTP adapters directory
				handlerFilePath = filepath.Join(httpDir, filename)
			}

			// Ensure parent directory exists
			parentDir := filepath.Dir(handlerFilePath)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				fmt.Printf("Error creating directory %s: %v\n", parentDir, err)
				continue
			}

			// Check if file exists, don't overwrite unless explicitly requested
			if _, err := os.Stat(handlerFilePath); os.IsNotExist(err) || *overwrite {
				// Write the file
				if err := os.WriteFile(handlerFilePath, []byte(code), 0644); err != nil {
					fmt.Printf("Error writing file %s: %v\n", handlerFilePath, err)
					continue
				}
				fmt.Printf("Generated HTTP handler in %s\n", handlerFilePath)
			} else {
				fmt.Printf("HTTP handler file %s already exists. Skipping (use --overwrite to force overwrite)\n", handlerFilePath)
			}
		}
	}

	// Regenerate routes.go if any components were generated
	if *genServices || *genMongo || *genHTTP || *initProject {
		// Create main generator for routes update
		mainGen, err := generator.NewMainGenerator(apiParser, importPath, templateFS)
		if err != nil {
			fmt.Printf("Error creating main generator for routes: %v\n", err)
		} else {
			// Configure main generator
			mainGen.SetMongoURI("mongodb://localhost:27017")
			mainGen.SetDBName(filepath.Base(targetModuleName))
			mainGen.SetDefaultPort("8080")

			// Generate routes.go with current feature flags
			files, err := mainGen.GenerateWithFeatures(*genMongo, *genMongo, *genHTTP)
			if err != nil {
				fmt.Printf("Error generating routes: %v\n", err)
			} else {
				// Only write routes.go (always overwrite)
				if routesContent, exists := files["routes.go"]; exists {
					cmdDir := filepath.Join(*outputDir, "cmd", filepath.Base(targetModuleName))
					routesPath := filepath.Join(cmdDir, "routes.go")

					// Ensure directory exists
					if err := os.MkdirAll(cmdDir, 0755); err != nil {
						fmt.Printf("Error creating directory for routes.go: %v\n", err)
					} else {
						if err := os.WriteFile(routesPath, []byte(routesContent), 0644); err != nil {
							fmt.Printf("Error writing routes.go: %v\n", err)
						} else {
							fmt.Printf("Updated routes.go in %s\n", routesPath)
						}
					}
				}
			}
		}
	}

	// Generate .env file
	if *initProject {
		envTemplate, err := template.ParseFS(templateFS, config.EnvTemplate)
		if err != nil {
			fmt.Printf("Error parsing .env template: %v\n", err)
		} else {
			var buf bytes.Buffer
			dbNameDefault := filepath.Base(targetModuleName)
			if dbNameDefault == "." {
				dbNameDefault = "api"
			}

			if err := envTemplate.Execute(&buf, map[string]string{
				"MongoURI":    "mongodb://localhost:27017",
				"DBName":      dbNameDefault,
				"DefaultPort": "8080",
			}); err != nil {
				fmt.Printf("Error executing .env template: %v\n", err)
			} else {
				envPath := filepath.Join(*outputDir, config.EnvFile)
				if _, err := os.Stat(envPath); os.IsNotExist(err) || *overwrite {
					if err := os.WriteFile(envPath, buf.Bytes(), 0644); err != nil {
						fmt.Printf("Error writing .env file: %v\n", err)
					} else {
						fmt.Printf("Generated .env file in %s\n", envPath)
					}
				} else {
					fmt.Printf(".env file already exists. Skipping (use --overwrite to force overwrite)\n")
				}
			}
		}
	}
}

// getModuleNameFromPath attempts to determine the Go module name from go.mod in specified directory
func getModuleNameFromPath(dir string) (string, error) {
	// Try to open go.mod in the specified directory
	goModPath := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}

	// Find the module line
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1], nil
			}
		}
	}

	return "", fmt.Errorf("module declaration not found in go.mod")
}
