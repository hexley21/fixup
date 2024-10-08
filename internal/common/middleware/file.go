package middleware

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/hexley21/fixup/pkg/http/rest"
)

var (
	ErrTooManyFiles   = rest.NewBadRequestError(nil, rest.MsgTooManyFiles)
	ErrNotEnoughFiles = rest.NewBadRequestError(nil, rest.MsgNotEnoughFiles)
	ErrNoFile         = rest.NewBadRequestError(nil, rest.MsgNoFile)
)

func (f *Middleware) NewAllowFilesAmount(size int64, key string, amount int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			form, err := f.binder.BindMultipartForm(r, size)
			if err != nil {
				f.writer.WriteError(w, rest.NewReadFileError(err))
				return
			}

			files := form.File[key]

			if len(files) > amount {
				f.writer.WriteError(w, ErrTooManyFiles)
				return
			}

			if len(files) == 0 {
				f.writer.WriteError(w, ErrNoFile)
				return
			}

			if len(files) < amount {
				f.writer.WriteError(w, ErrNotEnoughFiles)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (f *Middleware) NewAllowContentType(size int64, key string, types ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			form, err := f.binder.BindMultipartForm(r, size)
			if err != nil {
				f.writer.WriteError(w, err)
				return
			}

			for _, file := range form.File[key] {
				contentType := file.Header.Get("Content-Type")
				if slices.Contains(types, contentType) {
					continue
				}
				f.writer.WriteError(w, rest.NewBadRequestError(nil, fmt.Sprintf("Invalid file type: %s, for file: %s", contentType, file.Filename)))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
