package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

// CentralDBProvider tracks the locations of all workspaces.
type CentralDBProvider struct {
	db *sql.DB
}

// NewCentralDBProvider opens or initializes the central database at the binary location.
func NewCentralDBProvider() (*CentralDBProvider, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("could not determine executable path: %v", err)
	}
	dbPath := filepath.Join(filepath.Dir(execPath), "central.db")
	db, err := sql.Open("libsql", "file:"+dbPath)
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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		root_path TEXT UNIQUE,
		config TEXT
	)`)
	return err
}

// AddWorkspace adds a new workspace to the central database and returns its ID.
func (c *CentralDBProvider) AddWorkspace(rootPath, config string) (int, error) {
	result, err := c.db.Exec("INSERT INTO workspaces (root_path, config) VALUES (?, ?)", rootPath, config)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return int(id), err
}

func (c *CentralDBProvider) UpdateWorkspaceConfig(workspaceID int, config string) (bool, error) {
	_, err := c.db.Exec("UPDATE workspaces SET config = ? WHERE id = ?", config, workspaceID)
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetWorkspacePath retrieves the root path of a workspace by its ID.
func (c *CentralDBProvider) GetWorkspacePath(workspaceID int) (string, error) {
	var rootPath string
	err := c.db.QueryRow("SELECT root_path FROM workspaces WHERE id = ?", workspaceID).Scan(&rootPath)
	return rootPath, err
}

func (c *CentralDBProvider) GetWorkspaceID(rootPath string) (int, error) {
	var id int
	err := c.db.QueryRow("SELECT id FROM workspaces WHERE root_path = ?", rootPath).Scan(&id)
	return id, err
}

func (c *CentralDBProvider) GetWorkspaceConfig(workspaceID int) (string, error) {
	var config string
	err := c.db.QueryRow("SELECT config FROM workspaces WHERE id = ?", workspaceID).Scan(&config)
	return config, err
}

func (c *CentralDBProvider) SetWorkspaceConfig(workspaceID int, config string) error {
	_, err := c.db.Exec("UPDATE workspaces SET config = ? WHERE id = ?", config, workspaceID)
	return err
}

func (c *CentralDBProvider) DeleteWorkspace(workspaceID int) error {
	_, err := c.db.Exec("DELETE FROM workspaces WHERE id = ?", workspaceID)
	return err
}

func (c *CentralDBProvider) ListWorkspaces() ([]int, error) {
	rows, err := c.db.Query("SELECT id FROM workspaces")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaceIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		workspaceIDs = append(workspaceIDs, id)
	}
	return workspaceIDs, nil
}

func (c *CentralDBProvider) WorkspaceExists(workspaceID int) (bool, error) {
	var exists bool
	err := c.db.QueryRow("SELECT EXISTS(SELECT 1 FROM workspaces WHERE id = ?)", workspaceID).Scan(&exists)
	return exists, err
}

// Close closes the central database connection.
func (c *CentralDBProvider) Close() error {
	return c.db.Close()
}
