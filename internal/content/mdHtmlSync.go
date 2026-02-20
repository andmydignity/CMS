package content

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

var ErrDidntExist = errors.New("didn't exist in the first place")

func checksumCalculate(pathTo string) (string, error) {
	file, err := os.Open(pathTo)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	cheksum := hash.Sum(nil)
	return hex.EncodeToString(cheksum), nil
}

func appendChecksum(mdDir, mdFile, checksum string) error {
	db, err := sql.Open("sqlite", filepath.Join(mdDir, "checksum.db"))
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS checksums (filename TEXT PRIMARY KEY,hash TEXT NOT NULL)`)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `
		INSERT INTO checksums (filename, hash)
		VALUES (?, ?)
		ON CONFLICT(filename) DO UPDATE SET hash = excluded.hash`, mdFile, checksum)
	return err
}

func compareChecksum(mdDir, mdFile, checksum string) (bool, error) {
	db, err := sql.Open("sqlite", filepath.Join(mdDir, "checksum.db"))
	if err != nil {
		return false, err
	}
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS checksums (filename TEXT PRIMARY KEY,hash TEXT NOT NULL)`)
	if err != nil {
		return false, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	row := db.QueryRowContext(ctx, "SELECT 1 FROM checksums WHERE filename=? AND hash=?", mdFile, checksum)
	var exist int
	err = row.Scan(&exist)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, ErrDidntExist
		}
		return false, err
	}
	return true, nil
}

func deleteChecksum(mdDir, mdFile string) error {
	db, err := sql.Open("sqlite", filepath.Join(mdDir, "checksum.db"))
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS checksums (filename TEXT PRIMARY KEY,hash TEXT NOT NULL)`)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := db.ExecContext(ctx, "DELETE FROM checksums WHERE filename=?", mdFile)
	if err != nil {
		return err
	}
	if affected, err := res.RowsAffected(); affected != 1 || err != nil {
		if err != nil {
			return err
		}
		return ErrDidntExist
	}
	return nil
}

// TODO:Recursive support
// TODO: Purge non-existant files from DB
// TODO: Pass db between functions
func Sync(mdDir string) error {
	entries, err := os.ReadDir(mdDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		file := entry.Name()
		if file == "checksum.db" {
			continue
		}
		checksum, err := checksumCalculate(filepath.Join(mdDir, file))
		if err != nil {
			return err
		}
		equal, err := compareChecksum(mdDir, file, checksum)
		if errors.Is(err, ErrDidntExist) {
			err = appendChecksum(mdDir, file, checksum)
			if err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
		if !equal {
			err = deleteChecksum(mdDir, file)
			if err != nil {
				return err
			}
			err = appendChecksum(mdDir, file, checksum)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
