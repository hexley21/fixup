package request_util

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/hexley21/fixup/pkg/http/rest"
)

var (
	ErrInvalidPage = rest.NewBadRequestError(errors.New("invalid page parameter"))
	ErrInvalidPerPage = rest.NewBadRequestError(errors.New("invalid per_page parameter"))
)

func ParseLimitAndOffset(r *http.Request, maxPerPage int64, defaultPerPage int64) (int64, int64, *rest.ErrorResponse) {
	pageParam := r.URL.Query().Get("page")
	perPageParam := r.URL.Query().Get("per_page")

	var page int64
	var perPage int64

	page, err := strconv.ParseInt(pageParam, 10, 64)
	if err != nil {
		return 0, 0, rest.NewBadRequestError(ErrInvalidPage)
	}

	if perPageParam != "" {
		perPage, err = strconv.ParseInt(perPageParam, 10, 64)
		if err != nil || perPage < 0 {
			return 0, 0, rest.NewBadRequestError(ErrInvalidPerPage)
		}
	}

	if perPage == 0 || perPage > maxPerPage {
		perPage = defaultPerPage
	}

	if page < 1 {
		return 0, 0, rest.NewBadRequestError(ErrInvalidPage)
	}

	return perPage, perPage * (page - 1), nil
}