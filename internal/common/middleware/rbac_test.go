package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/internal/common/util/ctxutil"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func BasicHandler(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func TestAllowRoles_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctxutil.SetJwtRole(c, enum.UserRoleCUSTOMER)

	mw := middleware.AllowRoles(enum.UserRoleCUSTOMER, enum.UserRoleMODERATOR, enum.UserRoleADMIN)

	assert.NoError(t, mw(BasicHandler)(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAllowRoles_InsuffucientRights(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctxutil.SetJwtRole(c, enum.UserRoleCUSTOMER)

	mw := middleware.AllowRoles(enum.UserRoleADMIN, enum.UserRoleMODERATOR)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, mw(BasicHandler)(c), &errResp) {
		assert.Equal(t, rest.MsgInsufficientRights, errResp.Message)
		assert.Equal(t, http.StatusForbidden, errResp.Status)
	}
}

func TestAllowRole_JwtNotImplemented(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := middleware.AllowSelfOrRole(enum.UserRoleMODERATOR, enum.UserRoleADMIN)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, mw(BasicHandler)(c), &errResp) {
		assert.Equal(t, ctxutil.ErrJwtNotImplemented, errResp.Cause)
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}

func TestAllowSelfOrRole_SelfMe(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("me")

	ctxutil.SetJwtRole(c, enum.UserRoleCUSTOMER)
	ctxutil.SetJwtId(c, "123")

	mw := middleware.AllowSelfOrRole(enum.UserRoleMODERATOR, enum.UserRoleADMIN)

	assert.NoError(t, mw(BasicHandler)(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	id, err := ctxutil.GetParamId(c)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), id)
}

func TestAllowSelfOrRole_SelfId(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("123")

	ctxutil.SetJwtRole(c, enum.UserRoleCUSTOMER)
	ctxutil.SetJwtId(c, "123")

	mw := middleware.AllowSelfOrRole(enum.UserRoleMODERATOR, enum.UserRoleADMIN)

	assert.NoError(t, mw(BasicHandler)(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	id, err := ctxutil.GetParamId(c)
	assert.NoError(t, err)
	assert.Equal(t, int64(123), id)
}

func TestAllowSelfOrRole_AllowedRole(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("321")

	ctxutil.SetJwtRole(c, enum.UserRoleADMIN)
	ctxutil.SetJwtId(c, "123")

	mw := middleware.AllowSelfOrRole(enum.UserRoleMODERATOR, enum.UserRoleADMIN)

	assert.NoError(t, mw(BasicHandler)(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	id, err := ctxutil.GetParamId(c)
	assert.NoError(t, err)
	assert.Equal(t, int64(321), id)
}

func TestAllowSelfOrRole_InsufficientRights(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("321")

	ctxutil.SetJwtRole(c, enum.UserRoleCUSTOMER)
	ctxutil.SetJwtId(c, "123")

	mw := middleware.AllowSelfOrRole(enum.UserRoleMODERATOR, enum.UserRoleADMIN)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, mw(BasicHandler)(c), &errResp) {
		assert.Equal(t, rest.MsgInsufficientRights, errResp.Message)
		assert.Equal(t, http.StatusForbidden, errResp.Status)
	}

	id, err := ctxutil.GetParamId(c)
	assert.Error(t, err)
	assert.Equal(t, int64(0), id)
}

func TestAllowSelfOrRole_JwtNotImplemented(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	t.Run("Id", func(t *testing.T) {
		c := e.NewContext(req, rec)

		mw := middleware.AllowSelfOrRole(enum.UserRoleMODERATOR, enum.UserRoleADMIN)

		var errResp *rest.ErrorResponse
		if assert.ErrorAs(t, mw(BasicHandler)(c), &errResp) {
			assert.Equal(t, ctxutil.ErrJwtNotImplemented, errResp.Cause)
			assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
			assert.Equal(t, http.StatusInternalServerError, errResp.Status)
		}

		id, err := ctxutil.GetParamId(c)
		assert.Error(t, err)
		assert.Equal(t, int64(0), id)
	})

	t.Run("Role", func(t *testing.T) {
		c := e.NewContext(req, rec)
		ctxutil.SetJwtId(c, "123")

		mw := middleware.AllowSelfOrRole(enum.UserRoleMODERATOR, enum.UserRoleADMIN)

		var errResp *rest.ErrorResponse
		if assert.ErrorAs(t, mw(BasicHandler)(c), &errResp) {
			assert.Equal(t, ctxutil.ErrJwtNotImplemented, errResp.Cause)
			assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
			assert.Equal(t, http.StatusInternalServerError, errResp.Status)
		}

		id, err := ctxutil.GetParamId(c)
		assert.Error(t, err)
		assert.Equal(t, int64(0), id)
	})
}

func TestAllowVerified_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctxutil.SetJwtUserStatus(c, true)

	mw := middleware.AllowVerified(true)

	assert.NoError(t, mw(BasicHandler)(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestAllowVerified_ErrorNotVerified(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctxutil.SetJwtUserStatus(c, false)

	mw := middleware.AllowVerified(true)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, mw(BasicHandler)(c), &errResp) {
		assert.Equal(t, middleware.ErrUserNotVerified.Cause, errResp.Cause)
		assert.Equal(t, rest.MsgUserIsNotVerified, errResp.Message)
		assert.Equal(t, http.StatusForbidden, errResp.Status)
	}
}

func TestAllowVerified_ErrorVerified(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	ctxutil.SetJwtUserStatus(c, true)

	mw := middleware.AllowVerified(false)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, mw(BasicHandler)(c), &errResp) {
		assert.Equal(t, middleware.ErrUserVerified.Cause, errResp.Cause)
		assert.Equal(t, rest.MsgUserIsVerified, errResp.Message)
		assert.Equal(t, http.StatusForbidden, errResp.Status)
	}
}

func TestAllowVerified_JwtNotImplemented(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := middleware.AllowVerified(true)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, mw(BasicHandler)(c), &errResp) {
		assert.Equal(t, ctxutil.ErrJwtNotImplemented, errResp.Cause)
		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
		assert.Equal(t, http.StatusInternalServerError, errResp.Status)
	}
}
