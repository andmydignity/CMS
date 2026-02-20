package server

import "net/http"

func (cms *CmsStruct) internalError(w http.ResponseWriter, err error) {
	// recoverPanic passes nil since recover() returns an interface
	cms.Logger.Error(err.Error())
	w.WriteHeader(http.StatusInternalServerError)

	// TODO:Implement internalError()
}

func (cms *CmsStruct) tooManyRequests(w http.ResponseWriter) {
	w.WriteHeader(http.StatusTooManyRequests)
	// TODO:IMplement
}
