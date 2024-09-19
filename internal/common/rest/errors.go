package rest

import (
	"fmt"
	"net/http"
)

const (
	MsgInvalidArguments = "Invalid arguments"

	MsgUserNotFound      = "User not found"
	MsgUserAlreadyExists = "User already exists"

	MsgIncorrectPassword    = "Password is incorrect"
	MsgIncorrectEmailOrPass = "Email or Password is incorrect"

	MsgFileReadError = "Failed read file"
	MsgNoFile = "No file provided"
	MsgTooManyFiles = "Too many files"

	MsgInvalidToken = "Invalid token"
	MsgMissingAuthorizationHeader = "Authorization header is missing"
	MsgMissingBearerToken = "Bearer token is missing"

	MsgInternalServerError = "Something went wrong"
)

func NewInvalidArgumentsError(cause error) *ErrorResponse {
	return newError(cause, MsgInvalidArguments, http.StatusBadRequest)
}

func NewBindError(cause error) *ErrorResponse {
	return newError(fmt.Errorf("bind failed: %w", cause), MsgInvalidArguments, http.StatusBadRequest)
}

func NewValidationError(cause error) *ErrorResponse {
	return newError(fmt.Errorf("validation failed: %w", cause), MsgInvalidArguments, http.StatusBadRequest)
}

func NewReadFileError(cause error) *ErrorResponse {
	return newError(cause, MsgFileReadError, http.StatusInternalServerError)
}
