package deskfs

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type NodeType int

const (
	Directory NodeType = iota
	File
)

type FileMetadata struct {
	Extension  string
	Size       int64
	ModifiedAt time.Time
}

type TreeNode struct {
	Name     string
	Type     NodeType
	Metadata *FileMetadata
	Children []*TreeNode
	Parent   *TreeNode
}

type DirectoryTree struct {
	Root  *TreeNode
	Cache map[string]*TreeNode
	sync.Mutex
}

func NewDirectoryTree(rootPath string) (*DirectoryTree, error) {
	if rootPath == "" {
		return nil, fmt.Errorf("root path cannot be empty")
	}

	return &DirectoryTree{
		Root: &TreeNode{
			Name:     rootPath,
			Type:     Directory,
			Metadata: nil,
			Children: []*TreeNode{},
			Parent:   nil,
		},
		Cache: make(map[string]*TreeNode),
	}, nil
}

// AddFile adds a file node to the tree at a specified path

func (dt *DirectoryTree) AddFile(path string, metadata FileMetadata) error {
	node := dt.Root
	segments := strings.Split(path, string(os.PathSeparator))
	for _, segment := range segments {
		found := false
		for _, child := range node.Children {
			if child.Name == segment && child.Type == Directory {
				node = child
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("directory path %s does not exist", path)
		}
	}
	fileNode := &TreeNode{
		Name:     segments[len(segments)-1],
		Type:     File,
		Metadata: &metadata,
		Parent:   node,
	}
	node.Children = append(node.Children, fileNode)
	return nil
}

// AddDirectory adds a directory node to the tree at a specified path
func (t *DirectoryTree) AddDirectory(path string) (*TreeNode, error) {
	node := t.Root
	segments := strings.Split(path, string(os.PathSeparator))
	for _, segment := range segments {
		found := false
		for _, child := range node.Children {
			if child.Name == segment && child.Type == Directory {
				node = child
				found = true
				break
			}
		}
		if !found {
			// Create missing directories in path
			newDir := &TreeNode{
				Name:     segment,
				Type:     Directory,
				Parent:   node,
				Children: []*TreeNode{},
			}
			node.Children = append(node.Children, newDir)
			node = newDir
		}
	}
	return node, nil
}

// FlattenTree flattens the directory tree for LLM context generation
func (t *DirectoryTree) FlattenTree() []string {
	var result []string
	var dfs func(node *TreeNode, depth int)
	dfs = func(node *TreeNode, depth int) {
		indent := strings.Repeat("  ", depth)
		if node.Type == Directory {
			result = append(result, fmt.Sprintf("%s- %s/", indent, node.Name))
			for _, child := range node.Children {
				dfs(child, depth+1)
			}
		} else {
			result = append(result, fmt.Sprintf("%s- %s (%s)", indent, node.Name, node.Metadata.Extension))
		}
	}
	dfs(t.Root, 0)
	return result
}

func (dt *DirectoryTree) AddNode(parent *TreeNode, name string, nodeType NodeType, metadata *FileMetadata) *TreeNode {
	node := &TreeNode{
		Name:     name,
		Type:     nodeType,
		Metadata: metadata,
		Children: make([]*TreeNode, 0),
		Parent:   parent,
	}
	parent.Children = append(parent.Children, node)
	return node
}
