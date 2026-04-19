package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"cms/internal/filesync"
	"cms/internal/globals"

	"github.com/julienschmidt/httprouter"
)

func (cms *CmsStruct) homeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Encoding", "gzip")
	if len(globals.HomePageCahce) != 0 {
		w.Write(globals.HomePageCahce)
	} else {
		home, err := os.ReadFile(filepath.Join(globals.AssetsPath, "homePage", "home.html"))
		if err != nil {
			cms.internalError(w, err)
			return
		}
		w.Write(home)
	}
}

func (cms *CmsStruct) pageHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Encoding", "gzip")
	name := strings.TrimPrefix(ps.ByName("name"), "/")
	if name == "" {
		cms.badRequest(w, "Page name empty.")
		return
	}
	path := filepath.Join(globals.AssetsPath, "pages", name+".html")
	if page := filesync.FromCache(path); page != nil {
		w.Write(page)
		return
	}
	if _, err := os.Stat(path); err != nil {
		cms.notFound(w)
		return
	}
	data, err := filesync.AppendToCache(path)
	if err != nil {
		cms.internalError(w, err)
		return
	}
	w.Write(data)
}

func (cms *CmsStruct) searchPageHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	w.Header().Set("Content-Encoding", "gzip")
	if len(globals.SearchPageCache) != 0 {
		w.Write(globals.SearchPageCache)
	} else {
		search, err := os.ReadFile(filepath.Join(globals.AssetsPath, "searchPage", "search.html"))
		if err != nil {
			cms.internalError(w, err)
			return
		}
		w.Write(search)

	}
}
