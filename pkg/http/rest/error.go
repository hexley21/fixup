package rest

import (
	"fmt"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

const (
	MsgInternalServerError = "Something went wrong"

	MsgInvalidArguments = "Invalid arguments"
	MsgInvalidId        = "Invalid ID"

	MsgFileReadError = "Failed read file"
)

var (
	ErrInsufficientRights = NewForbiddenError(errors.New("insufficient rights"))

	ErrNoFile         = NewBadRequestError(errors.New("no file provided"))
	ErrNotEnoughFiles = NewBadRequestError(errors.New("not enough files"))
	ErrTooManyFiles   = NewBadRequestError(errors.New("too many files"))
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

func NewInternalServerErrorf(format string, args ...any) *ErrorResponse {
	return newError(fmt.Errorf(format, args...), http.StatusInternalServerError, MsgInternalServerError)
}

// App oriented errors

func NewInvalidArgumentsError(cause error) *ErrorResponse {
	return newError(cause, http.StatusBadRequest, MsgInvalidArguments)
}

func NewInvalidIdError(cause error) *ErrorResponse {
	return newError(cause, http.StatusBadRequest, MsgInvalidId)
}

func NewReadFileError(cause error) *ErrorResponse {
	return newError(cause, http.StatusBadRequest, MsgFileReadError)
}
