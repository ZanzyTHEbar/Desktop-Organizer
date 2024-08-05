package fs

import (
	"desktop-cleaner/internal/terminal"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	ignore "github.com/sabhiram/go-gitignore"
)

type CleanerPaths struct {
	ActivePaths           map[string]bool
	AllPaths              map[string]bool
	DesktopCleanerIgnored *ignore.GitIgnore
	IgnoredPaths          map[string]string
}

type FilePathParams struct {
	Recursive       bool
	RemoveAfter     bool
	CopyFiles       bool
	NamesOnly       bool
	ForceSkipIgnore bool
	MaxDepth        int
	GitEnabled      bool
}

type DesktopFS struct {
	HomeDir        string
	Cwd            string
	CacheDir       string
	HomeDCDir      string
	ProjectRoot    string
	InstanceConfig *Config
	term           *terminal.Terminal
}

func NewDesktopFS(term *terminal.Terminal) *DesktopFS {
	var err error
	cwd, err := os.Getwd()

	if err != nil {
		term.OutputErrorAndExit("Error getting current working directory: %v", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		term.OutputErrorAndExit("Couldn't find home directory: %v", err)
	}

	var homeDCDir string
	var cacheDir string
	var projectRoot string

	if os.Getenv("DESKTOP_CLEANER_ENV") == "development" {
		homeDCDir = filepath.Join(home, ".desktop-cleaner-dev")
	} else {
		homeDCDir = filepath.Join(home, ".desktop-cleaner")
	}

	err = os.MkdirAll(homeDCDir, os.ModePerm)
	if err != nil {
		term.OutputErrorAndExit("Error creating config directory: %v", err.Error())
	}

	cacheDir = filepath.Join(homeDCDir, "cache")

	homeDCDir = findDesktopCleaner(cwd)
	if homeDCDir != "" {
		projectRoot = cwd
	}

	return &DesktopFS{
		HomeDir:     home,
		Cwd:         cwd,
		CacheDir:    cacheDir,
		HomeDCDir:   homeDCDir,
		ProjectRoot: projectRoot,
		term:        term,
	}
}

func (dfs *DesktopFS) InitConfig(path *string) {
	// Load the configuration file
	config, err := NewConfig(path)
	if err != nil {
		dfs.term.OutputErrorAndExit("Error loading configuration: %v", err)
	}

	dfs.InstanceConfig = config
}

func (dfs *DesktopFS) FindOrCreateDesktopCleaner() (string, bool, error) {
	dfs.HomeDCDir = findDesktopCleaner(dfs.Cwd)
	if dfs.HomeDCDir != "" {
		dfs.ProjectRoot = dfs.Cwd
		return dfs.HomeDCDir, false, nil
	}

	// Determine the directory path
	var dir string
	if os.Getenv("DESKTOP_CLEANER_ENV") == "development" {
		dir = filepath.Join(dfs.Cwd, ".desktop-cleaner-dev")
	} else {
		dir = filepath.Join(dfs.Cwd, ".desktop-cleaner")
	}

	err := os.Mkdir(dir, os.ModePerm)
	if err != nil {
		return "", false, err
	}
	dfs.HomeDCDir = dir
	dfs.ProjectRoot = dfs.Cwd

	return dir, true, nil
}

func (dfs *DesktopFS) ProjectRootIsGitRepo() bool {
	if dfs.ProjectRoot == "" {
		return false
	}

	return IsGitRepo(dfs.ProjectRoot)
}

func (dfs *DesktopFS) GetCleanerPaths(baseDir string) (*CleanerPaths, error) {
	if dfs.ProjectRoot == "" {
		return nil, fmt.Errorf("no project root found")
	}

	return dfs.GetPaths(baseDir, dfs.ProjectRoot)
}

func (dfs *DesktopFS) GetPaths(baseDir, currentDir string) (*CleanerPaths, error) {
	ignored, err := dfs.GetDesktopCleanerIgnore(currentDir)

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

func (dfs *DesktopFS) GetDesktopCleanerIgnore(dir string) (*ignore.GitIgnore, error) {
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

func (dfs *DesktopFS) GetBaseDirForFilePaths(paths []string) string {
	baseDir := dfs.ProjectRoot
	dirsUp := 0

	for _, path := range paths {
		currentDir := dfs.ProjectRoot

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

// Copy a file from src to dst
func (dfs *DesktopFS) CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

// Copy a directory and its contents recursively with concurrency
func (dfs *DesktopFS) CopyDir(srcDir, dstDir string, remove bool) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	errCh := make(chan error, 1)

	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return err
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		wg.Add(1)
		go func(entry os.DirEntry) {
			defer wg.Done()

			srcPath := filepath.Join(srcDir, entry.Name())
			dstPath := filepath.Join(dstDir, entry.Name())

			fileInfo, err := entry.Info()
			if err != nil {
				mu.Lock()
				errCh <- err
				mu.Unlock()
				return
			}

			if fileInfo.IsDir() {
				if err := dfs.CopyDir(srcPath, dstPath, remove); err != nil {
					mu.Lock()
					errCh <- err
					mu.Unlock()
					return
				}
			} else {
				if err := dfs.CopyFile(srcPath, dstPath); err != nil {
					mu.Lock()
					errCh <- err
					mu.Unlock()
					return
				}
				if remove {
					if err := os.Remove(srcPath); err != nil {
						mu.Lock()
						errCh <- err
						mu.Unlock()
						return
					}
				}
			}
		}(entry)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	if err, ok := <-errCh; ok {
		return err
	}

	if remove {
		if err := os.Remove(srcDir); err != nil {
			return err
		}
	}
	return nil
}

// Copy files and directories with support for recursion and remove flag
func (dfs *DesktopFS) Copy(src, dst string, recursive, remove bool) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		if !recursive {
			return fmt.Errorf("source is a directory, use recursive flag to copy directories")
		}
		return dfs.CopyDir(src, dst, remove)
	}
	if err := dfs.CopyFile(src, dst); err != nil {
		return err
	}
	if remove {
		return os.Remove(src)
	}
	return nil
}

func (dfs *DesktopFS) Move(src, dst string, recursive, removeAfter bool) error {
	if recursive {
		return fmt.Errorf("move does not support recursive operation")
	}
	if err := os.Rename(src, dst); err != nil {
		// If rename fails due to cross-device link, fallback to copy and delete
		if linkErr, ok := err.(*os.LinkError); ok && linkErr.Err == syscall.EXDEV {
			if err := dfs.Copy(src, dst, false, removeAfter); err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

// Function to move files to trash
func (dfs *DesktopFS) MoveToTrash(path string) error {
	if err := os.Rename(path, filepath.Join(dfs.CacheDir, filepath.Base(path))); err != nil {
		return err
	}

	return nil
}

// Move or copy files based on the configuration
func (dfs *DesktopFS) EnhancedOrganize(srcDir string, targetDir string, cfg Config, params *FilePathParams) error {

	if targetDir == "" {
		targetDir = srcDir
	}

	if cfg.TargetDir != "" {
		targetDir = cfg.TargetDir
	}

	if params.GitEnabled && dfs.ProjectRootIsGitRepo() {
		if err := InitGitRepo(srcDir); err != nil {
			return err
		}
	}

	var paths []string
	var err error

	if params.GitEnabled {
		cleanerPaths, err := dfs.GetPaths(srcDir, srcDir)
		if err != nil {
			return err
		}
		for path := range cleanerPaths.ActivePaths {
			paths = append(paths, filepath.Join(srcDir, path))
		}
	} else {
		paths, err = dfs.ParseInputPaths([]string{srcDir}, &FilePathParams{Recursive: params.Recursive, NamesOnly: params.NamesOnly})
		slog.Debug("Parsed paths, moving on to organizing files ...")
		if err != nil {
			return err
		}
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for _, path := range paths {
		slog.Debug(fmt.Sprintf("Processing path: %s", path))
		info, err := os.Stat(path)
		if err != nil {
			return err
		}

		// Skip git directory
		if strings.Contains(path, ".git") {
			continue
		}

		// Skip subdirectories unless Recursive flag is set
		if info.IsDir() && path != srcDir {
			if !params.Recursive {
				slog.Warn(fmt.Sprintf("Skipping directory %s - recursion is not set.", path))
				continue
			}
			wg.Add(1)
			go func(_srcDir string) {
				defer wg.Done()
				if err := dfs.EnhancedOrganize(_srcDir, targetDir, cfg, params); err != nil {
					errCh <- err
					return
				}
			}(path)
			continue
		}

		slog.Debug("Generating folder structure ...")

		ext := strings.ToLower(filepath.Ext(info.Name()))
		var folder string
		found := false
		for f, exts := range cfg.FileTypes {
			for _, e := range exts {
				if e == ext {
					folder = f
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		slog.Debug(fmt.Sprintf("Folder: %s, found - %v", folder, found))

		if !found {
			continue
		}

		for _, nested := range cfg.NestedDirs[folder] {
			folder = filepath.Join(folder, nested)
		}

		slog.Debug(fmt.Sprintf("Folder Path: %s", folder))

		destDir := filepath.Join(targetDir, folder)

		slog.Debug(fmt.Sprintf("Destination Directory: %s", destDir))

		if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
			return err
		}

		destPath := filepath.Join(destDir, info.Name())

		slog.Debug(fmt.Sprintf("Destination Path: %s", destPath))

		wg.Add(1)
		go func(srcPath, dstPath string) {
			defer wg.Done()
			if params.CopyFiles {
				if err := dfs.Copy(srcPath, dstPath, params.Recursive, params.RemoveAfter); err != nil {
					errCh <- err
					return
				}
			} else {
				if err := dfs.Move(srcPath, dstPath, params.Recursive, params.RemoveAfter); err != nil {
					errCh <- err
					return
				}
			}
		}(path, destPath)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	if err, ok := <-errCh; ok {
		return fmt.Errorf("failed to organize files: %w", err)
	}

	if params.GitEnabled && dfs.ProjectRootIsGitRepo() {
		if err := GitAddAndCommit(srcDir, "Organized files", true); err != nil {
			return err
		}
	}

	return nil
}

// Parse Paths WITHOUT git support
func (dfs *DesktopFS) ParseInputPaths(fileOrDirPaths []string, params *FilePathParams) ([]string, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	resPaths := []string{}

	for _, path := range fileOrDirPaths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				mu.Lock()
				defer mu.Unlock()
				if firstErr != nil {
					return firstErr // If an error was encountered, stop walking
				}

				// Skip subdirectories unless Recursive flag is set
				if info.IsDir() && p != path && !params.Recursive {
					return filepath.SkipDir
				}

				if info.IsDir() {
					if info.Name() == ".git" {
						return filepath.SkipDir
					}

					// calculate directory depth from base
					depth := strings.Count(path[len(p):], string(filepath.Separator))
					if params.MaxDepth != -1 && depth > params.MaxDepth {
						return filepath.SkipDir
					}

					//if !(params.Recursive || params.NamesOnly) {
					//	return fmt.Errorf("cannot process directory %s: --recursive or --names-only flag not set", path)
					//}

					if params.NamesOnly {
						// add directory name to results
						resPaths = append(resPaths, path)
					}
				} else {
					// add file path to results
					resPaths = append(resPaths, path)
				}

				return nil
			})

			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}(path)
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	slog.Info(fmt.Sprintf("Parsed %d paths: %v", len(resPaths), resPaths))

	return resPaths, nil
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
