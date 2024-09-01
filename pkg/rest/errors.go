package rest

import (
	"errors"
	"net/http"
)

var (
	ErrJwtNotImplemented = NewInternalServerError(errors.New("jwt middleware not implemented"))
	ErrInsufficientRights = NewForbiddenError(nil, "Not enough permissions")
)

func NewBadRequestError(cause error, message string) ErrorResponse {
	return newError(cause, message, http.StatusBadRequest)
}

func NewUnauthorizedError(cause error, message string) ErrorResponse {
	return newError(cause, message, http.StatusUnauthorized)
}

func NewForbiddenError(cause error, message string) ErrorResponse {
	return newError(cause, message, http.StatusForbidden)
}

func NewNotFoundError(cause error, message string) ErrorResponse {
	return newError(cause, message, http.StatusNotFound)
}

func NewConflictError(cause error, message string) ErrorResponse {
	return newError(cause, message, http.StatusConflict)
}

func NewInternalServerError(cause error) ErrorResponse {
	return newError(cause, "Something went wrong", http.StatusInternalServerError)
}

func NewInvalidArgumentsError(cause error) ErrorResponse {
	return newError(cause, "Invalid arguments", http.StatusBadRequest)
}