package config

// Directory structure constants for generated projects
const (
	// Root directories
	InternalDir = "internal"

	// Package directories
	PkgDir      = "internal/pkg"
	DomainDir   = "internal/pkg/domain"
	HttpUtilDir = "internal/pkg/httputil"

	// Service directories
	ServicesDir = "internal/services"

	// Adapter directories
	AdaptersDir      = "internal/adapters"
	HttpAdaptersDir  = "internal/adapters/http"
	MongoAdaptersDir = "internal/adapters/mongo"
)

// Package names constants
const (
	// Default package names
	DefaultAPIPackage     = "api"
	DefaultHandlerPackage = "handler"
	DefaultRepoPackage    = "repository"
	ServicePackage        = "service"
	DomainPackage         = "domain"
	HttpUtilPackage       = "httputil"
)

// File names constants
const (
	GoModFile          = "go.mod"
	TypesFile          = "types.go"
	ErrorsFile         = "errors.go"
	RouterFile         = "router.go"
	HttpUtilsFile      = "http_utils.go"
	HandlerWrapperFile = "handler_wrapper.go"
	MainFile           = "main.go"
	EnvFile            = ".env"
)

// Template paths constants
const (
	DomainErrorsTemplate = "templates/domain/errors.go.tmpl"
	DomainTypesTemplate  = "templates/domain/types.go.tmpl"
	MainTemplate         = "templates/main.go.tmpl"
	EnvTemplate          = "templates/env.tmpl"
)

// Import path helpers
func GetDomainImportPath(baseImportPath string) string {
	return baseImportPath + "/" + DomainDir
}

func GetServicesImportPath(baseImportPath, domain string) string {
	return baseImportPath + "/" + ServicesDir + "/" + domain
}

func GetHttpUtilImportPath(baseImportPath string) string {
	return baseImportPath + "/" + HttpUtilDir
}
