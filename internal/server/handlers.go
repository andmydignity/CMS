package server

import (
	"fmt"
	"net/http"
)

func (cms *CmsStruct) homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Nothing here for now.")
}
