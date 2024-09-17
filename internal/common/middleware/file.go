package middleware

import (
	"fmt"
	"slices"

	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/labstack/echo/v4"
)

var (
	errTooManyFiles = rest.NewBadRequestError(nil, rest.MsgTooManyFiles)
	errNoFile       = rest.NewBadRequestError(nil, rest.MsgNoFile)
)

func AllowFilesAmount(key string, amount int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			form, err := c.MultipartForm()
			if err != nil {
				return rest.NewReadFileError(err)
			}

			files := form.File[key]

			if len(files) > amount {
				return errTooManyFiles
			}
			if len(files) < amount {
				return errNoFile
			}
			return next(c)
		}
	}
}

func AllowContentType(key string, types ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			form, err := c.MultipartForm()
			if err != nil {
				return rest.NewReadFileError(err)
			}

			for _, file := range form.File[key] {
				contentType := file.Header.Get("Content-Type")
				if slices.Contains(types, contentType) {
					continue
				}
				return rest.NewBadRequestError(nil, fmt.Sprintf("Invalid file type: %s, for file: %s", contentType, file.Filename))
			}

			return next(c)
		}
	}
}
