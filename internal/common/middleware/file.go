package middleware

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/hexley21/fixup/pkg/http/rest"
)

// NewAllowFilesAmount creates a middleware that validates the number of files uploaded in a multipart form.
// It checks if the number of files associated with the given key matches the specified amount.
// If there are too many, too few, or no files, it writes an appropriate error response.
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
				f.writer.WriteError(w, rest.ErrTooManyFiles)
				return
			}

			if len(files) == 0 {
				f.writer.WriteError(w, rest.ErrNoFile)
				return
			}

			if len(files) < amount {
				f.writer.WriteError(w, rest.ErrNotEnoughFiles)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// NewAllowContentType creates a middleware that validates the content type of files uploaded in a multipart form.
// It checks if the content type of each file associated with the given key matches one of the allowed types.
// If any file has an invalid content type, it writes an error response.
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
				f.writer.WriteError(w, rest.NewBadRequestError(fmt.Errorf("invalid file type: %s, for file: %s", contentType, file.Filename)))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
