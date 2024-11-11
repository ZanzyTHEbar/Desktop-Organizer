package db

import (
	"database/sql"
	"desktop-cleaner/internal/graph"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	_ "github.com/tursodatabase/go-libsql"
)

// CentralDBProvider tracks the locations of all workspaces.
type CentralDBProvider struct {
	db            *sql.DB
	DirectoryTree *graph.DirectoryTree
}

const centralDBFileName = "central.db"

// NewCentralDBProvider opens or initializes the central database at the binary location.
func NewCentralDBProvider() (*CentralDBProvider, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not get user home directory: %v", err)
	}

	// Construct config path
	configPath := filepath.Join(homeDir, ".config", "desktop_cleaner")

	// Ensure the config directory exists
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return nil, fmt.Errorf("could not create config directory: %v", err)
	}

	dbPath := filepath.Join(configPath, centralDBFileName)

	slog.Info("Central database path:", "path", dbPath)

	db, err := ConnectToDB(dbPath)
	if err != nil {
		return nil, err
	}

	provider := &CentralDBProvider{db: db}
	if err := provider.init(); err != nil {
		return nil, err
	}
	return provider, nil
}

// init sets up the central database tables.
func (c *CentralDBProvider) init() error {
	_, err := c.db.Exec(`CREATE TABLE IF NOT EXISTS workspaces (
		id TEXT PRIMARY KEY UNIQUE,
		root_path TEXT,
		config TEXT
		time_stamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

// AddWorkspace adds a new workspace to the central database and returns its ID.
func (c *CentralDBProvider) AddWorkspace(rootPath, config string) (int, error) {
	slog.Debug(fmt.Sprintf("Adding workspace with root path %s\n", rootPath))

	// Create a new workspace entry in the database
	result, err := c.db.Exec("INSERT INTO workspaces (root_path, config) VALUES (?, ?)", rootPath, config)
	if err != nil {
		return 0, err
	}

	// Create the workspace directory and workspace database

	slog.Debug(fmt.Sprintf("Successfully created Workspace"))

	id, err := result.LastInsertId()
	return int(id), err
}

func (c *CentralDBProvider) UpdateWorkspaceConfig(workspaceID uuid.UUID, config string) (bool, error) {
	_, err := c.db.Exec("UPDATE workspaces SET config = ? WHERE id = ?", config, workspaceID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *CentralDBProvider) GetWorkspace(id uuid.UUID) (*Workspace, error) {
	var workspace Workspace
	err := c.db.QueryRow("SELECT * FROM workspaces WHERE id = ?", id).Scan(&workspace.ID, &workspace.RootPath, &workspace.Config)
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

// GetWorkspacePath retrieves the root path of a workspace by its ID.
func (c *CentralDBProvider) GetWorkspacePath(workspaceID uuid.UUID) (string, error) {
	var rootPath string
	err := c.db.QueryRow("SELECT root_path FROM workspaces WHERE id = ?", workspaceID).Scan(&rootPath)
	return rootPath, err
}

func (c *CentralDBProvider) GetWorkspaceID(rootPath string) (int, error) {
	var id int
	err := c.db.QueryRow("SELECT id FROM workspaces WHERE root_path = ?", rootPath).Scan(&id)
	return id, err
}

func (c *CentralDBProvider) GetWorkspaceConfig(workspaceID uuid.UUID) (string, error) {
	var config string
	err := c.db.QueryRow("SELECT config FROM workspaces WHERE id = ?", workspaceID).Scan(&config)
	return config, err
}

func (c *CentralDBProvider) SetWorkspaceConfig(workspaceID uuid.UUID, config string) error {
	_, err := c.db.Exec("UPDATE workspaces SET config = ? WHERE id = ?", config, workspaceID)
	return err
}

func (c *CentralDBProvider) DeleteWorkspace(workspaceID uuid.UUID) error {
	_, err := c.db.Exec("DELETE FROM workspaces WHERE id = ?", workspaceID)
	return err
}

func (c *CentralDBProvider) ListWorkspaces() ([]Workspace, error) {
	rows, err := c.db.Query("SELECT id, root_path, config, timestamp FROM workspaces ORDER BY time_stamp ASC;")
	if err != nil {
		return nil, fmt.Errorf("failed to query workspaces: %v", err)
	}
	defer rows.Close()

	var workspaces []Workspace

	for rows.Next() {
		var workspace Workspace
		// Scan directly into the Workspace struct fields
		if err := rows.Scan(&workspace.ID, &workspace.RootPath, &workspace.Config); err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %v", err)
		}
		workspaces = append(workspaces, workspace)
	}

	// Check for any errors encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	return workspaces, nil
}

func (c *CentralDBProvider) WorkspaceExists(workspaceID uuid.UUID) (bool, error) {
	var exists bool
	err := c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM workspaces WHERE id = ?)", workspaceID).Scan(&exists)
	return exists, err
}

// Close closes the central database connection.
func (c *CentralDBProvider) Close() error {
	return c.db.Close()
}
