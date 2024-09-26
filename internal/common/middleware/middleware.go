package middleware

import (
	"github.com/hexley21/fixup/pkg/http/binder"
	"github.com/hexley21/fixup/pkg/http/writer"
)

const(
	MsgInsufficientRights = "Insufficient rights"
	MsgUserIsVerified     = "User has to be not-verified"
	MsgUserIsNotVerified  = "User is not verified"

	MsgNoFile         = "No file provided"
	MsgTooManyFiles   = "Too many files"
	MsgNotEnoughFiles = "Not enough files"

	MsgMissingAuthorizationHeader = "Authorization header is missing"
	MsgMissingBearerToken         = "Bearer token is missing"
)

type MiddlewareFactory struct {
	binder binder.FullBinder
	writer writer.HTTPErrorWriter
}

func NewMiddlewareFactory(binder binder.FullBinder, writer writer.HTTPErrorWriter) *MiddlewareFactory{
	return &MiddlewareFactory{
		binder, writer,
	}
}
