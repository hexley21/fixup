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

func (e ErrorResponse) Error() string {
	return fmt.Sprintf("status: %d - error: %s", e.Status, e.Message)
}

func newError(err error, status int) ErrorResponse {
	return ErrorResponse{
		Message: err.Error(),
		Status:  status,
	}
}
