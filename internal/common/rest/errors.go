package rest

import (
	"net/http"
)

func NewInvalidArgumentsError(cause error) ErrorResponse {
	return newError(cause, "Invalid arguments", http.StatusBadRequest)
}

func NewReadFileError(cause error) ErrorResponse {
	return newError(cause, "Failed read file", http.StatusInternalServerError)
}
