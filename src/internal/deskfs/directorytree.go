package deskfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type DirectoryTree struct {
	Root  *DirectoryNode
	Cache map[string]*DirectoryNode
	mu    sync.Mutex
}

func NewDirectoryTree(rootPath string) (*DirectoryTree, error) {
	if rootPath == "" {
		return nil, fmt.Errorf("root path cannot be empty")
	}

	return &DirectoryTree{
		Root: &DirectoryNode{
			Path:     rootPath,
			Type:     Directory,
			Parent:   nil,
			Children: make([]*DirectoryNode, 0),
			Files:    make([]*FileNode, 0),
		},
		Cache: make(map[string]*DirectoryNode),
	}, nil
}

// Flatten recursively collects all directories and files in a flat list of paths
func (tree *DirectoryTree) Flatten() []string {

	var paths []string
	tree.flattenNode(tree.Root, tree.Root.Path, &paths)
	return paths
}

// flattenNode is a helper function for Flatten, processing each node recursively
func (tree *DirectoryTree) flattenNode(node *DirectoryNode, currentPath string, paths *[]string) {

	// Add current directory path to paths
	*paths = append(*paths, currentPath)

	// Recursively process each child directory
	for _, child := range node.Children {
		childPath := filepath.Join(currentPath, child.Path)
		tree.flattenNode(child, childPath, paths)
	}

	// Add all files in this directory to paths
	for _, file := range node.Files {
		filePath := filepath.Join(currentPath, file.Path)
		*paths = append(*paths, filePath)
	}
}

// SafeCacheSet safely sets a value in the Cache map
func (tree *DirectoryTree) SafeCacheSet(key string, value *DirectoryNode) {
	tree.mu.Lock()
	defer tree.mu.Unlock()

	tree.Cache[key] = value
}

// SafeCacheGet safely retrieves a value from the Cache map
func (tree *DirectoryTree) SafeCacheGet(key string) (*DirectoryNode, bool) {
	tree.mu.Lock()
	defer tree.mu.Unlock()

	value, exists := tree.Cache[key]
	return value, exists
}

// AddDirectory adds a directory node to the tree at a specified path
func (tree *DirectoryTree) AddDirectory(path string) (*DirectoryNode, error) {

	node := tree.Root
	segments := strings.Split(path, string(os.PathSeparator))
	for _, segment := range segments {
		found := false
		for _, child := range node.Children {
			if child.Path == segment && child.Type == Directory {
				node = child
				found = true
				break
			}
		}
		if !found {
			// Create missing directories in path
			newDir := &DirectoryNode{
				Path:     segment,
				Type:     Directory,
				Parent:   node,
				Children: []*DirectoryNode{},
				Files:    []*FileNode{},
			}
			node.Children = append(node.Children, newDir)
			node = newDir
		}
	}
	return node, nil
}

// FindOrCreatePath traverses the tree to find or create a directory path
func (tree *DirectoryTree) FindOrCreatePath(path []string) *DirectoryNode {

	current := tree.Root
	for _, dir := range path {
		var next *DirectoryNode
		for _, child := range current.Children {
			if child.Path == dir {
				next = child
				break
			}
		}
		if next == nil {
			next = current.AddChildDirectory(dir)
		}
		current = next
	}
	return current
}

// AddFile adds a file node to the tree at a specified path.
// If intermediate directories don't exist, it creates them.
func (tree *DirectoryTree) AddFile(path string, filePath string, size int64, modifiedAt time.Time) error {

	// Split the path into directories and then find or create the path
	targetNode := tree.FindOrCreatePath(filepath.SplitList(path))

	// Now that we're at the target directory, add the file node
	targetNode.AddFile(&FileNode{
		Path:       filePath,
		Extension:  filepath.Ext(filePath),
		Size:       size,
		ModifiedAt: modifiedAt,
	})

	return nil
}
