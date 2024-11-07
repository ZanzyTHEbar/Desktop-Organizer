package deskfs

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type DebugLevelType string

const (
	DebugLevelInfo  DebugLevelType = "info"
	DebugLevelDebug DebugLevelType = "debug"
	DebugLevelWarn  DebugLevelType = "warn"
	DebugLevelError DebugLevelType = "error"
	DebugLevelTrace DebugLevelType = "trace"
	DebugLevelOff   DebugLevelType = "off"
)

type Logger struct {
	Style string         `mapstructure:"style"`
	Level DebugLevelType `mapstructure:"level"`
}

// Config holds the mapping of file types to extensions
type Config struct {
	FileTypes  map[string][]string `mapstructure:"file_types"`
	NestedDirs map[string][]string `mapstructure:"nested_dirs"`
	TargetDir  string              `mapstructure:"target_dir"`
	CacheDir   string              `mapstructure:"cache_dir"`
	Logger     Logger              `mapstructure:"logger"`
	cfg        *viper.Viper
}

func NewConfig(optionalPath *string) (*Config, error) {
	cfg := viper.New()

	cfg.SetConfigType("toml") // REQUIRED if the config file does not have the extension in the name

	// Step 1: First, try to load config from CWD
	cwdConfigPath := ".desktop_cleaner.toml"
	if _, err := os.Stat(cwdConfigPath); err == nil {
		cfg.SetConfigFile(cwdConfigPath)
	} else if optionalPath != nil {
		// Step 2: If CWD config is not found, check the provided path
		cfg.SetConfigFile(*optionalPath)
	} else {
		// Step 3: If no path is provided, look for the config in the default locations
		cfg.SetConfigName(".desktop_cleaner") // name of config file (without extension)
		//cfg.AddConfigPath("/etc/desktop_cleaner")          // look for config in the home directory
		cfg.AddConfigPath("$HOME/.config/desktop_cleaner") // look for config in the home directory
		// Look for config in current directory first
		cfg.AddConfigPath(".")
	}

	var defaultConfig Config

	if err := cfg.ReadInConfig(); err != nil {
		// Create a default config file if it doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {

			// set the default config file location

			defaultConfigPath := filepath.Join(os.Getenv("HOME"), ".config", "desktop_cleaner", ".desktop_cleaner.toml")

			cfg.SetConfigFile(defaultConfigPath)

			slog.Warn(fmt.Sprintf("Config file not found. Creating a default config file at %+v\n", cfg.ConfigFileUsed()))

			defaultConfig = getDefaultConfig()

			// log the default config
			slog.Debug(fmt.Sprintf("Default Config: %v", defaultConfig))

			cfg.Set("file_types", defaultConfig.FileTypes)
			cfg.Set("nested_dirs", defaultConfig.NestedDirs)
			cfg.Set("cache_dir", defaultConfig.CacheDir)
			cfg.Set("logger", defaultConfig.Logger)

			// print the contents of viper
			slog.Warn(fmt.Sprintf("Viper: %v", cfg.AllSettings()))

			// Set the cache directory to the same directory as the config file
			defaultConfig.CacheDir = filepath.Join(filepath.Dir(cfg.ConfigFileUsed()), defaultConfig.CacheDir)

			// Create the directory if it doesn't exist
			if _, err := os.Stat(filepath.Dir(defaultConfig.CacheDir)); os.IsNotExist(err) {
				if err := os.MkdirAll(filepath.Dir(defaultConfig.CacheDir), 0755); err != nil {
					slog.Error(fmt.Sprintf("Error creating cache directory at %s", defaultConfig.CacheDir))
					return nil, err
				}
			}

			// marshal the default config to viper

			if err := cfg.WriteConfigAs(defaultConfigPath); err != nil {
				slog.Error(fmt.Sprintf("Error creating default config file at %s", defaultConfigPath))
				return nil, err
			} else {
				slog.Info(fmt.Sprintf("Default config file created at %s", defaultConfigPath))
			}
		} else {
			return nil, err
		}
	}

	if err := cfg.Unmarshal(&defaultConfig); err != nil {
		return nil, err
	}

	return &defaultConfig, nil
}

func (c *Config) SaveConfig(config *Config, filePath string) error {
	c.cfg.Set("file_types", config.FileTypes)
	c.cfg.Set("target_dir", config.TargetDir)
	c.cfg.Set("nested_dirs", config.NestedDirs)
	c.cfg.Set("logger.style", config.Logger.Style)
	c.cfg.Set("logger.level", config.Logger.Level)
	c.cfg.Set("cache_dir", config.CacheDir)

	if err := c.cfg.WriteConfig(); err != nil {
		return err
	}

	return nil
}

// Returns the default configuration
func getDefaultConfig() Config {
	return Config{
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
		Logger: Logger{
			Style: "json",
			Level: DebugLevelInfo,
		},
		CacheDir: ".cache",
	}
}
