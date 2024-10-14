package request_util

import (
	"net/http"
	"strconv"

	"github.com/hexley21/fixup/internal/common/app_error"
	"github.com/hexley21/fixup/pkg/http/rest"
)

func ParsePagination(r *http.Request) (*rest.ErrorResponse, int, int) {
	pageParam := r.URL.Query().Get("page")
	perPageParam := r.URL.Query().Get("per_page")

	var page int
	var perPage int

	page, err := strconv.Atoi(pageParam)
	if err != nil {
		return rest.NewBadRequestError(err, app_error.MsgInvalidPage), 0, 0
	}

	if perPageParam != "" {
		perPage, err = strconv.Atoi(perPageParam)
		if err != nil || perPage < 1{
			return rest.NewBadRequestError(err, app_error.MsgInvalidPerPage), 0, 0
		}
	}

	if page < 1 {
		return rest.NewBadRequestError(nil, app_error.MsgInvalidPage), 0, 0
	}

	return nil, page, perPage
}

func ParseOffsetAndLimit(r *http.Request, maxPerPage int, defaultPerPage int) (int32, int32, *rest.ErrorResponse) {
	pageParam := r.URL.Query().Get("page")
	perPageParam := r.URL.Query().Get("per_page")

	var page int
	var perPage int

	page, err := strconv.Atoi(pageParam)
	if err != nil {
		return 0, 0, rest.NewBadRequestError(err, app_error.MsgInvalidPage)
	}

	if perPageParam != "" {
		perPage, err = strconv.Atoi(perPageParam)
		if err != nil || perPage < 0 {
			return 0, 0, rest.NewBadRequestError(err, app_error.MsgInvalidPage)
		}

		if perPage == 0 || perPage > maxPerPage {
			perPage = defaultPerPage
		}
	}

	if page < 1 {
		return 0, 0, rest.NewBadRequestError(err, app_error.MsgInvalidPage)
	}

	return int32(perPage * (page - 1)), int32(perPage), nil
}