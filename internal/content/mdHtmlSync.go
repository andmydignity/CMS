package content

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cms/internal/render"

	_ "modernc.org/sqlite"
)

const dbLocation = "databases"

var ErrDidntExist = errors.New("didn't exist in the first place")

func openDB() (*sql.DB, error) {
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

// TODO:Add actually creating the files.
func Sync(mdDir string) error {
	db, err := openDB()
	if err != nil {
		return err
	}
	defer db.Close()
	var entries []os.DirEntry
	err = filepath.WalkDir(mdDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		entries = append(entries, d)
		return nil
	})
	files := []string{}
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Shouldn't be possible but still check.
		if entry.IsDir() {
			continue
		}
		file := entry.Name()

		checksum, err := checksumCalculate(filepath.Join(mdDir, file))
		if err != nil {
			return err
		}
		err = appendChecksum(db, file, checksum)
		if err != nil {
			return err
		}
		files = append(files, file)
		fullpath := filepath.Join(mdDir, file)
		extensionSanitized, _ := strings.CutSuffix(file, ".md")
		err = render.SaveMdtoHTML(fullpath, filepath.Join("assets", "pages", extensionSanitized))
		if err != nil {
			return err
		}
	}

	err = purgeNonExistent(db, files)
	if err != nil {
		return err
	}
	return nil
}
