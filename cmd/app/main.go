package main

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	paths "cms/internal"
	"cms/internal/server"

	"gopkg.in/yaml.v2"
)

type config struct {
	Port          int      `yaml:"port"`
	CertPath      string   `yaml:"certPath"`
	KeyPath       string   `yaml:"keyPath"`
	CardsInHome   int      `yaml:"cardsInHome"`
	MdPath        string   `yaml:"mdPath"`
	SiteName      string   `yaml:"siteName"`
	LogoPath      string   `yaml:"logoPath"`
	FaviconPath   string   `yaml:"faviconPath"`
	HTTPSMode     bool     `yaml:"httpsMode"`
	Ratelimit     bool     `yaml:"ratelimit"`
	Replenishment float64  `yaml:"replenishment"`
	Burst         int      `yaml:"burst"`
	Domains       []string `yaml:"domains"`
}

func OpenDB(dbName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", filepath.Join(paths.DBPath, dbName))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS checksums (filename TEXT PRIMARY KEY CHECK (filename LIKE '%.md'),hash TEXT NOT NULL)`)
	if err != nil {
		return nil, err
	}
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS pages (url TEXT PRIMARY KEY,title TEXT NOT NULL, overview TEXT, overviewImg TEXT ,modifiedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP )`)
	return db, err
}

func main() {
	paths.SetPaths()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	db, err := OpenDB("database.db")
	if err != nil {
		logger.Error("Couldn't open DB! Error:" + err.Error())
		os.Exit(3)
	}
	file, err := os.Open(filepath.Join(paths.BinaryPath, "config.yaml"))
	if err != nil {
		logger.Error("Couldn't access config.yaml file!", "error", err.Error())
		os.Exit(2)
	}
	var cfg config
	decoder := yaml.NewDecoder(file)
	if err = decoder.Decode(&cfg); err != nil {
		logger.Error("Error while parsing YAML config!", "error", err.Error())
		os.Exit(3)
	}

	cmsConfig := server.CmsConfig{cfg.Port, cfg.CardsInHome, struct {
		Rps   float64
		Burst int
	}{cfg.Replenishment, cfg.Burst}, cfg.HTTPSMode, cfg.Ratelimit, cfg.CertPath, cfg.KeyPath, cfg.MdPath, cfg.SiteName, cfg.LogoPath, cfg.FaviconPath, cfg.Domains}
	cms := server.CmsStruct{logger, &cmsConfig, db}
	err = cms.Start()
	if !errors.Is(err, http.ErrServerClosed) && err != nil {
		cms.Logger.Error("Error while closing the server.", "error", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
