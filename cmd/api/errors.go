package main

import (
	"log"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, req *http.Request, err error) {

	log.Printf("internal server error: %s path: %s error: %s", req.Method, req.URL.Path, err.Error())
	WriteError(w, http.StatusInternalServerError, "the server encountered a problem")

}

func (app *application) forbiddenResponse(w http.ResponseWriter, req *http.Request) {

	log.Printf("forbidden error: %s path: %s error: %s", req.Method, req.URL.Path, "error")
	WriteError(w, http.StatusForbidden, "forbidden")

}

func (app *application) databaseStoreError(w http.ResponseWriter, req *http.Request, err error) {

	log.Printf("db error: %s path: %s error: %s", req.Method, req.URL.Path, err.Error())
	WriteError(w, http.StatusInternalServerError, err.Error())

}

func (app *application) badRequestResponse(w http.ResponseWriter, req *http.Request, err error) {

	log.Printf("bad request error: %s path: %s error: %s", req.Method, req.URL.Path, err)
	WriteError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundError(w http.ResponseWriter, req *http.Request, err error) {
	log.Printf("not found error: %s path: %s error: %s", req.Method, req.URL.Path, err)
	WriteError(w, http.StatusNotFound, "resource not found")
}

func (app *application) conflictError(w http.ResponseWriter, req *http.Request, err error) {
	log.Printf("conflict error: %s path: %s error: %s", req.Method, req.URL.Path, err)
	WriteError(w, http.StatusConflict, err.Error())
}

func (app *application) unAuthorizedError(w http.ResponseWriter, req *http.Request, err error) {
	log.Printf("unauthorized error: %s path: %s error: %s", req.Method, req.URL.Path, err)
	WriteError(w, http.StatusUnauthorized, "unauthorized")
}
