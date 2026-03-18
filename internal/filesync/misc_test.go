package filesync

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func mockDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal("Failed to create mockup db.", "error", err.Error())
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS checksums (filename TEXT PRIMARY KEY CHECK (filename LIKE '%.md'),hash TEXT NOT NULL)`)
	if err != nil {
		t.Fatal("Failed to create checksum table", "error", err.Error())
		return nil
	}
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS pages (url TEXT PRIMARY KEY,title TEXT NOT NULL, overview TEXT, overviewImg TEXT ,modifiedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP )`)
	if err != nil {
		t.Fatal("Failed to create pages table.", "error", err.Error())
	}
	t.Cleanup(func() {
		db.Close()
	})
	return db
}

func TestPurgeOrphanDbEntries(t *testing.T) {
	tests := []struct {
		name          string
		existingFiles []string
		filesInDB     []string
		wantErr       bool
	}{
		{"empty dir shall pass", []string{}, []string{}, false},
		{"2 existing 4 in db", []string{"a.md", "b.md"}, []string{"a.md", "b.md", "c.md", "d.md"}, false},
		{"2 md 1 png |DB 4md ", []string{"loa.md", "bob.md", "semih.png"}, []string{"loa.md", "bob.md", "nah.md", "lol.md"}, false},
	}
	db := mockDB(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mdDir := t.TempDir()
			mdFiles := []string{}
			for i, filename := range tt.filesInDB {
				filename = filepath.Join(mdDir, filename)
				tt.filesInDB[i] = filename
				err := appendChecksum(db, filename, "")
				if err != nil {
					t.Error("Append checksum failed. " + err.Error())
				}
			}
			for i, file := range tt.existingFiles {
				file = filepath.Join(mdDir, file)
				if _, has := strings.CutSuffix(file, ".md"); has {
					mdFiles = append(mdFiles, file)
				}
				tt.existingFiles[i] = file
				err := os.WriteFile(file, nil, 0o655)
				if err != nil {
					t.Error("WriteFile failed. " + err.Error())
				}
			}
			gotErr := false
			errText := ""
			err := purgeOrphans(db, tt.existingFiles, mdDir)
			if err != nil {
				gotErr = true
				errText = err.Error()
			}
			if gotErr != tt.wantErr {
				t.Errorf("gotErr: %v wantErr: %v mismatch. "+errText, gotErr, tt.wantErr)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			res, err := db.QueryContext(ctx, "SELECT filename FROM checksums")
			if err != nil {
				t.Error("Couldn't get query results! " + errText)
			}
			mdDirFilesInDB := []string{}
			for res.Next() {
				filename := ""
				res.Scan(&filename)
				mdDirFilesInDB = append(mdDirFilesInDB, filename)
			}
			if len(mdDirFilesInDB) != len(mdFiles) {
				t.Errorf(`In DB: %v 
			Existing Files: %v`, mdDirFilesInDB, tt.existingFiles)
			}
			// n2 look up is bad but we are gonna test with a maximum of 5 anyways.
			for _, inDB := range mdDirFilesInDB {
				if !slices.Contains(mdFiles, inDB) {
					fmt.Printf(`In DB: %v 
			Existing Files: %v \n`, mdDirFilesInDB, tt.existingFiles)
					t.Errorf("DB entry %v is an orphan even after the orphans were purged.", inDB)
				}
			}
		})
	}
}
