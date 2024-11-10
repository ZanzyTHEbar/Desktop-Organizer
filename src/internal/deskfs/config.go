package deskfs

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/ZanzyTHEbar/assert-lib"
	gobaselogger "github.com/ZanzyTHEbar/go-basetools/logger"
)

var (
	// DefaultConfigPath is the default path to the config file
	DefaultConfigPath   = filepath.Join(os.Getenv("HOME"), ".config", "desktop_cleaner")
	DefaultConfigName   = ".desktop_cleaner"
	ConfigAssertHandler = assert.NewAssertHandler()
)

// Config holds the mapping of file types to extensions
type DeskFSConfig struct {
	gobaselogger.Config
	DirectoryTree *DirectoryTree `toml:"directory_tree"`
	FileTypeTree  *FileTypeTree  `toml:"file_type_tree"`
	TargetDir     string         `toml:"target_dir"`
	CacheDir      string         `toml:"cache_dir"`
}

type IntermediateConfig struct {
	gobaselogger.Config
	FileTypes map[string][]string `toml:"file_types"` // Ensure TOML tag matches the file
	CacheDir  string              `toml:"cache_dir"`
}

func CreateDirIfNotExist(path string) {
	// Create the directory if it doesn't exist
	if _, err := os.Stat(filepath.Dir(path)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			slog.Info(fmt.Sprintf("Path %s: %v", filepath.Dir(path), err))
			errMsg := fmt.Sprintf("Error creating directory at %s", filepath.Dir(path))
			ConfigAssertHandler.NoError(context.Background(), err, errMsg, slog.Error)
		}
	}
}

func NewIntermediateConfig(optionalPath string) *IntermediateConfig {
	var configPath string

	// Step 1: Determine the configuration file path
	tomlFileName := DefaultConfigName + ".toml"
	if _, err := os.Stat(tomlFileName); err != nil && optionalPath == "" {
		slog.Debug(fmt.Sprintf("Error loading config file: %v\n", err))
		configPath = tomlFileName
	} else if optionalPath != "" {
		configPath = optionalPath
	} else {
		configPath = filepath.Join(DefaultConfigPath, tomlFileName)
	}

	slog.Info(fmt.Sprintf("\nConfig path: %s\n", configPath))

	var defaultConfig IntermediateConfig

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig = getDefaultConfig()
		slog.Info(fmt.Sprintf("\nPath %s: %v", filepath.Dir(configPath), err))
		CreateDirIfNotExist(filepath.Dir(configPath))
		file, err := os.Create(configPath)
		if err != nil {
			slog.Error(fmt.Sprintf("Error creating default config file: %v", err))
			return nil
		}
		defer file.Close()

		encoder := toml.NewEncoder(file)
		if err := encoder.Encode(defaultConfig); err != nil {
			slog.Error(fmt.Sprintf("Error writing default config file: %v", err))
			return nil
		}
		slog.Info(fmt.Sprintf("Default config file created at %s", configPath))
	} else {
		// Step 3: Decode the existing config file
		fmt.Printf("Loading config file from: %s\n", configPath)
		var tempConfig map[string]interface{}
		if _, err := toml.DecodeFile(configPath, &tempConfig); err != nil {
			slog.Error(fmt.Sprintf("Error decoding config file: %v", err))
			return nil
		}

		fmt.Printf("TempConfig (raw): %+v\n", tempConfig)

		// Decode configuration file into IntermediateConfig
		if _, err := toml.DecodeFile(configPath, &defaultConfig); err != nil {
			slog.Error(fmt.Sprintf("Error decoding config file to struct: %v", err))
			return nil
		}

		// Debugging: Print the loaded configuration values to ensure correctness
		//fmt.Printf("Loaded file_types (after decode): %+v\n", defaultConfig.FileTypes)
	}

	// Step 4: Confirm loaded config (case-sensitive)
	fmt.Printf("Loaded file_types (case-sensitive): %+v\n", defaultConfig.FileTypes)

	return &defaultConfig
}

func NewDeskFSConfig() *DeskFSConfig {
	assertHandler := assert.NewAssertHandler()

	cwd, err := os.Getwd()
	assertHandler.NoError(context.Background(), err, fmt.Sprintf("Error getting current working directory: %v", err), slog.Error)

	directoryTree, err := NewDirectoryTree(cwd)
	assertHandler.NoError(context.Background(), err, fmt.Sprintf("Error creating directory tree: %v", err), slog.Error)

	return &DeskFSConfig{
		DirectoryTree: directoryTree,
		FileTypeTree:  NewFileTypeTree(),
	}
}

func (dfc *DeskFSConfig) BuildFileTypeTree(config *IntermediateConfig) *DeskFSConfig {
	// Populate FileTypeTree using the intermediate config data
	dfc.FileTypeTree.PopulateFileTypes(config.FileTypes)
	return dfc
}

func (dfc *IntermediateConfig) SaveConfig(config *IntermediateConfig, filePath string) error {
	dfc.Config.Cfg.Set("file_types", config.FileTypes)
	dfc.Config.Cfg.Set("logger.style", config.Logger.Style)
	dfc.Config.Cfg.Set("logger.level", config.Logger.Level)
	dfc.Config.Cfg.Set("cache_dir", config.CacheDir)

	if err := dfc.Config.Cfg.WriteConfig(); err != nil {
		return err
	}

	return nil
}

// Returns the default configuration
func getDefaultConfig() IntermediateConfig {
	return IntermediateConfig{
		FileTypes: map[string][]string{
			"Notes":      {".md", ".rtf", ".txt"},
			"Docs":       {".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx"},
			"EXE":        {".exe", ".appimage", ".msi"},
			"Vids":       {".mp4", ".mov", ".avi", ".mkv"},
			"Compressed": {".zip", ".rar", ".tar", ".gz", ".7z"},
			"Scripts":    {".sh", ".bat"},
			"Installers": {".deb", ".rpm"},
			"Books":      {".epub", ".mobi"},
			"Music":      {".mp3", ".wav", ".ogg", ".flac"},
			"PDFS":       {".pdf"},
			"Pics":       {".bmp", ".gif", ".jpg", ".jpeg", ".svg", ".png"},
			"Torrents":   {".torrent"},
			"CODE": {
				".c", ".h", ".py", ".rs", ".go", ".js", ".ts", ".jsx", ".tsx", ".html",
				".css", ".php", ".java", ".cpp", ".cs", ".vb", ".sql", ".pl", ".swift",
				".kt", ".r", ".m", ".asm",
			},
			"Markup": {
				".json", ".xml", ".yml", ".yaml", ".ini", ".toml", ".cfg", ".conf", ".log",
			},
		},
		Config: gobaselogger.Config{
			Logger: gobaselogger.Logger{
				Style: "json",
				Level: gobaselogger.LoggerLevels["info"].String(),
			},
		},
		CacheDir: ".cache",
	}
}
