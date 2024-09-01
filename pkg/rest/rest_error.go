package rest

import (
	"fmt"
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
