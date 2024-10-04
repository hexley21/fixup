package request_util_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hexley21/fixup/internal/common/app_error"
	"github.com/hexley21/fixup/internal/common/util/request_util"
	"github.com/stretchr/testify/assert"
)

func TestParsePagination_Success(t *testing.T) {
	q := make(url.Values)
	q.Set("page", "1")
	q.Set("per_page", "1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)

	err, page, perPage := request_util.ParsePagination(req)

	assert.Equal(t, 1, page)
	assert.Equal(t, 1, perPage)
	assert.Nil(t, err)
}

func TestParsePagination_EmptyPerPage(t *testing.T) {
	q := make(url.Values)
	q.Set("page", "1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)

	err, page, perPage := request_util.ParsePagination(req)

	assert.Equal(t, 1, page)
	assert.Equal(t, 0, perPage)
	assert.Nil(t, err)
}

func TestParsePagination_InvalidPage(t *testing.T) {
	q := make(url.Values)
	q.Set("page", "0")
	q.Set("per_page", "1")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)

	err, page, perPage := request_util.ParsePagination(req)

	assert.Equal(t, 0, page)
	assert.Equal(t, 0, perPage)

	assert.Nil(t, err.Cause)
	assert.Equal(t, app_error.MsgInvalidPage, err.Message)
	assert.Equal(t, http.StatusBadRequest, err.Status)
}

func TestParsePagination_NonNumberPage(t *testing.T) {
	q := make(url.Values)
	q.Set("page", "abc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)

	err, page, perPage := request_util.ParsePagination(req)

	assert.Equal(t, 0, page)
	assert.Equal(t, 0, perPage)

	assert.NotNil(t, err.Cause)
	assert.Equal(t, app_error.MsgInvalidPage, err.Message)
	assert.Equal(t, http.StatusBadRequest, err.Status)
}

func TestParsePagination_InvalidPerPage(t *testing.T) {
	q := make(url.Values)
	q.Set("page", "1")
	q.Set("per_page", "0")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)

	err, page, perPage := request_util.ParsePagination(req)

	assert.Equal(t, 0, page)
	assert.Equal(t, 0, perPage)

	assert.Nil(t, err.Cause)
	assert.Equal(t, app_error.MsgInvalidPerPage, err.Message)
	assert.Equal(t, http.StatusBadRequest, err.Status)
}

func TestParsePagination_NonNumberPerPage(t *testing.T) {
	q := make(url.Values)
	q.Set("page", "1")
	q.Set("per_page", "abc")

	req := httptest.NewRequest(http.MethodGet, "/?"+q.Encode(), nil)

	err, page, perPage := request_util.ParsePagination(req)

	assert.Equal(t, 0, page)
	assert.Equal(t, 0, perPage)

	assert.NotNil(t, err.Cause)
	assert.Equal(t, app_error.MsgInvalidPerPage, err.Message)
	assert.Equal(t, http.StatusBadRequest, err.Status)
}