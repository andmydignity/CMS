package filesync

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"
)

func checksumCalculate(pathTo string) (string, error) {
	file, err := os.Open(pathTo)
	// Windows and its file locks, man...
	if err != nil {
		if runtime.GOOS == "windows" {
			for i := 0; i < 5; i++ {
				time.Sleep(50 * time.Millisecond)
				file, err = os.Open(pathTo)
				if err == nil {
					break
				}
			}
			if err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	cheksum := hash.Sum(nil)
	return hex.EncodeToString(cheksum), nil
}

func appendChecksum(db *sql.DB, mdFile, checksum string) error {
	if mdFile == "" {
		return fmt.Errorf("empty mdFile passed")
	}
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `
		INSERT INTO checksums (filename, hash)
		VALUES (?, ?)
		ON CONFLICT(filename) DO UPDATE SET hash = excluded.hash`, mdFile, checksum)
	return err
}

func isChecksumSame(db *sql.DB, mdFile, checksum string) (bool, error) {
	if mdFile == "" {
		return false, fmt.Errorf("empty mdFile passed")
	}
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	row := db.QueryRowContext(ctx, "SELECT 1 FROM checksums WHERE filename=? AND hash=?", mdFile, checksum)
	var exist int
	err = row.Scan(&exist)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func deleteChecksum(db *sql.DB, mdFile string) error {
	if mdFile == "" {
		return fmt.Errorf("empty mdFile passed")
	}
	var err error
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
