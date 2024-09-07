package rest

import (
	"errors"
	"net/http"
)

var (
	// 400
	ErrInvalidArguments = NewInvalidArgumentsError(nil)
	ErrInvalidFileType = NewBadRequestError(nil, "Invalid file type")
	ErrTooManyFiles = NewBadRequestError(nil, "Too many files")
	ErrNoFile = NewBadRequestError(nil, "No file provided")
	// 403
	ErrInsufficientRights = NewForbiddenError(nil, "Not enough permissions")
	// 500
	ErrJwtNotImplemented = NewInternalServerError(errors.New("jwt middleware not implemented"))
)

func NewInvalidArgumentsError(cause error) ErrorResponse {
	return newError(cause, "Invalid arguments", http.StatusBadRequest)
}

func NewReadFileError(cause error) ErrorResponse {
	return newError(cause, "Failed read file", http.StatusInternalServerError)
}
