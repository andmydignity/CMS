package server

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	paths "cms/internal"
	"cms/internal/filesync"
	"cms/internal/render"

	"github.com/julienschmidt/httprouter"
)

func (cms *CmsStruct) homeHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	if len(render.HomePageCache) != 0 {
		w.Write(render.HomePageCache)
	} else {
		home, err := os.ReadFile(filepath.Join(paths.AssetsPath, "homePage", "home.html"))
		if err != nil {
			cms.internalError(w, err)
			return
		}
		w.Write(home)
	}
}

func (cms *CmsStruct) pageHandler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	name := strings.TrimPrefix(ps.ByName("name"), "/")
	if name == "" {
		cms.badRequest(w, "Page name empty.")
		return
	}
	path := filepath.Join(paths.AssetsPath, "pages", name+".html")
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
	type dataStruct struct {
		Title       string
		SiteName    string
		Year        int
		FaviconPath string
		LogoPath    string
		Pages       []render.PageInfo
	}

	pages, err := render.GetPages(32, cms.DB)
	if err != nil {
		cms.internalError(w, err)
		return
	}
	ds := dataStruct{cms.Config.SiteName, cms.Config.SiteName, time.Now().Year(), cms.Config.FaviconPath, cms.Config.LogoPath, pages}
	ts, err := template.ParseFiles(filepath.Join(paths.AssetsPath, "searchPage", "base.tmpl"), filepath.Join(paths.AssetsPath, "searchPage", "navbar.tmpl"), filepath.Join(paths.AssetsPath, "searchPage", "footer.tmpl"))
	if err != nil {
		cms.internalError(w, err)
		return
	}

	ts.ExecuteTemplate(w, "base.tmpl", ds)
}
