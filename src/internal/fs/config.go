package fs

import (
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
	Style string
	Level DebugLevelType
}

// Config holds the mapping of file types to extensions
type Config struct {
	FileTypes  map[string][]string `mapstructure:"file_types"`
	TargetDir  string              `mapstructure:"target_dir"`
	NestedDirs map[string][]string `mapstructure:"nested_dirs"`
	CacheDir   string              `mapstructure:"cache_dir"`
	Logger     Logger
	cfg        *viper.Viper
}

func NewConfig(path *string) (*Config, error) {
	cfg := viper.New()

	cfg.SetConfigName(".config") // name of config file (without extension)

	if path != nil {
		cfg.SetConfigFile(*path)
	} else {
		cfg.AddConfigPath("/etc/desktop_cleaner")          // look for config in the home directory
		cfg.AddConfigPath("$HOME/.config/desktop_cleaner") // look for config in the home directory
		cfg.AddConfigPath(".")
	}

	logger := Logger{
		Style: "json",
		Level: DebugLevelInfo,
	}

	config := Config{
		FileTypes:  make(map[string][]string),
		TargetDir:  "",
		NestedDirs: make(map[string][]string),
		Logger:     logger,
		cfg:        cfg,
	}

	if err := cfg.ReadInConfig(); err != nil {
		// Create a default config file if it doesn't exist
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			config = *getDefaultConfig()

			// Set the cache directory to the same directory as the config file
			config.CacheDir = filepath.Join(filepath.Dir(cfg.ConfigFileUsed()), config.CacheDir)

			if err := cfg.WriteConfig(); err != nil {
				return nil, err
			}
			return nil, err
		}

		return nil, err
	}

	if err := cfg.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
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
func getDefaultConfig() *Config {
	return &Config{
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
