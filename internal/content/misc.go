// Package content
package content

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"time"
)

func OpenDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite", filepath.Join(dbLocation, "checksum.db"))
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS checksums (filename TEXT PRIMARY KEY,hash TEXT NOT NULL)`)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func deleteHTML(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}

func purgeNonExistent(db *sql.DB, fileNames []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Build O(1) lookup set
	set := make(map[string]struct{}, len(fileNames))
	for _, f := range fileNames {
		set[f] = struct{}{}
	}

	rows, err := db.QueryContext(ctx, "SELECT filename FROM checksums")
	if err != nil {
		return err
	}
	defer rows.Close()

	var file string
	for rows.Next() {
		if err := rows.Scan(&file); err != nil {
			return err
		}

		// O(1) lookup instead of slices.Contains (O(n))
		if _, ok := set[file]; !ok {
			if err := deleteChecksum(db, file); err != nil {
				return err
			}

			// TODO: Implement HTML file deletion.
		}
	}

	return rows.Err()
}
