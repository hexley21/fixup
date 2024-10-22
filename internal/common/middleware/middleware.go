package middleware

import (
	"errors"

	"github.com/hexley21/fixup/pkg/http/binder"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/http/writer"
)

var (
	ErrMissingAuthorizationHeader = rest.NewUnauthorizedError(errors.New("authorization header is missing"))
	ErrMissingBearerToken         = rest.NewUnauthorizedError(errors.New("bearer token is missing"))
)

type Middleware struct {
	binder binder.FullBinder
	writer writer.HTTPErrorWriter
}

func NewMiddleware(binder binder.FullBinder, writer writer.HTTPErrorWriter) *Middleware {
	return &Middleware{
		binder: binder,
		writer: writer,
	}
}
