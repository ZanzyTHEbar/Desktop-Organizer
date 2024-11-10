package deskfs

import (
	"time"
)

type NodeType int

const (
	Directory NodeType = iota
	File
)

// FileNode represents a file with metadata
type FileNode struct {
	Path       string
	Name       string
	Extension  string
	Size       int64
	ModifiedAt time.Time
	Metadata   Metadata
}

type DirectoryNode struct {
	Path     string
	Type     NodeType
	Parent   *DirectoryNode
	Children []*DirectoryNode
	Files    []*FileNode
	Metadata Metadata
}

// NewDirectoryNode creates a new DirectoryNode
func NewDirectoryNode(path string, parent *DirectoryNode) *DirectoryNode {
	return &DirectoryNode{
		Path:     path,
		Parent:   parent,
		Children: []*DirectoryNode{},
		Files:    []*FileNode{},
	}
}

// AddChildDirectory adds a child directory to the current directory
func (directorynode *DirectoryNode) AddChildDirectory(path string) *DirectoryNode {
	child := NewDirectoryNode(path, directorynode)
	directorynode.Children = append(directorynode.Children, child)
	return child
}

// AddFile adds a file to the current directory
func (directorynode *DirectoryNode) AddFile(file *FileNode) *DirectoryNode {
	directorynode.Files = append(directorynode.Files, file)
	return directorynode
}

func (directorynode *DirectoryNode) String() string {
	return directorynode.Path
}

func (directorynode *DirectoryNode) IsRoot() bool {
	return directorynode.Parent == nil
}

func (directorynode *DirectoryNode) IsLeaf() bool {
	return len(directorynode.Children) == 0
}

func (directorynode *DirectoryNode) IsFile() bool {
	return directorynode.Type == File
}

func (directorynode *DirectoryNode) IsDir() bool {
	return directorynode.Type == Directory
}
