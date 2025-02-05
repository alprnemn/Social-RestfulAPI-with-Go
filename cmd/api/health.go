package main

import (
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, req *http.Request) {

	err := WriteJSON(w, http.StatusOK,
		map[string]string{"message": "ok"})

	if err != nil {
		app.JsonResponse(w, http.StatusInternalServerError, "error occurred while writing json")
	}

}
