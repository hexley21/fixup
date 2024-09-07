package rest

import (
	"fmt"
	"net/http"
)

type ErrorResponse struct {
	Cause   error  `json:"-"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e ErrorResponse) Error() string {
	if e.Cause == nil {
		return fmt.Sprintf("status: %d - message: %s", e.Status, e.Message)
	}

	return fmt.Sprintf("status: %d - message: %s - cause: %s", e.Status, e.Message, e.Cause.Error())
}

func newError(cause error, message string, status int) ErrorResponse {
	return ErrorResponse{
		Cause:   cause,
		Message: message,
		Status:  status,
	}
}

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
