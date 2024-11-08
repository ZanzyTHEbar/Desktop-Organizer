package deskfs

import (
	"context"
	"desktop-cleaner/internal/terminal"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"github.com/ZanzyTHEbar/assert-lib"

	ignore "github.com/sabhiram/go-gitignore"
)

type FilePathParams struct {
	RemoveAfter     bool
	NamesOnly       bool
	ForceSkipIgnore bool
	Recursive       bool
	MaxDepth        int
	GitEnabled      bool
	CopyFiles       bool
	SourceDir       string
	TargetDir       string
	DryRun          bool
}

type DesktopFS struct {
	HomeDir        string
	Cwd            string
	CacheDir       string
	HomeDCDir      string
	DirectoryTree  *DirectoryTree
	InstanceConfig *DeskFSConfig
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

	homeDCDir := findDesktopCleaner(cwd)
	cacheDir := filepath.Join(homeDCDir, ".cache")

	return &DesktopFS{
		HomeDir:   home,
		Cwd:       cwd,
		CacheDir:  cacheDir,
		HomeDCDir: homeDCDir,
		term:      term,
	}
}

// CalculateMaxDepth calculates the maximum depth of the directory structure in `sourceDir`.
func CalculateMaxDepth(sourceDir string) (int, error) {
	if sourceDir == "" {
		return 0, fmt.Errorf("source directory path cannot be empty")
	}

	// Initialize the maximum depth counter
	maxDepth := 0

	// Walk through the directory structure of sourceDir
	err := filepath.WalkDir(sourceDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate depth relative to sourceDir
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		// Calculate the depth by counting separators in the relative path
		if relPath != "." { // Skip the root itself
			depth := strings.Count(relPath, string(os.PathSeparator)) + 1
			if depth > maxDepth {
				maxDepth = depth
			}
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("error calculating max depth: %w", err)
	}

	return maxDepth, nil
}

// Move or copy files based on the configuration
func (dfs *DesktopFS) EnhancedOrganize(cfg *DeskFSConfig, params *FilePathParams) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure context is canceled after function e

	if params.GitEnabled {
		// Clear uncommitted changes or stash them based on user input
		if err := dfs.clearChangesIfNeeded(dfs.Cwd, params); err != nil {
			return fmt.Errorf("failed to clear changes: %w", err)
		}

		if err := dfs.handleUncommittedChanges(dfs.Cwd, params); err != nil {
			return fmt.Errorf("failed to handle uncommitted changes: %w", err)
		}
	}

	// Calculate the maximum depth of SourceDir
	maxDepth, err := CalculateMaxDepth(params.SourceDir)
	if err != nil {
		return fmt.Errorf("failed to calculate max depth: %w", err)
	}

	if err := dfs.buildTreeAndCache(params.SourceDir, params.Recursive, maxDepth); err != nil {
		return fmt.Errorf("failed to build directory tree: %w", err)
	}

	var wg sync.WaitGroup
	var once sync.Once
	errCh := make(chan error, 1)

	// Traverse and organize files based on config
	dfs.traverseAndOrganize(ctx, cancel, dfs.DirectoryTree.Root, cfg, params, &wg, errCh)

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		once.Do(func() { close(errCh) })
	}()

	if err, ok := <-errCh; ok {
		cancel() // Cancel ongoing operations
		return fmt.Errorf("failed to organize files: %w", err)
	}

	// Commit changes if Git is enabled
	if params.GitEnabled {
		if err := dfs.GitAddAndCommit(dfs.Cwd, fmt.Sprintf("Organized files for %s", dfs.Cwd)); err != nil {
			return fmt.Errorf("failed to commit to git: %w", err)
		}

		// Pop the stash if any changes were stashed before organizing
		if err := dfs.GitStashPop(dfs.Cwd, true); err != nil {
			return fmt.Errorf("error popping git stash after organizing: %w", err)
		}
	}

	return nil
}

func (dfs *DesktopFS) InitConfig(optionalConfigPath *string) {
	// Call NewConfig with the provided path (can be nil if no path is specified)
	config := NewIntermediateConfig(optionalConfigPath)

	fmt.Printf("Config: % #v\n", config)

	deskfsConfig := NewDeskFSConfig()

	// Build FileTypeTree
	deskfsConfig = deskfsConfig.BuildFileTypeTree(config)

	// Set the loaded configuration for this instance
	dfs.InstanceConfig = deskfsConfig
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

// Copy copies a file or directory to the destination path.
// It uses recursion for directories if the recursive flag is enabled.
func (dfs *DesktopFS) Copy(node *DirectoryNode, dst string, recursive bool, remove bool, dryrun bool) error {
	if len(node.Children) > 0 || len(node.Files) > 0 { // Check if node is a directory
		if !recursive {
			return fmt.Errorf("source is a directory, use recursive flag to copy directories")
		}

		// Ensure destination directory exists
		if err := os.MkdirAll(dst, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dst, err)
		}

		// Copy each child directory
		for _, childDir := range node.Children {
			childDst := filepath.Join(dst, childDir.Path)
			if dryrun {
				fmt.Printf("Dry run: moving %s to %s\n", node.Path, dst)
				return nil
			}
			if err := dfs.Copy(childDir, childDst, recursive, remove, dryrun); err != nil {
				return err
			}
		}

		// Copy each file in the directory
		for _, fileNode := range node.Files {
			fileDst := filepath.Join(dst, fileNode.Path)
			if dryrun {
				fmt.Printf("Dry run: moving %s to %s\n", fileNode.Path, dst)
				return nil
			}
			if err := dfs.copyFile(fileNode, fileDst, remove, dryrun); err != nil {
				return err
			}
		}

		// Optionally remove the original directory after copying
		if remove {
			return os.RemoveAll(node.Path)
		}
		return nil
	}
	return fmt.Errorf("node has no files or directories to copy")
}

// Helper function for copying a file
func (dfs *DesktopFS) copyFile(fileNode *FileNode, dst string, remove bool, dryrun bool) error {

	if dryrun {
		fmt.Printf("Dry run: moving %s to %s\n", fileNode.Path, dst)
		return nil
	}

	srcFile, err := os.Open(fileNode.Path)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", fileNode.Path, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file %s to %s: %w", fileNode.Path, dst, err)
	}

	// Optionally remove the original file after copying
	if remove {
		if err := os.Remove(fileNode.Path); err != nil {
			return fmt.Errorf("failed to remove original file %s after copy: %w", fileNode.Path, err)
		}
	}
	return nil
}

// Move attempts to move a file or directory from src to dst.
// If a cross-device link error occurs, it falls back to copying and deleting the original.
func (dfs *DesktopFS) Move(node *DirectoryNode, dst string, recursive bool, dryrun bool) error {

	if dryrun {
		fmt.Printf("Dry run: moving %s to %s\n", node.Path, dst)
		return nil
	}

	// Try renaming (moving) the directory node directly
	if err := os.Rename(node.Path, dst); err != nil {
		// If we encounter a cross-device link error, fall back to copy and delete
		if linkErr, ok := err.(*os.LinkError); ok && linkErr.Err == syscall.EXDEV {
			slog.Warn(fmt.Sprintf("Cross-device error detected: falling back to copy for %s\n", node.Path))
			if err := dfs.Copy(node, dst, recursive, true, dryrun); err != nil {
				return fmt.Errorf("failed to copy file for cross-device move: %w", err)
			}
			return nil
		} else {
			return fmt.Errorf("failed to move directory: %w", err)
		}
	}
	return nil
}

// MoveToTrash moves a file or directory to the trash (cache) directory
func (dfs *DesktopFS) MoveToTrash(node *DirectoryNode) error {
	dst := filepath.Join(dfs.CacheDir, filepath.Base(node.Path))
	return os.Rename(node.Path, dst)
}

// buildTreeAndCache recursively builds a directory tree and populates a cache
func (dfs *DesktopFS) buildTreeAndCache(rootPath string, recursive bool, maxDepth int) error {

	// Initialize the DirectoryTree and Cache
	if dfs.DirectoryTree == nil {
		newDirectoryTree, err := NewDirectoryTree(rootPath)
		if err != nil {
			return fmt.Errorf("failed to create directory tree: %w", err)
		}
		dfs.DirectoryTree = newDirectoryTree
	}

	if dfs.DirectoryTree.Cache == nil {
		dfs.DirectoryTree.Cache = make(map[string]*DirectoryNode)
	}

	return dfs.buildTreeNodes(dfs.DirectoryTree.Root, recursive, maxDepth, 0)
}

// Recursive helper to populate the directory tree with DirectoryNode entries
func (dfs *DesktopFS) buildTreeNodes(node *DirectoryNode, recursive bool, maxDepth int, currentDepth int) error {
	// Check if the current depth exceeds the maxDepth
	if currentDepth > maxDepth {
		fmt.Printf("Max depth of %d reached at %s. Skipping deeper levels.\n", maxDepth, node.Path)
		return nil
	}

	entries, err := os.ReadDir(node.Path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		childPath := filepath.Join(node.Path, entry.Name())
		var child *DirectoryNode

		if entry.IsDir() {
			childDir := NewDirectoryNode(childPath, node)
			node.Children = append(node.Children, childDir)
			dfs.DirectoryTree.SafeCacheSet(childPath, childDir)

			if !recursive {
				continue
			}

			if err := dfs.buildTreeNodes(childDir, recursive, maxDepth, currentDepth+1); err != nil {
				return err
			}
		} else {
			entryInfo, err := entry.Info()
			if err != nil {
				slog.Warn(fmt.Sprintf("Error getting file info for %s: %v", entry.Name(), err))
			}

			size := entryInfo.Size()
			modtime := entryInfo.ModTime()

			childFile := &FileNode{
				Path:       childPath,
				Name:       entry.Name(),
				Extension:  strings.ToLower(filepath.Ext(entry.Name())),
				Size:       size,
				ModifiedAt: modtime,
			}
			child = node.AddFile(childFile)
			dfs.DirectoryTree.SafeCacheSet(childPath, child)
		}
	}

	return nil
}

// traverseAndOrganize traverses the tree and organizes files based on the configuration
func (dfs *DesktopFS) traverseAndOrganize(ctx context.Context, cancel context.CancelFunc, node *DirectoryNode, cfg *DeskFSConfig, params *FilePathParams, wg *sync.WaitGroup, errCh chan error) {
	// Process each file within the directory
	for _, fileNode := range node.Files {
		wg.Add(1)
		go func(fileNode *FileNode) {
			defer wg.Done()

			// Respect context cancellation
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Determine the target folder based on file extension
			targetDir, found := dfs.determineTargetFolder(ctx, fileNode, cfg)
			if !found {
				fmt.Printf("Skipping file %s as no target path found\n", fileNode.Name)
				return // Skip files without a target folder
			}

			// Construct the correct destination directory and path
			destDir := filepath.Join(params.TargetDir, targetDir)
			fmt.Printf("Creating directory: %s\n", destDir)
			destPath := filepath.Join(destDir, filepath.Base(fileNode.Path)) // Only the base name
			fmt.Printf("Moving file %s to %s\n", fileNode.Path, destPath)

			// Ensure target directory exists before moving or copying files
			if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
				select {
				case errCh <- fmt.Errorf("failed to create target directory %s: %w", destDir, err):
					cancel() // Cancel all ongoing operations
				default:
				}
				return
			}

			// Copy or move the file based on params
			var fileErr error
			if params.CopyFiles {
				fileErr = dfs.copyFile(fileNode, destPath, params.RemoveAfter, params.DryRun)
			} else {
				fileErr = dfs.Move(&DirectoryNode{Path: fileNode.Path}, destPath, false, params.DryRun)
			}
			// Send error to errCh and cancel context on first failure
			if fileErr != nil {
				select {
				case errCh <- fmt.Errorf("file operation failed: %w", fileErr):
					cancel() // Cancel all ongoing operations
				default:
				}
			}

		}(fileNode)
	}

	// Process each child directory
	for _, childDir := range node.Children {
		if params.Recursive {
			dfs.traverseAndOrganize(ctx, cancel, childDir, cfg, params, wg, errCh)
		}
	}
}

// determineTargetFolder traverses the FileTypeTree in DeskFSConfig to find the appropriate folder
// based on the file's extension. It returns the path to the target folder if a match is found.
func (dfs *DesktopFS) determineTargetFolder(ctx context.Context, fileNode *FileNode, cfg *DeskFSConfig) (string, bool) {
	ext := fileNode.Extension

	path, found := dfs.findFolderForExtension(ctx, cfg.FileTypeTree.Root, ext)
	if found {
		fmt.Printf("File %s with extension %s mapped to path: %s\n", fileNode.Name, ext, path)
	} else {
		fmt.Printf("No mapping found for file %s with extension %s\n", fileNode.Name, ext)
	}
	return path, found
}

// Helper recursive function to search for the appropriate folder in the FileTypeTree.
func (dfs *DesktopFS) findFolderForExtension(ctx context.Context, node *FileTypeNode, ext string) (string, bool) {
	// Traverse the tree to find a matching extension in the nodes
	if node.AllowsExtension(ext) {
		return buildPathFromNode(ctx, node), true
	}

	// Continue to search for extensions in children
	for _, child := range node.Children {
		if path, found := dfs.findFolderForExtension(ctx, child, ext); found {
			return path, true
		}
	}

	return "", false
}

// buildPathFromNode constructs the path from the root to the given node.
func buildPathFromNode(ctx context.Context, node *FileTypeNode) string {
	// If this is the root node, start from its children
	if node.IsRoot() && len(node.Children) >= 1 {
		// Start from the first child to avoid adding "root" to the path
		node = node.Children[0]
	}

	assertHandler := assert.NewAssertHandler()
	assertHandler.SetExitFunc(func(int) {
		slog.Error("[Path Assertion Error]: assertion failure")
	})

	// Ensure that the node has a valid name
	if node.Name == "" {
		assertHandler.Never(ctx, fmt.Sprintf("Node has an invalid or empty name: %v", node), slog.Error)
	}

	pathSegments := []string{node.Name}
	for current := node.Parent; current != nil; current = current.Parent {
		assertHandler.Assert(ctx, current.Name != "", "Invalid node name detected", slog.Error)
		if current.IsRoot() {
			break // Skip "root" in the path
		}
		pathSegments = append([]string{current.Name}, pathSegments...)
	}
	assertHandler.Assert(ctx, node.IsRoot() || node.Parent != nil, "Root Node should not have a parent", slog.Error)

	finalPath := filepath.Join(pathSegments...)
	fmt.Printf("Final constructed path (with case preserved): %s\n", finalPath)

	assertHandler.Assert(ctx, finalPath != "", "Constructed path should not be empty", slog.Error)

	return finalPath
}

func findDesktopCleaner(baseDir string) string {
	var dir string
	const devEnv = "development"
	const prodEnv = "production"
	const folderName = ".desktop-cleaner"
	const env = "DESKTOP_CLEANER_ENV"

	envValue, envSet := os.LookupEnv(env)

	if !envSet {
		return ""
	}

	dir = filepath.Join(baseDir, folderName+"-"+envValue)
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		return baseDir
	}

	return dir
}
