// Package globals sets globals (yeah gotta find a better description)
package globals

import (
	"os"
	"path/filepath"
)

var (
	AssetsPath, CertsPath, DBPath, BinaryPath string
	HomePageCahce, SearchPageCache            []byte
)

func getBinaryDir() (string, error) {
	exPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	resolvedPath, err := filepath.EvalSymlinks(exPath)
	if err != nil {
		return "", err
	}
	return filepath.Dir(resolvedPath), nil
}

func SetPaths() error {
	binDir, err := getBinaryDir()
	if err != nil {
		return err
	}
	AssetsPath = filepath.Join(binDir, "assets")
	CertsPath = filepath.Join(binDir, "certs")
	DBPath = filepath.Join(binDir, "databases")
	BinaryPath = binDir
	return nil
}
