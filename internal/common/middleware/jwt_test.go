package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/common/jwt"
	mock_jwt "github.com/hexley21/fixup/internal/common/jwt/mock"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	userClaims = jwt.NewClaims("1", string(enum.UserRoleCUSTOMER), true, time.Hour)
)

func setupJWT(t *testing.T) (*gomock.Controller, echo.MiddlewareFunc, *mock_jwt.MockJwtVerifier) {
	ctrl := gomock.NewController(t)
	mockJwtVerifier := mock_jwt.NewMockJwtVerifier(ctrl)

	return ctrl, middleware.JWT(mockJwtVerifier), mockJwtVerifier
}

func TestJWT_MissingAuthorizationHeader(t *testing.T) {
	JWTMiddleware := middleware.JWT(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, JWTMiddleware(BasicHandler)(c), &errResp) {
		assert.Equal(t, rest.MsgMissingAuthorizationHeader, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, errResp.Status)
	}
}

func TestJWT_MissingBearerToken(t *testing.T) {
	JWTMiddleware := middleware.JWT(nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "InvalidToken")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, JWTMiddleware(BasicHandler)(c), &errResp) {
		assert.Equal(t, rest.MsgMissingBearerToken, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, errResp.Status)
	}
}

func TestJWT_InvalidToken(t *testing.T) {
	ctrl, JWTMiddleware, mockJWTVerifier := setupJWT(t)
	defer ctrl.Finish()

	mockJWTVerifier.EXPECT().VerifyJWT(gomock.Any()).Return(jwt.UserClaims{}, rest.NewUnauthorizedError(errors.New(""), rest.MsgInvalidToken))

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var errResp *rest.ErrorResponse
	if assert.ErrorAs(t, JWTMiddleware(BasicHandler)(c), &errResp) {
		assert.Equal(t, rest.MsgInvalidToken, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, errResp.Status)
	}
}

func TestJWT_ValidToken(t *testing.T) {
	ctrl, JWTMiddleware, mockJWTVerifier := setupJWT(t)
	defer ctrl.Finish()

	mockJWTVerifier.EXPECT().VerifyJWT(gomock.Any()).Return(userClaims, nil)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer validtoken")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	assert.NoError(t, JWTMiddleware(BasicHandler)(c))
	assert.Equal(t, http.StatusOK, rec.Code)
}
