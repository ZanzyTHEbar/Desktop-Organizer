package deskfs

import (
	"os"
	"time"
)

// Relationship captures an edge between two nodes with a specific meaning
type Relationship struct {
	RelatedNode string // Path of the related node
	Type        string // Type of the relationship (e.g., "parent", "sibling", "dependency", "contains")
}

// Metadata holds additional information for each node in the DirectoryTree
type Metadata struct {
	Size          int64          // Size of the file or directory
	ModifiedAt    time.Time      // Last modified time
	CreatedAt     time.Time      // Creation time (if available)
	NodeType      string         // "file" or "directory"
	Permissions   os.FileMode    // File permissions
	Owner         string         // Owner of the file (if available)
	Tags          []string       // Tags associated with the file or directory
	Relationships []Relationship // Relationships to other nodes
}

// GenerateMetadata generates metadata for a given file or directory node
func GenerateMetadata(nodePath string) (Metadata, error) {
	fileInfo, err := os.Stat(nodePath)
	if err != nil {
		return Metadata{}, err
	}

	// Get file permissions and modification time
	permissions := fileInfo.Mode()
	modifiedAt := fileInfo.ModTime()

	// Placeholder for createdAt, as it may not be available on all systems
	createdAt := time.Time{}

	// Set NodeType to "file" or "directory"
	nodeType := "file"
	if fileInfo.IsDir() {
		nodeType = "directory"
	}

	// Create metadata struct
	metadata := Metadata{
		Size:        fileInfo.Size(),
		ModifiedAt:  modifiedAt,
		CreatedAt:   createdAt,
		NodeType:    nodeType,
		Permissions: permissions,
		Owner:       "unknown",  // TODO: Owner retrieval can be implemented based on platform
		Tags:        []string{}, // Initialize with an empty list of tags
	}

	return metadata, nil
}

// AddMetadataToTree recursively traverses the DirectoryTree and adds metadata to each node
func (dfs *DesktopFS) AddMetadataToTree(node *DirectoryNode) error {
	// Generate metadata for the current directory node
	metadata, err := GenerateMetadata(node.Path)
	if err != nil {
		return err
	}
	// Add tags to metadata
	AddTagsToMetadata(&metadata)
	node.Metadata = metadata

	// Add metadata to all files within the directory
	for _, fileNode := range node.Files {
		fileMetadata, err := GenerateMetadata(fileNode.Path)
		if err != nil {
			return err
		}
		// Add tags to file metadata
		AddTagsToMetadata(&fileMetadata)
		fileNode.Metadata = fileMetadata
	}

	// Recursively add metadata to child directories
	for _, childDir := range node.Children {
		if err := dfs.AddMetadataToTree(childDir); err != nil {
			return err
		}
	}

	return nil
}

// FlattenMetadata flattens metadata into a map that can be used for LLM input
func (dfs *DesktopFS) FlattenMetadata(node *DirectoryNode) map[string]interface{} {
	flatMetadata := make(map[string]interface{})

	// Add directory node metadata
	flatMetadata[node.Path] = node.Metadata

	// Add files metadata
	for _, fileNode := range node.Files {
		flatMetadata[fileNode.Path] = fileNode.Metadata
	}

	// Recursively add child directory metadata
	for _, childDir := range node.Children {
		childMetadata := dfs.FlattenMetadata(childDir)
		for key, value := range childMetadata {
			flatMetadata[key] = value
		}
	}

	return flatMetadata
}
