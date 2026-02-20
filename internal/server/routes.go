package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (cms *CmsStruct) routes() http.Handler {
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/", cms.homeHandler)
	return cms.uncaughtErrorMiddleware(cms.rateLimitMiddleware(router))
}
