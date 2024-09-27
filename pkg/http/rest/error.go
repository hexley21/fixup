package rest

import (
	"net/http"
	"strconv"
	"strings"
)

const (
	MsgInvalidArguments = "Invalid arguments"

	MsgFileReadError  = "Failed read file"

	MsgInternalServerError = "Something went wrong"

	MsgUnsupportedMedia = "Unsupported media type"
	MsgEmptyBody        = "Empty body"
)

type ErrorResponse struct {
	Cause   error  `json:"-"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func (e *ErrorResponse) Error() string {
	sb := strings.Builder{}

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

func NewBadRequestError(cause error, message string) *ErrorResponse {
	return newError(cause, http.StatusBadRequest, message)
}

func NewUnauthorizedError(cause error, message string) *ErrorResponse {
	return newError(cause, http.StatusUnauthorized, message)
}

func NewForbiddenError(cause error, message string) *ErrorResponse {
	return newError(cause, http.StatusForbidden, message)
}

func NewNotFoundError(cause error, message string) *ErrorResponse {
	return newError(cause, http.StatusNotFound, message)
}

func NewConflictError(cause error, message string) *ErrorResponse {
	return newError(cause, http.StatusConflict, message)
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
