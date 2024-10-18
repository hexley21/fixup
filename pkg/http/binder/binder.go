package binder

import (
	"errors"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/hexley21/fixup/pkg/http/rest"
)


var (
	ErrUnsupportedMediaType = rest.NewBadRequestError(errors.New("unsupported media type"))
	ErrEmptyBody = rest.NewBadRequestError(errors.New("empty body"))
)

type FullBinder interface {
	JSONBinder
	FormBinder
}

type JSONBinder interface {
	BindJSON(r *http.Request, i any) *rest.ErrorResponse
}

type FormBinder interface {
	BindForm(r *http.Request) (url.Values, *rest.ErrorResponse)
	BindMultipartForm(r *http.Request, maxSize int64) (*multipart.Form, *rest.ErrorResponse)
}
