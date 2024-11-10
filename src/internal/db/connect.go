package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
)

func ConnectToDB(path string) (*sql.DB, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		slog.Info("Database not found, creating a new one")
		file, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("could not create db at path %s): %v", path, err)
		}
		file.Close()
	}

	db, err := sql.Open("libsql", "file:"+path+"?_foreign_keys=1")
	if err != nil {
		return nil, err
	}

	return db, nil
}
