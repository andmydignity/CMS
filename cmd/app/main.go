package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"cms/internal/server"
)

func main() {
	cmsConfig := server.CmsConfig{8080, struct {
		Rps   float64
		Burst int
	}{10, 20}, true, "", "", "mdFiles", 5}
	cms := server.CmsStruct{slog.New(slog.NewTextHandler(os.Stdout, nil)), &cmsConfig}
	err := cms.Start()
	if !errors.Is(err, http.ErrServerClosed) && err != nil {
		cms.Logger.Error("Error while closing the server.", "error", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
