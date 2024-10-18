package rest

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
)

const (
	MsgInvalidArguments = "Invalid arguments"
	MsgInvalidId        = "Invalid ID"

	MsgFileReadError = "Failed read file"

	MsgInternalServerError = "Something went wrong"

	MsgTooManyFiles   = "Too many files"
	MsgNotEnoughFiles = "Not enough files"
)

var (
	ErrNoFile = errors.New("no file provided")
)

type ErrorResponse struct {
	Cause   error  `json:"-"`
	Message string `json:"message"`
	Status  int    `json:"-"`
}

func (e *ErrorResponse) Error() string {
	var sb strings.Builder

	sb.WriteString("status: ")
	sb.WriteString(strconv.Itoa(e.Status))
	sb.WriteString(" - message: ")
	sb.WriteString(e.Message)

	if e.Cause != nil {
		sb.WriteString(" - cause: ")
		sb.WriteString(e.Cause.Error())
	}

	return sb.String()
}

func newError(cause error, status int, message string) *ErrorResponse {
	return &ErrorResponse{
		Cause:   cause,
		Message: message,
		Status:  status,
	}
}

func NewBadRequestError(cause error) *ErrorResponse {
	return newError(cause, http.StatusBadRequest, cause.Error())
}

func NewUnauthorizedError(cause error) *ErrorResponse {
	return newError(cause, http.StatusUnauthorized, cause.Error())
}

func NewForbiddenError(cause error) *ErrorResponse {
	return newError(cause, http.StatusForbidden, cause.Error())
}

func NewNotFoundError(cause error) *ErrorResponse {
	return newError(cause, http.StatusNotFound, cause.Error())
}

func NewNotFoundMessageError(cause error, message string) *ErrorResponse {
	return newError(cause, http.StatusNotFound, message)
}

func NewConflictError(cause error) *ErrorResponse {
	return newError(cause, http.StatusConflict, cause.Error())
}

func NewInternalServerError(cause error) *ErrorResponse {
	return newError(cause, http.StatusInternalServerError, MsgInternalServerError)
}

func NewInvalidArgumentsError(cause error) *ErrorResponse {
	return newError(cause, http.StatusBadRequest, MsgInvalidArguments)
}

func NewReadFileError(cause error) *ErrorResponse {
	return newError(cause, http.StatusBadRequest, MsgFileReadError)
}
