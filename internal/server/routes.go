package server

import (
	"net/http"
	"path/filepath"

	paths "cms/internal"

	"github.com/julienschmidt/httprouter"
)

func (cms *CmsStruct) routes(ratelimitMode bool) http.Handler {
	router := httprouter.New()
	router.GET("/", cms.homeHandler)
	router.GET("/*name", cms.pageHandler)
	router.ServeFiles("/assets/style/*filepath", http.Dir(filepath.Join(paths.AssetsPath, "style")))
	router.ServeFiles("/assets/media/*filepath", http.Dir(filepath.Join(paths.AssetsPath, "media")))
	if !ratelimitMode {
		return cms.uncaughtErrorMiddleware(router)
	}
	return cms.uncaughtErrorMiddleware(cms.rateLimitMiddleware(router))
}
