package fs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	ignore "github.com/sabhiram/go-gitignore"
	"gopkg.in/ini.v1"
)

var Cwd string
var CacheDir string

var ConfigFile string
var HomeDir string
var HomeDCDir string

type DebugLevelType string

const (
	DebugLevelInfo  DebugLevelType = "info"
	DebugLevelDebug DebugLevelType = "debug"
	DebugLevelWarn  DebugLevelType = "warn"
	DebugLevelError DebugLevelType = "error"
	DebugLevelTrace DebugLevelType = "trace"
	DebugLevelOff   DebugLevelType = "off"
)

// Config holds the mapping of file types to extensions
type Config struct {
	FileTypes  map[string][]string
	DebugLevel DebugLevelType
}

func init() {
	var err error
	Cwd, err = os.Getwd()
	if err != nil {
		// term.OutputErrorAndExit("Error getting current working directory: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		// term.OutputErrorAndExit("Couldn't find home directory: %v", err)
	}

	HomeDir = home

	if os.Getenv("DESKTOP_CLEANER_ENV") == "development" {
		HomeDCDir = filepath.Join(home, ".desktop-cleaner-dev")
	} else {
		HomeDCDir = filepath.Join(home, ".desktop-cleaner")
	}

	err = os.MkdirAll(HomeDCDir, os.ModePerm)
	if err != nil {
		// term.OutputErrorAndExit("Error creating config directory: %v", err.Error())
	}

	CacheDir = filepath.Join(HomeDCDir, "cache")
	err = os.MkdirAll(filepath.Join(CacheDir, "tiktoken"), os.ModePerm)
	if err != nil {
		// term.OutputErrorAndExit("Error creating cache directory: %v", err.Error())
	}

	err = os.Setenv("TIKTOKEN_CACHE_DIR", CacheDir)
	if err != nil {
		// term.OutputErrorAndExit("Error setting cache directory: %v", err.Error())
	}

	HomeDCDir = findDesktopCleaner(Cwd)
	if HomeDCDir != "" {
		ProjectRoot = Cwd
	}

	ConfigFile = filepath.Join(HomeDCDir, "config.ini")

	// Load the configuration file
	_, err = LoadConfig(ConfigFile)
	if err != nil {
		// term.OutputErrorAndExit("Error loading configuration: %v", err)
	}
}

// LoadConfig loads the configuration from an ini file
func LoadConfig(filePath string) (*Config, error) {
	if _, err := os.Stat(filePath); errors.Is(err, fs.ErrNotExist) {
		// Couldn't load the config file, so create it
		cfg := ini.Empty()

		// Write the default configuration to the file
		err = cfg.ReflectFrom(getDefaultConfig())
		if err != nil {
			// term.OutputErrorAndExit("Error creating config file: %v", err)
		}

		err = cfg.SaveTo(ConfigFile)
		if err != nil {
			// term.OutputErrorAndExit("Error saving config file: %v", err)
		}
	}

	cfg, err := ini.Load(filePath)
	if err != nil {
		return nil, term.OutputErrorAndExit("failed to read config file: %v", err)
	}

	fileTypesSection := cfg.Section("file_types")
	fileTypes := make(map[string][]string)
	for _, key := range fileTypesSection.Keys() {
		// Remove the brackets and split the string by commas
		trimmed := strings.Trim(key.String(), "[]")
		extensions := strings.Split(trimmed, ",")
		// Trim spaces from each extension
		for i, ext := range extensions {
			extensions[i] = strings.TrimSpace(ext)
		}
		fileTypes[key.Name()] = extensions
	}

	debugLevelStr := cfg.Section("debug").Key("level").MustString(string(DebugLevelOff))
	debugLevel := DebugLevelType(debugLevelStr)

	// Validate debug level
	switch debugLevel {
	case DebugLevelInfo, DebugLevelDebug, DebugLevelWarn, DebugLevelError, DebugLevelTrace, DebugLevelOff:
		// valid debug level
	default:
		return nil, fmt.Errorf("invalid debug level: %s", debugLevel)
	}

	return &Config{FileTypes: fileTypes, DebugLevel: debugLevel}, nil
}

// getDefaultConfig returns the default configuration
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
		DebugLevel: "info",
	}
}

func FindOrCreateDesktopCleaner() (string, bool, error) {
	HomeDCDir = findDesktopCleaner(Cwd)
	if HomeDCDir != "" {
		ProjectRoot = Cwd
		return HomeDCDir, false, nil
	}

	// Determine the directory path
	var dir string
	if os.Getenv("DESKTOP_CLEANER_ENV") == "development" {
		dir = filepath.Join(Cwd, ".desktop-cleaner-dev")
	} else {
		dir = filepath.Join(Cwd, ".desktop-cleaner")
	}

	err := os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return "", false, err
	}
	HomeDCDir = dir
	ProjectRoot = Cwd

	return dir, true, nil
}

func ProjectRootIsGitRepo() bool {
	if ProjectRoot == "" {
		return false
	}

	return IsGitRepo(ProjectRoot)
}

func IsGitRepo(dir string) bool {
	isGitRepo := false

	if isCommandAvailable("git") {
		// check whether we're in a git repo
		cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")

		cmd.Dir = dir

		err := cmd.Run()

		if err == nil {
			isGitRepo = true
		}
	}

	return isGitRepo
}

type CleanerPaths struct {
	ActivePaths           map[string]bool
	AllPaths              map[string]bool
	DesktopCleanerIgnored *ignore.GitIgnore
	IgnoredPaths          map[string]string
}

func GetCleanerPaths(baseDir string) (*CleanerPaths, error) {
	if ProjectRoot == "" {
		return nil, fmt.Errorf("no project root found")
	}

	return GetPaths(baseDir, ProjectRoot)
}

func GetPaths(baseDir, currentDir string) (*CleanerPaths, error) {
	ignored, err := GetDesktopCleanerIgnore(currentDir)

	if err != nil {
		return nil, err
	}

	allPaths := map[string]bool{}
	activePaths := map[string]bool{}

	allDirs := map[string]bool{}
	activeDirs := map[string]bool{}

	isGitRepo := IsGitRepo(baseDir)

	errCh := make(chan error)
	var mu sync.Mutex
	numRoutines := 0

	deletedFiles := map[string]bool{}

	if isGitRepo {

		// Use git status to find deleted files
		numRoutines++
		go func() {
			cmd := exec.Command("git", "rev-parse", "--show-toplevel")
			output, err := cmd.Output()
			if err != nil {
				errCh <- fmt.Errorf("error getting git root: %s", err)
				return
			}
			repoRoot := strings.TrimSpace(string(output))

			cmd = exec.Command("git", "status", "--porcelain")
			cmd.Dir = baseDir
			out, err := cmd.Output()
			if err != nil {
				errCh <- fmt.Errorf("error getting git status: %s", err)
			}

			lines := strings.Split(string(out), "\n")

			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "D ") {
					path := strings.TrimSpace(line[2:])
					absPath := filepath.Join(repoRoot, path)
					relPath, err := filepath.Rel(currentDir, absPath)
					if err != nil {
						errCh <- fmt.Errorf("error getting relative path: %s", err)
						return
					}
					deletedFiles[relPath] = true
				}
			}

			errCh <- nil
		}()

		// combine `git ls-files` and `git ls-files --others --exclude-standard`
		// to get all files in the repo

		numRoutines++
		go func() {
			// get all tracked files in the repo
			cmd := exec.Command("git", "ls-files")
			cmd.Dir = baseDir
			out, err := cmd.Output()

			if err != nil {
				errCh <- fmt.Errorf("error getting files in git repo: %s", err)
				return
			}

			files := strings.Split(string(out), "\n")

			mu.Lock()
			defer mu.Unlock()
			for _, file := range files {
				absFile := filepath.Join(baseDir, file)
				relFile, err := filepath.Rel(currentDir, absFile)

				if err != nil {
					errCh <- fmt.Errorf("error getting relative path: %s", err)
					return
				}

				if ignored != nil && ignored.MatchesPath(relFile) {
					continue
				}

				activePaths[relFile] = true

				parentDir := relFile
				for parentDir != "." && parentDir != "/" && parentDir != "" {
					parentDir = filepath.Dir(parentDir)
					activeDirs[parentDir] = true
				}
			}

			errCh <- nil
		}()

		// get all untracked non-ignored files in the repo
		numRoutines++
		go func() {
			cmd := exec.Command("git", "ls-files", "--others", "--exclude-standard")
			cmd.Dir = baseDir
			out, err := cmd.Output()

			if err != nil {
				errCh <- fmt.Errorf("error getting untracked files in git repo: %s", err)
				return
			}

			files := strings.Split(string(out), "\n")

			mu.Lock()
			defer mu.Unlock()
			for _, file := range files {
				absFile := filepath.Join(baseDir, file)
				relFile, err := filepath.Rel(currentDir, absFile)

				if err != nil {
					errCh <- fmt.Errorf("error getting relative path: %s", err)
					return
				}

				if ignored != nil && ignored.MatchesPath(relFile) {
					continue
				}

				activePaths[relFile] = true

				parentDir := relFile
				for parentDir != "." && parentDir != "/" && parentDir != "" {
					parentDir = filepath.Dir(parentDir)
					activeDirs[parentDir] = true
				}
			}

			errCh <- nil
		}()
	}

	// get all paths in the directory
	numRoutines++
	go func() {
		err = filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				if info.Name() == ".git" {
					return filepath.SkipDir
				}
				if info.Name() == ".desktop-cleaner" || info.Name() == ".desktop-cleaner-dev" {
					return filepath.SkipDir
				}

				relPath, err := filepath.Rel(currentDir, path)
				if err != nil {
					return err
				}

				allDirs[relPath] = true

				if ignored != nil && ignored.MatchesPath(relPath) {
					return filepath.SkipDir
				}
			} else {
				relPath, err := filepath.Rel(currentDir, path)
				if err != nil {
					return err
				}

				allPaths[relPath] = true

				if ignored != nil && ignored.MatchesPath(relPath) {
					return nil
				}

				if !isGitRepo {
					mu.Lock()
					defer mu.Unlock()
					activePaths[relPath] = true

					parentDir := relPath
					for parentDir != "." && parentDir != "/" && parentDir != "" {
						parentDir = filepath.Dir(parentDir)
						activeDirs[parentDir] = true
					}
				}
			}

			return nil
		})

		if err != nil {
			errCh <- fmt.Errorf("error walking directory: %s", err)
			return
		}

		errCh <- nil
	}()

	for i := 0; i < numRoutines; i++ {
		err := <-errCh
		if err != nil {
			return nil, err
		}
	}

	for dir := range allDirs {
		allPaths[dir] = true
	}

	for dir := range activeDirs {
		activePaths[dir] = true
	}

	// remove deleted files from active paths
	for path := range deletedFiles {
		delete(activePaths, path)
	}

	ignoredPaths := map[string]string{}
	for path := range allPaths {
		if _, ok := activePaths[path]; !ok {
			if ignored != nil && ignored.MatchesPath(path) {
				ignoredPaths[path] = "dekstop-cleaner"
			} else {
				ignoredPaths[path] = "git"
			}
		}
	}

	return &CleanerPaths{
		ActivePaths:           activePaths,
		AllPaths:              allPaths,
		DesktopCleanerIgnored: ignored,
		IgnoredPaths:          ignoredPaths,
	}, nil
}

func GetDesktopCleanerIgnore(dir string) (*ignore.GitIgnore, error) {
	ignorePath := filepath.Join(dir, ".desktop-cleaner-ignore")

	if _, err := os.Stat(ignorePath); err == nil {
		ignored, err := ignore.CompileIgnoreFile(ignorePath)

		if err != nil {
			return nil, fmt.Errorf("error reading .desktop-cleaner-ignore file: %s", err)
		}

		return ignored, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("error checking for .desktop-cleaner-ignore file: %s", err)
	}

	return nil, nil
}

func GetParentProjectIdsWithPaths() ([][2]string, error) {
	var parentProjectIds [][2]string
	currentDir := filepath.Dir(Cwd)

	for currentDir != "/" {
		HomeDCDir := findDesktopCleaner(currentDir)
		projectSettingsPath := filepath.Join(HomeDCDir, "project.json")
		if _, err := os.Stat(projectSettingsPath); err == nil {
			bytes, err := os.ReadFile(projectSettingsPath)
			if err != nil {
				return nil, fmt.Errorf("error reading projectId file: %s", err)
			}

			var settings types.CurrentProjectSettings
			err = json.Unmarshal(bytes, &settings)

			if err != nil {
				term.OutputErrorAndExit("error unmarshalling project.json: %v", err)
			}

			projectId := string(settings.Id)
			parentProjectIds = append(parentProjectIds, [2]string{currentDir, projectId})
		}
		currentDir = filepath.Dir(currentDir)
	}

	return parentProjectIds, nil
}

func GetChildProjectIdsWithPaths(ctx context.Context) ([][2]string, error) {
	var childProjectIds [][2]string

	err := filepath.Walk(Cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// if permission denied, skip the path
			if os.IsPermission(err) {
				if info.IsDir() {
					return filepath.SkipDir
				} else {
					return nil
				}
			}

			return err
		}

		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("context timeout")
		default:
		}

		if info.IsDir() && path != Cwd {
			HomeDCDir := findDesktopCleaner(path)
			projectSettingsPath := filepath.Join(HomeDCDir, "project.json")
			if _, err := os.Stat(projectSettingsPath); err == nil {
				bytes, err := os.ReadFile(projectSettingsPath)
				if err != nil {
					return fmt.Errorf("error reading projectId file: %s", err)
				}
				var settings types.CurrentProjectSettings
				err = json.Unmarshal(bytes, &settings)

				if err != nil {
					term.OutputErrorAndExit("error unmarshalling project.json: %v", err)
				}

				projectId := string(settings.Id)
				childProjectIds = append(childProjectIds, [2]string{path, projectId})
			}
		}
		return nil
	})

	if err != nil {
		if err.Error() == "context timeout" {
			return childProjectIds, nil
		}

		return nil, fmt.Errorf("error walking the path %s: %s", Cwd, err)
	}

	return childProjectIds, nil
}

func GetBaseDirForContexts(contexts []*shared.Context) string {
	var paths []string

	for _, context := range contexts {
		if context.FilePath != "" {
			paths = append(paths, context.FilePath)
		}
	}

	return GetBaseDirForFilePaths(paths)
}

func GetBaseDirForFilePaths(paths []string) string {
	baseDir := ProjectRoot
	dirsUp := 0

	for _, path := range paths {
		currentDir := ProjectRoot

		pathSplit := strings.Split(path, string(os.PathSeparator))

		n := 0
		for _, p := range pathSplit {
			if p == ".." {
				n++
				currentDir = filepath.Dir(currentDir)
			} else {
				break
			}
		}

		if n > dirsUp {
			dirsUp = n
			baseDir = currentDir
		}
	}

	return baseDir
}

func findDesktopCleaner(baseDir string) string {
	var dir string
	if os.Getenv("DESKTOP_CLEANER_ENV") == "development" {
		dir = filepath.Join(baseDir, ".desktop-cleaner-dev")
	} else {
		dir = filepath.Join(baseDir, ".desktop-cleaner")
	}
	if _, err := os.Stat(dir); !errors.Is(err, fs.ErrNotExist) {
		return dir
	}

	return ""
}

func isCommandAvailable(name string) bool {
	cmd := exec.Command(name, "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}
