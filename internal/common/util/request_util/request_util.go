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

func ParseOffsetAndLimit(r *http.Request, maxPerPage int, defaultPerPage int) (int32, int32, *rest.ErrorResponse) {
	pageParam := r.URL.Query().Get("page")
	perPageParam := r.URL.Query().Get("per_page")

	var page int
	var perPage int

	page, err := strconv.Atoi(pageParam)
	if err != nil {
		return 0, 0, rest.NewBadRequestError(ErrInvalidPage)
	}

	if perPageParam != "" {
		perPage, err = strconv.Atoi(perPageParam)
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

	return int32(perPage * (page - 1)), int32(perPage), nil
}