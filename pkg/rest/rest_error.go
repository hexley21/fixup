package rest

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	BadRequest          = newError(errors.New("bad request"), http.StatusBadRequest)
	Unauthorized        = newError(errors.New("inauthorized"), http.StatusUnauthorized)
	Forbidden           = newError(errors.New("forbidden"), http.StatusForbidden)
	NotFound            = newError(errors.New("not found"), http.StatusNotFound)
	Conflict            = newError(errors.New("conflict"), http.StatusConflict)
	InternalServerError = newError(errors.New("internal server error"), http.StatusInternalServerError)
)

type ErrorResponse struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("status: %d - error: %s", e.Status, e.Message)
}

func newError(err error, status int) *ErrorResponse {
	return &ErrorResponse{
		Message: err.Error(),
		Status:  status,
	}
}

func writeError(w http.ResponseWriter, err *ErrorResponse) error {
	if e := write(w, err); e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return e
	}

	return err
}

func WriteCustomError(w http.ResponseWriter, err error, status int) error {
	return writeError(w, newError(err, status))
}

func WriteBadRequestError(w http.ResponseWriter) error {
	return writeError(w, BadRequest)
}

func WriteUnauthorizedError(w http.ResponseWriter) error {
	return writeError(w, Unauthorized)
}


func WriteForbiddenError(w http.ResponseWriter) error {
	return writeError(w, Forbidden)
}

func WriteNotFoundError(w http.ResponseWriter) error {
	return writeError(w, NotFound)
}

func WriteConflictError(w http.ResponseWriter) error {
	return writeError(w, Conflict)
}

func WriteInternalServerError(w http.ResponseWriter) error {
	return writeError(w, InternalServerError)
}
