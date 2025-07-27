package cli

import (
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock embed.FS for testing
var mockTemplateFS embed.FS

func TestGenerationConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *GenerationConfig
		validate func(t *testing.T, config *GenerationConfig)
	}{
		{
			name: "default_config",
			config: &GenerationConfig{
				SpecFile:    "../../examples/petstore/openapi.yaml",
				OutputDir:   t.TempDir(),
				PackageName: "api",
				GenTypes:    true,
				InitProject: true,
			},
			validate: func(t *testing.T, config *GenerationConfig) {
				assert.Equal(t, "api", config.PackageName)
				assert.True(t, config.GenTypes)
				assert.True(t, config.InitProject)
			},
		},
		{
			name: "full_generation_config",
			config: &GenerationConfig{
				SpecFile:    "../../examples/petstore/openapi.yaml",
				OutputDir:   t.TempDir(),
				GenTypes:    true,
				GenServices: true,
				GenMongo:    true,
				GenHTTP:     true,
				InitProject: true,
				Overwrite:   true,
			},
			validate: func(t *testing.T, config *GenerationConfig) {
				assert.True(t, config.GenTypes)
				assert.True(t, config.GenServices)
				assert.True(t, config.GenMongo)
				assert.True(t, config.GenHTTP)
				assert.True(t, config.InitProject)
				assert.True(t, config.Overwrite)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.config)
		})
	}
}

func TestNewGenerationPipeline(t *testing.T) {
	tempDir := t.TempDir()

	config := &GenerationConfig{
		SpecFile:  "../../examples/petstore/openapi.yaml",
		OutputDir: tempDir,
	}

	pipeline, err := NewGenerationPipeline(config, mockTemplateFS)
	require.NoError(t, err, "Should create pipeline successfully")

	assert.NotNil(t, pipeline.parser, "Parser should be initialized")
	assert.Equal(t, config, pipeline.config, "Config should be stored")
	assert.NotEmpty(t, pipeline.targetModule, "Target module should be detected")
	assert.NotEmpty(t, pipeline.importPath, "Import path should be set")

	// Check that output directory was created
	assert.DirExists(t, tempDir, "Output directory should exist")

	// Check that go.mod was created
	goModPath := filepath.Join(tempDir, "go.mod")
	assert.FileExists(t, goModPath, "go.mod should be created")
}

func TestNewGenerationPipeline_InvalidSpec(t *testing.T) {
	config := &GenerationConfig{
		SpecFile:  "nonexistent.yaml",
		OutputDir: t.TempDir(),
	}

	_, err := NewGenerationPipeline(config, mockTemplateFS)
	assert.Error(t, err, "Should fail with invalid spec file")
	assert.Contains(t, err.Error(), "error parsing OpenAPI spec")
}

func TestGetSchemaNames(t *testing.T) {
	tempDir := t.TempDir()

	// Create pipeline
	config := &GenerationConfig{
		SpecFile:  "../../examples/petstore/openapi.yaml",
		OutputDir: tempDir,
	}
	pipeline, err := NewGenerationPipeline(config, mockTemplateFS)
	require.NoError(t, err)

	t.Run("all_schemas", func(t *testing.T) {
		pipeline.config.SchemaName = ""
		schemas, err := pipeline.getSchemaNames()
		require.NoError(t, err)

		// Should return all schemas from petstore spec
		assert.Contains(t, schemas, "Pet", "Should include Pet schema")
		assert.Contains(t, schemas, "Order", "Should include Order schema")
		assert.Greater(t, len(schemas), 0, "Should have at least one schema")
	})

	t.Run("specific_schema", func(t *testing.T) {
		pipeline.config.SchemaName = "Pet"
		schemas, err := pipeline.getSchemaNames()
		require.NoError(t, err)

		assert.Equal(t, []string{"Pet"}, schemas, "Should return only Pet schema")
	})

	t.Run("nonexistent_schema", func(t *testing.T) {
		pipeline.config.SchemaName = "NonExistent"
		_, err := pipeline.getSchemaNames()
		assert.Error(t, err, "Should fail with nonexistent schema")
		assert.Contains(t, err.Error(), "not found in spec")
	})
}

func TestInitializeGoModule(t *testing.T) {
	t.Run("create_new_module", func(t *testing.T) {
		tempDir := t.TempDir()

		moduleName, err := initializeGoModule(tempDir)
		require.NoError(t, err)

		assert.NotEmpty(t, moduleName, "Module name should not be empty")

		// Check go.mod exists
		goModPath := filepath.Join(tempDir, "go.mod")
		assert.FileExists(t, goModPath, "go.mod should be created")

		// Check content
		content, err := os.ReadFile(goModPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "module", "go.mod should contain module declaration")
	})

	t.Run("existing_module", func(t *testing.T) {
		tempDir := t.TempDir()

		// Create existing go.mod
		goModContent := "module test-existing\n\ngo 1.24\n"
		goModPath := filepath.Join(tempDir, "go.mod")
		err := os.WriteFile(goModPath, []byte(goModContent), 0644)
		require.NoError(t, err)

		moduleName, err := initializeGoModule(tempDir)
		require.NoError(t, err)

		assert.Equal(t, "test-existing", moduleName, "Should use existing module name")
	})
}

// TestGenerationPipelineIntegration tests the pipeline with stubs
func TestGenerationPipelineIntegration(t *testing.T) {
	tempDir := t.TempDir()

	config := &GenerationConfig{
		SpecFile:    "../../examples/petstore/openapi.yaml",
		OutputDir:   tempDir,
		InitProject: true,
		GenTypes:    true,
	}

	pipeline, err := NewGenerationPipeline(config, mockTemplateFS)
	require.NoError(t, err)

	// Execute pipeline (with stub implementations)
	err = pipeline.Execute()
	assert.NoError(t, err, "Pipeline execution should succeed with stubs")

	// Basic validation - directory structure should exist
	assert.DirExists(t, filepath.Join(tempDir, "internal"), "Internal directory should be created")
}
