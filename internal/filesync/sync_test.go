package filesync

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	paths "cms/internal"
	"cms/internal/render"
)

// What a pain in the ass to test.
func TestFirstSync(t *testing.T) {
	tests := []struct {
		name        string
		wantErr     bool
		files       []string
		wantedFiles []string
	}{
		{"2 md", false, []string{"lol.md", filepath.Join("adir", "funny.md")}, []string{"lol", filepath.Join("adir", "funny")}},
		{"2 md 1 png", false, []string{"haha.md", filepath.Join("dir", "lol.png"), filepath.Join("dir", "ımkındadümbaşş.md")}, []string{"haha", filepath.Join("dir", "ımkındadümbaşş")}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := mockDB(t)
			tempDir := t.TempDir()
			mdDir := filepath.Join(tempDir, "mdDir")
			pageDir := filepath.Join(tempDir, "assets", "pages")
			err := os.MkdirAll(mdDir, 0o777)
			if err != nil {
				t.Fatal("Couldn't create mdDir!")
			}
			err = os.MkdirAll(pageDir, 0o777)
			if err != nil {
				t.Fatal("Couldn't create pageDir!")
			}
			paths.AssetsPath = filepath.Join(tempDir, "assets")
			os.MkdirAll(filepath.Join(paths.AssetsPath, "templates"), 0o777)
			os.Create(filepath.Join(paths.AssetsPath, "templates", "base.tmpl"))
			os.MkdirAll(filepath.Join(paths.AssetsPath, "homePage", "templates"), 0o777)
			os.Create(filepath.Join(paths.AssetsPath, "homePage", "templates", "base.tmpl"))

			for _, file := range test.files {
				os.MkdirAll(filepath.Dir(filepath.Join(mdDir, file)), 0o777)
				os.Create(filepath.Join(mdDir, file))
			}
			rdrconf := render.RenderConfig{"Test", "", ""}
			err = FirstSync(mdDir, db, &rdrconf)

			if (err != nil && test.wantErr == false) || (err == nil && test.wantErr == true) {
				errText := ""
				if err != nil {
					errText = err.Error()
				}
				t.Fatalf("Error wantErr mismatch. wantErr %v. Error msg: %v", test.wantErr, errText)
			}
			foundMd := []string{}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			res, err := db.QueryContext(ctx, "SELECT filename FROM checksums")
			if err != nil {
				t.Fatal("Query failed: " + err.Error())
			}
			for res.Next() {
				filename := ""
				res.Scan(&filename)
				foundMd = append(foundMd, filename)
			}
			if len(foundMd) != len(test.wantedFiles) {
				t.Fatalf("Lengths don't match between foundMd and wantedFiles. foundMd: %v wantedFiles: %v", foundMd, test.wantedFiles)
			}
			pagesGenerated := []string{}
			filepath.WalkDir(filepath.Join(paths.AssetsPath, "pages"), func(path string, d fs.DirEntry, err error) error {
				if stat, _ := os.Stat(path); stat.IsDir() {
					return nil
				}
				pagesGenerated = append(pagesGenerated, path)
				return nil
			})
			if len(test.wantedFiles) != len(pagesGenerated) {
				t.Fatalf("wantedFiles and pagesGenerated have different lengths. wantedFiles: %v pagesGenerated: %v", test.wantedFiles, pagesGenerated)
			}
			for _, file := range test.wantedFiles {
				if !slices.Contains(foundMd, filepath.Join(mdDir, file)+".md") {
					t.Fatalf("%v not found in foundMd. foundMd: %v", filepath.Join(mdDir, file)+".md", foundMd)
				}
				if !slices.Contains(pagesGenerated, filepath.Join(paths.AssetsPath, "pages", file)+".html") {
					t.Fatalf("%v not found in pagesGenerated. pagesGenerated: %v", filepath.Join(paths.AssetsPath, "pages", file)+".html", pagesGenerated)
				}
			}
		})
	}
}
