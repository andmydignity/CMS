package content

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"cms/internal/render"

	"github.com/fsnotify/fsnotify"
	"github.com/sgtdi/fswatcher"
	_ "modernc.org/sqlite"
)

const dbLocation = "databases"

var ErrDidntExist = errors.New("didn't exist in the first place")

func FirstSync(mdDir string, db *sql.DB) error {
	var entries []string
	err := filepath.WalkDir(mdDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(mdDir, path)
		if err != nil {
			return err
		}
		entries = append(entries, rel)
		return nil
	})
	if err != nil {
		return err
	}
	fullpaths := []string{}
	for _, file := range entries {
		if file == "." {
			continue
		}
		fullpath := filepath.Join(mdDir, file)
		checksum, err := checksumCalculate(fullpath)
		if err != nil {
			return err
		}
		err = appendChecksum(db, fullpath, checksum)
		if err != nil {
			return err
		}

		extensionSanitized, _ := strings.CutSuffix(file, ".md")
		err = render.SaveMdtoHTML(fullpath, filepath.Join("assets", "pages", extensionSanitized))
		if err != nil {
			return err
		}
		fullpaths = append(fullpaths, fullpath)
	}
	err = purgeNonExistent(db, fullpaths)
	if err != nil {
		return err
	}
	return nil
}

func Sync(ctx context.Context, db *sql.DB, mdDir string, logger *slog.Logger) error {
	watcher, err := fswatcher.New(fswatcher.WithPath(mdDir), fswatcher.WithSeverity(fswatcher.SeverityInfo))
	if err != nil {
		return err
	}
	defer watcher.Close()
	go watcher.Watch(ctx)
	for event := range watcher.Events() {
		for _, t := range event.Types {
			fswatcher.Debouncer
		}
	}
	for {
		select {
		case <-ctx.Done():
			logger.Info("Sync shutting down")
			return nil

		case err := <-watcher.Errors:
			logger.Error("Watcher error", "error", err)

		case event := <-watcher.Events:

			if event.Has(fsnotify.Create) || event.Has(fsnotify.Write) {

				st, err := os.Stat(event.Name)
				if err == nil && st.IsDir() {
					// Add new directories dynamically
					if err := watcher.Add(event.Name); err != nil {
						logger.Error("Failed to add new dir", "error", err)
					}
					continue
				}

				checksum, err := checksumCalculate(event.Name)
				if err != nil {
					logger.Error("Checksum error", "error", err)
					continue
				}

				if err := appendChecksum(db, event.Name, checksum); err != nil {
					logger.Error("DB error", "error", err)
					continue
				}

				suffixCut, _ := strings.CutSuffix(event.Name, ".md")
				extensionSanitized, _ := strings.CutPrefix(suffixCut, mdDir)

				if err := render.SaveMdtoHTML(
					event.Name,
					filepath.Join("assets", "pages", extensionSanitized),
				); err != nil {
					logger.Error("Render error", "error", err)
					continue
				}

			} else if event.Has(fsnotify.Remove) || event.Has(fsnotify.Rename) {
				if err := deleteChecksum(db, event.Name); err != nil {
					logger.Error("Delete checksum error", "error", err)
				}
				// This is quite clunky, should ideally be changed but works for now.
				suffixCut, _ := strings.CutSuffix(event.Name, ".md")
				extensionSanitized, _ := strings.CutPrefix(suffixCut, mdDir)
				if err = deleteHTML(filepath.Join("assets", "pages", fmt.Sprintf("%v.html", extensionSanitized))); err != nil {
					logger.Error("Delete HTML error", "error", err.Error())
				}
			}
		}
	}
}
