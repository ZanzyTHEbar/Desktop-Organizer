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

// AddRelationships adds relationships between nodes in the DirectoryTree
func (directorynode *DirectoryNode) AddRelationships(node *DirectoryNode) {
	// Add parent-child relationships
	for _, child := range node.Children {
		directorynode.AddRelationship(node, child.Path, "contains")
		directorynode.AddRelationship(child, node.Path, "parent")
		// Recursively add relationships to children
		directorynode.AddRelationships(child)
	}

	for _, file := range node.Files {
		directorynode.AddRelationship(node, file.Path, "contains")
	}

	// Add temporal relationships
	allNodes := directorynode.collectAllNodes(node)
	for i, nodeA := range allNodes {
		for j, nodeB := range allNodes {
			if i != j && nodeA.Metadata.ModifiedAt.Sub(nodeB.Metadata.ModifiedAt) < time.Hour*24 {
				// If nodes were modified within a day of each other, relate them
				relationship := Relationship{
					RelatedNode: nodeB.Path,
					Type:        "modified-around-same-time",
				}
				nodeA.Metadata.Relationships = append(nodeA.Metadata.Relationships, relationship)
			}
		}
	}
}

func (directorynode *DirectoryNode) AddRelationship(node *DirectoryNode, relatedPath string, relType string) {
	relationship := Relationship{
		RelatedNode: relatedPath,
		Type:        relType,
	}
	node.Metadata.Relationships = append(node.Metadata.Relationships, relationship)
}

// AddChildDirectory adds a child directory to the current directory
func (directorynode *DirectoryNode) AddChildDirectory(path string) *DirectoryNode {
	child := NewDirectoryNode(path, directorynode)
	directorynode.Children = append(directorynode.Children, child)
	return child
}

// collectAllNodes collects all nodes (both directories and files) from the given DirectoryNode
func (directorynode *DirectoryNode) collectAllNodes(node *DirectoryNode) []*DirectoryNode {
	var nodes []*DirectoryNode
	nodes = append(nodes, node)
	for _, child := range node.Children {
		nodes = append(nodes, directorynode.collectAllNodes(child)...)
	}
	return nodes
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
