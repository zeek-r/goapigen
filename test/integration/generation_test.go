package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEndGeneration tests the complete generation pipeline
func TestEndToEndGeneration(t *testing.T) {
	tests := []struct {
		name     string
		flags    []string
		validate func(t *testing.T, outputDir string)
	}{
		{
			name:     "init_only",
			flags:    []string{"--init"},
			validate: validateInitGeneration,
		},
		{
			name:     "full_generation",
			flags:    []string{"--init", "--services", "--http"},
			validate: validateFullGeneration,
		},
		{
			name:     "types_only",
			flags:    []string{"--types"},
			validate: validateTypesGeneration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir := t.TempDir()

			// Build goapigen binary
			binaryPath := buildGoapigenBinary(t)

			// Prepare command
			args := []string{
				"--spec", "../../examples/petstore/openapi.yaml",
				"--output", tempDir,
			}
			args = append(args, tt.flags...)

			// Run generation
			cmd := exec.Command(binaryPath, args...)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "Generation failed with output: %s", string(output))

			// Validate generated code compiles
			if contains(tt.flags, "--init") {
				validateCompilation(t, tempDir)
			}

			// Run specific validation
			tt.validate(t, tempDir)
		})
	}
}

func buildGoapigenBinary(t *testing.T) string {
	binaryPath := filepath.Join(t.TempDir(), "goapigen")
	cmd := exec.Command("go", "build", "-o", binaryPath, "../../cmd/goapigen")
	require.NoError(t, cmd.Run(), "Failed to build goapigen binary")
	return binaryPath
}

func validateCompilation(t *testing.T, outputDir string) {
	// Find the generated project directory
	projectDirs, err := filepath.Glob(filepath.Join(outputDir, "cmd", "*"))
	require.NoError(t, err)
	require.Greater(t, len(projectDirs), 0, "No generated project found")

	projectDir := projectDirs[0]

	// Try to compile the generated code
	cmd := exec.Command("go", "build", ".")
	cmd.Dir = projectDir
	output, err := cmd.CombinedOutput()
	assert.NoError(t, err, "Generated code should compile successfully. Output: %s", string(output))
}

func validateInitGeneration(t *testing.T, outputDir string) {
	// Check essential files exist
	requiredFiles := []string{
		"go.mod",
		"cmd/*/main.go",
		"cmd/*/routes.go",
		"cmd/*/database.go",
		".env",
	}

	for _, pattern := range requiredFiles {
		matches, err := filepath.Glob(filepath.Join(outputDir, pattern))
		require.NoError(t, err)
		assert.Greater(t, len(matches), 0, "Required file pattern not found: %s", pattern)
	}
}

func validateFullGeneration(t *testing.T, outputDir string) {
	// Validate init files
	validateInitGeneration(t, outputDir)

	// Check additional files for full generation
	requiredPatterns := []string{
		"internal/pkg/domain/types.go",
		"internal/services/*/",
		"internal/adapters/http/*/",
	}

	for _, pattern := range requiredPatterns {
		matches, err := filepath.Glob(filepath.Join(outputDir, pattern))
		require.NoError(t, err)
		assert.Greater(t, len(matches), 0, "Required pattern not found: %s", pattern)
	}
}

func validateTypesGeneration(t *testing.T, outputDir string) {
	// Check domain types are generated
	typesFile := filepath.Join(outputDir, "internal/pkg/domain/types.go")
	assert.FileExists(t, typesFile, "Types file should be generated")

	// Read and validate content
	content, err := os.ReadFile(typesFile)
	require.NoError(t, err)

	// Check for expected types
	contentStr := string(content)
	assert.Contains(t, contentStr, "type Pet struct", "Pet type should be defined")
	assert.Contains(t, contentStr, "type Order struct", "Order type should be defined")
}

// TestRouteRegistration validates that generated APIs respond correctly
func TestRouteRegistration(t *testing.T) {
	t.Skip("Requires running server - implement after fixing 500 errors")

	// This test would:
	// 1. Generate a full project
	// 2. Start the server
	// 3. Make HTTP requests to validate routes work
	// 4. Check responses are not 404
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
