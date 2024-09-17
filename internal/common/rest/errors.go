package rest

import (
	"net/http"
)

const (
	MsgUserNotFound      = "User not found"
	MsgUserAlreadyExists = "User already exists"

	MsgIncorrectPassword    = "Password is incorrect"
	MsgIncorrectEmailOrPass = "Email or Password is incorrect"

	MsgInvalidToken = "Invalid token"
	MsgMissingAuthorizationHeader = "Authorization header is missing"
	MsgMissingBearerToken = "Bearer token is missing"

	MsgInternalServerError = "Something went wrong"
)

func NewInvalidArgumentsError(cause error) *ErrorResponse {
	return newError(cause, "Invalid arguments", http.StatusBadRequest)
}

func NewReadFileError(cause error) *ErrorResponse {
	return newError(cause, "Failed read file", http.StatusInternalServerError)
}
