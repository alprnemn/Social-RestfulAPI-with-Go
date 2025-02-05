package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func ParseJSON(w http.ResponseWriter, req *http.Request, data any) error {

	// determines max bytes size
	maxBytes := 1_048_578
	req.Body = http.MaxBytesReader(w, req.Body, int64(maxBytes))

	if req.Body == nil {
		return fmt.Errorf("missing request body")
	}

	decoder := json.NewDecoder(req.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	type envelope struct {
		Error string `json:"error"`
	}
	WriteJSON(w, status, &envelope{
		Error: message,
	})
}

func (app *application) JsonResponse(w http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}
	return WriteJSON(w, status, &envelope{Data: data})
}

func WriteJSON(w http.ResponseWriter, status int, data any) error {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}
