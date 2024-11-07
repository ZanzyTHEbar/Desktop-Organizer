package deskfs

import (
	"desktop-cleaner/internal/terminal"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

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

func (dfs *DesktopFS) InitConfig(optionalConfigPath *string) {
	// Call NewConfig with the provided path (can be nil if no path is specified)
	config, err := NewConfig(optionalConfigPath)
	if err != nil {
		dfs.term.OutputErrorAndExit("Error loading configuration: %v", err)
	}

	// Set the loaded configuration for this instance
	dfs.InstanceConfig = config
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
func (dfs *DesktopFS) Copy(node *TreeNode, dst string, recursive bool, remove bool) error {
	if node.Type == Directory {
		if !recursive {
			return fmt.Errorf("source is a directory, use recursive flag to copy directories")
		}

		// Ensure destination directory exists
		if err := os.MkdirAll(dst, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dst, err)
		}

		// Copy each child node within the directory
		for _, child := range node.Children {
			childDst := filepath.Join(dst, filepath.Base(child.Name))
			if err := dfs.Copy(child, childDst, recursive, remove); err != nil {
				return err
			}
		}

		// Optionally remove the original directory after copying
		if remove {
			return os.Remove(node.Name)
		}
		return nil
	}

	// For files, perform the actual copy operation
	srcFile, err := os.Open(node.Name)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", node.Name, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file %s to %s: %w", node.Name, dst, err)
	}

	// Optionally remove the original file after copying
	if remove {
		if err := os.Remove(node.Name); err != nil {
			return fmt.Errorf("failed to remove original file %s after copy: %w", node.Name, err)
		}
	}
	return nil
}

// Move attempts to move a file or directory from src to dst.
// If a cross-device link error occurs, it falls back to copying and deleting the original.
func (dfs *DesktopFS) Move(node *TreeNode, dst string, recursive bool) error {
	// Try renaming (moving) the node directly
	if err := os.Rename(node.Name, dst); err != nil {
		// If we encounter a cross-device link error, fall back to copy and delete
		if linkErr, ok := err.(*os.LinkError); ok && linkErr.Err == syscall.EXDEV {
			if err := dfs.Copy(node, dst, recursive, true); err != nil {
				return fmt.Errorf("failed to copy file for cross-device move: %w", err)
			}
			return nil
		} else {
			return fmt.Errorf("failed to move file: %w", err)
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

// buildTreeAndCache recursively builds a directory tree and populates a cache
func (dfs *DesktopFS) buildTreeAndCache(rootPath string, recursive bool) error {
	// Ensure DirectoryTree and Cache are initialized
	if dfs.DirectoryTree == nil {
		dfs.DirectoryTree = &DirectoryTree{Root: NewTreeNode(rootPath, Directory, nil)}
	}
	if dfs.DirectoryTree.Cache == nil {
		dfs.DirectoryTree.Cache = make(map[string]*TreeNode)
	}

	// Populate the tree starting from the root node
	return dfs.buildTreeNodes(dfs.DirectoryTree.Root, recursive)
}

// Recursive helper to populate the directory tree with TreeNode entries
func (dfs *DesktopFS) buildTreeNodes(node *TreeNode, recursive bool) error {
	entries, err := os.ReadDir(node.Name)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		childNode := &TreeNode{
			Name:     filepath.Join(node.Name, entry.Name()),
			Type:     Directory,
			Children: []*TreeNode{},
			Parent:   node,
		}

		if entry.IsDir() {
			if recursive {
				// Recursively build subdirectories
				if err := dfs.buildTreeNodes(childNode, recursive); err != nil {
					return err
				}
			}
		} else {
			childNode.Type = File
			fileInfo, _ := entry.Info()
			childNode.Metadata = &FileMetadata{
				Extension:  strings.ToLower(filepath.Ext(entry.Name())),
				Size:       fileInfo.Size(),
				ModifiedAt: fileInfo.ModTime(),
			}
		}
		node.Children = append(node.Children, childNode)
		dfs.DirectoryTree.Cache[childNode.Name] = childNode // Add to cache
	}

	return nil
}

// Move or copy files based on the configuration
func (dfs *DesktopFS) EnhancedOrganize(cfg *DeskFSConfig, params *FilePathParams) error {
	if params.GitEnabled {
		// Clear uncommitted changes or stash them based on user input
		if err := dfs.clearChangesIfNeeded(dfs.Cwd, params); err != nil {
			return fmt.Errorf("failed to clear changes: %w", err)
		}

		if err := dfs.handleUncommittedChanges(dfs.Cwd, params); err != nil {
			return fmt.Errorf("failed to handle uncommitted changes: %w", err)
		}
	}

	if err := dfs.buildTreeAndCache(params.SourceDir, params.Recursive); err != nil {
		return fmt.Errorf("failed to build directory tree: %w", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	// Traverse and organize files based on config
	dfs.traverseAndOrganize(dfs.DirectoryTree.Root, cfg, params, &wg, errCh)

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(errCh)
	}()

	if err, ok := <-errCh; ok {
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

// traverseAndOrganize traverses the tree and organizes files based on the configuration
func (dfs *DesktopFS) traverseAndOrganize(node *TreeNode, cfg *DeskFSConfig, params *FilePathParams, wg *sync.WaitGroup, errCh chan error) {
	for _, child := range node.Children {
		if child.Type == Directory {
			if params.Recursive {
				dfs.traverseAndOrganize(child, cfg, params, wg, errCh)
			}
		} else {
			wg.Add(1)
			go func(fileNode *TreeNode) {
				defer wg.Done()

				// Determine the target folder based on file extension
				targetFolder, found := dfs.determineTargetFolder(fileNode, cfg)
				if !found {
					return // Skip files without a target folder
				}

				// Construct destDir with params.TargetDir as the root
				destDir := filepath.Join(params.TargetDir, targetFolder)
				destPath := filepath.Join(destDir, filepath.Base(fileNode.Name))

				// Log paths to check values
				fmt.Printf("Source: %s\nTarget Dir: %s\nDestination Path: %s\n", fileNode.Name, destDir, destPath)

				// Ensure the target directory exists before moving or copying files
				if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
					errCh <- fmt.Errorf("failed to create target directory %s: %w", filepath.Dir(destPath), err)
				}
				// Ensure the target directory exists
				if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
					errCh <- fmt.Errorf("failed to create target directory: %w", err)
					return
				}

				if params.CopyFiles {
					if err := dfs.Copy(fileNode, destPath, params.Recursive, params.RemoveAfter); err != nil {
						errCh <- fmt.Errorf("failed to copy file: %w", err)
						return
					}
				} else {
					if err := dfs.Move(fileNode, destPath, params.Recursive); err != nil {
						errCh <- fmt.Errorf("failed to move file: %w", err)
						return
					}
				}
			}(child)
		}
	}
}

// determineTargetFolder identifies the appropriate folder for a file based on its extension
func (dfs *DesktopFS) determineTargetFolder(fileNode *TreeNode, cfg *DeskFSConfig) (string, bool) {
	ext := fileNode.Metadata.Extension
	for folder, extensions := range cfg.FileTypes {
		for _, allowedExt := range extensions {
			if ext == allowedExt {
				// Include nested directory structure if specified
				if nestedDirs, exists := cfg.NestedDirs[folder]; exists {
					return filepath.Join(folder, filepath.Join(nestedDirs...)), true
				}
				return folder, true
			}
		}
	}
	return "", false
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
