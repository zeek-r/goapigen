package config

// Directory structure constants for generated projects
const (
	// Default directories
	InternalDir      = "internal"
	PkgDir           = "internal/pkg"
	DomainDir        = "internal/pkg/domain"
	ServicesDir      = "internal/services"
	AdaptersDir      = "internal/adapters"
	HttpAdaptersDir  = "internal/adapters/http"
	MongoAdaptersDir = "internal/adapters/repository"
	HttpUtilDir      = "internal/pkg/httputil"
	LoggerDir        = "internal/pkg/logger"
	ConfigDir        = "internal/pkg/config"

	// Default package names
	DefaultAPIPackage     = "api"
	DefaultHandlerPackage = "http"
	DefaultRepoPackage    = "repository"
	ServicePackage        = "service"
	DomainPackage         = "domain"
	HttpUtilPackage       = "httputil"
	LoggerPackage         = "logger"
	ConfigPackage         = "config"

	// File names
	GoModFile          = "go.mod"
	TypesFile          = "types.go"
	ErrorsFile         = "errors.go"
	RouterFile         = "router.go"
	HttpUtilsFile      = "http_utils.go"
	HandlerWrapperFile = "handler_wrapper.go"
	MainFile           = "main.go"
	EnvFile            = ".env"
	LoggerFile         = "logger.go"
	LoggerTestFile     = "logger_test.go"
	ConfigFile         = "config.go"
	ConfigTestFile     = "config_test.go"

	// Template paths
	DomainErrorsTemplate = "templates/domain/errors.go.tmpl"
	DomainTypesTemplate  = "templates/domain/types.go.tmpl"
	MainTemplate         = "templates/main.go.tmpl"
	EnvTemplate          = "templates/env.tmpl"
	LoggerTemplate       = "templates/pkg/logger.go.tmpl"
	LoggerTestTemplate   = "templates/pkg/logger_test.go.tmpl"
	ConfigTemplate       = "templates/pkg/config.go.tmpl"
	ConfigTestTemplate   = "templates/pkg/config_test.go.tmpl"
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

func GetLoggerImportPath(baseImportPath string) string {
	return baseImportPath + "/" + LoggerDir
}

func GetConfigImportPath(baseImportPath string) string {
	return baseImportPath + "/" + ConfigDir
}
