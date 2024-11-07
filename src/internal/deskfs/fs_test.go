package deskfs

import (
	"os"
	"path/filepath"
	"testing"

	"desktop-cleaner/internal/terminal"

	"github.com/stretchr/testify/assert"
)

// Helper to create a temporary directory structure for tests
func setupTestDir(t *testing.T, structure map[string]string) (string, func()) {
	dir, err := os.MkdirTemp("", "desktop_cleaner_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create directories and files as defined in structure map
	for path, content := range structure {
		fullPath := filepath.Join(dir, path)
		if filepath.Ext(path) == "" {
			// Create directory
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				t.Fatalf("failed to create directory %s: %v", fullPath, err)
			}
		} else {
			// Ensure the parent directory exists
			parentDir := filepath.Dir(fullPath)
			if err := os.MkdirAll(parentDir, 0755); err != nil {
				t.Fatalf("failed to create parent directory %s: %v", parentDir, err)
			}
			// Create file with content
			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				t.Fatalf("failed to create file %s: %v", fullPath, err)
			}
		}
	}

	// Cleanup function to remove the directory after the test
	return dir, func() { os.RemoveAll(dir) }
}

// Helper to create a temporary config file with the specified contents
func createTestConfigFile(t *testing.T, content string) (string, func()) {
	tmpFile, err := os.CreateTemp("", "desktop_cleaner_config_*.toml")
	if err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}
	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write to temp config file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp config file: %v", err)
	}

	// Cleanup function to remove the config file after the test
	return tmpFile.Name(), func() { os.Remove(tmpFile.Name()) }
}

func TestNewConfig(t *testing.T) {
	// Test loading from the current working directory
	t.Run("loads from current working directory", func(t *testing.T) {
		dir, cleanup := setupTestDir(t, map[string]string{
			".desktop_cleaner.toml": "file_types = { \"docs\" = [\".docx\"] }",
		})
		defer cleanup()

		// Change to the test directory to simulate CWD config loading
		originalDir, _ := os.Getwd()
		defer os.Chdir(originalDir)
		os.Chdir(dir)

		config, err := NewConfig(nil)
		assert.NoError(t, err)
		assert.Contains(t, config.FileTypes, "docs")
		assert.Equal(t, []string{".docx"}, config.FileTypes["docs"])
	})

	// Test loading from an optional config path
	t.Run("loads from optional config path", func(t *testing.T) {
		configContent := "file_types = { \"pics\" = [\".jpg\", \".png\"] }"
		configPath, cleanup := createTestConfigFile(t, configContent)
		defer cleanup()

		config, err := NewConfig(&configPath)
		assert.NoError(t, err)
		assert.Contains(t, config.FileTypes, "pics")
		assert.Equal(t, []string{".jpg", ".png"}, config.FileTypes["pics"])
	})

	// Test default config creation if no config found
	t.Run("creates default config if no config found", func(t *testing.T) {
		config, err := NewConfig(nil)
		assert.NoError(t, err)
		assert.NotNil(t, config.FileTypes)
		assert.Contains(t, config.FileTypes, "notes")
		assert.Equal(t, []string{".md", ".rtf", ".txt"}, config.FileTypes["notes"])
	})
}

func TestBuildTreeAndCache(t *testing.T) {
	term := terminal.NewTerminal() // Initialize terminal instance
	dfs := NewDesktopFS(term)      // Initialize DesktopFS with terminal

	// Setup directories and files for testing tree building
	dir, cleanup := setupTestDir(t, map[string]string{
		"docs/report.docx": "",
		"pics/photo.jpg":   "",
		"scripts/setup.sh": "",
	})
	defer cleanup()

	// Set up DirectoryTree root and initialize the cache
	dfs.DirectoryTree = &DirectoryTree{Root: NewTreeNode(dir, Directory, nil)}
	dfs.DirectoryTree.Cache = make(map[string]*TreeNode)

	// Build the directory tree
	err := dfs.buildTreeAndCache(dir, true)
	assert.NoError(t, err)

	// Validate structure of the tree
	assert.Equal(t, 3, len(dfs.DirectoryTree.Root.Children))
	assert.Contains(t, dfs.DirectoryTree.Cache, filepath.Join(dir, "docs/report.docx"))
	assert.Contains(t, dfs.DirectoryTree.Cache, filepath.Join(dir, "pics/photo.jpg"))
	assert.Contains(t, dfs.DirectoryTree.Cache, filepath.Join(dir, "scripts/setup.sh"))
}

func TestEnhancedOrganize(t *testing.T) {
	term := terminal.NewTerminal()
	dfs := NewDesktopFS(term)

	// Setup directories and files for testing organization
	dir, cleanup := setupTestDir(t, map[string]string{
		"source/report.docx": "",
		"source/photo.jpg":   "",
		"source/setup.sh":    "",
		"target/.desktop_cleaner.toml": `
			file_types = { "docs" = [".docx"], "pics" = [".jpg"], "scripts" = [".sh"] }
			nested_dirs = { "docs" = ["Reports"], "pics" = ["Photos"], "scripts" = ["Setup"] }
		`,
	})
	defer cleanup()

	// Initialize config to point to the test target directory config

	configFile := filepath.Join(dir, "target/.desktop_cleaner.toml")

	dfs.InitConfig(&configFile)

	params := &FilePathParams{
		SourceDir:   filepath.Join(dir, "source"),
		TargetDir:   filepath.Join(dir, "target"),
		Recursive:   true,
		CopyFiles:   false,
		RemoveAfter: false,
	}

	// Run EnhancedOrganize
	err := dfs.EnhancedOrganize(dfs.InstanceConfig, params)
	assert.NoError(t, err)

	// Validate organized file paths
	assert.FileExists(t, filepath.Join(dir, "target/docs/Reports/report.docx"))
	assert.FileExists(t, filepath.Join(dir, "target/pics/Photos/photo.jpg"))
	assert.FileExists(t, filepath.Join(dir, "target/scripts/Setup/setup.sh"))
}
