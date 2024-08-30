package rest

import (
	"encoding/json"
	"net/http"
)

type apiResponse[T any] struct {
	Data T `json:"data"`
}

func newApiResponse[T any](data T) *apiResponse[T] {
	return &apiResponse[T]{Data: data}
}

func WriteResponse[T any](w http.ResponseWriter, data T, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(newApiResponse(data)); err != nil {
		return json.NewEncoder(w).Encode(err)
	}

	return nil
}

func WriteOkResponse[T any](w http.ResponseWriter, data T) error {
	return WriteResponse(w, data, http.StatusOK)
}
