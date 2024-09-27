package middleware_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/common/app_error"
	"github.com/hexley21/fixup/internal/common/auth_jwt"
	mock_jwt "github.com/hexley21/fixup/internal/common/auth_jwt/mock"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	userClaims = auth_jwt.NewClaims("1", string(enum.UserRoleCUSTOMER), true, time.Hour)
)

func setupJWT(t *testing.T) (*gomock.Controller, func(http.Handler) http.Handler, *mock_jwt.MockJWTVerifier) {
	ctrl := gomock.NewController(t)
	mockJwtVerifier := mock_jwt.NewMockJWTVerifier(ctrl)


	return ctrl, factory.NewJWT(mockJwtVerifier), mockJwtVerifier
}

func TestJWT_MissingAuthorizationHeader(t *testing.T) {
	ctrl, JWTMiddleware, _ := setupJWT(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	JWTMiddleware(BasicHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, middleware.MsgMissingAuthorizationHeader, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, errResp.Status)
	}

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWT_MissingBearerToken(t *testing.T) {
	ctrl, JWTMiddleware, _ := setupJWT(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "InvalidToken")
	rec := httptest.NewRecorder()

	JWTMiddleware(BasicHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, middleware.MsgMissingBearerToken, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, errResp.Status)
	}
}

func TestJWT_InvalidToken(t *testing.T) {
	ctrl, JWTMiddleware, mockJWTVerifier := setupJWT(t)
	defer ctrl.Finish()

	mockJWTVerifier.EXPECT().VerifyJWT(gomock.Any()).Return(auth_jwt.UserClaims{}, rest.NewUnauthorizedError(errors.New(""), app_error.MsgInvalidToken))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	rec := httptest.NewRecorder()

	JWTMiddleware(BasicHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	
	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, app_error.MsgInvalidToken, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, errResp.Status)
	}
}

func TestJWT_ValidToken(t *testing.T) {
	ctrl, JWTMiddleware, mockJWTVerifier := setupJWT(t)
	defer ctrl.Finish()

	mockJWTVerifier.EXPECT().VerifyJWT(gomock.Any()).Return(userClaims, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer validtoken")
	rec := httptest.NewRecorder()
	
	JWTMiddleware(BasicHandler()).ServeHTTP(rec, req)

	assert.Equal(t, "ok", rec.Body.String())
	assert.Equal(t, http.StatusOK, rec.Code)
}
