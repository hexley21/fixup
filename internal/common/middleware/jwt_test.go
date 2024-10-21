package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hexley21/fixup/internal/common/auth_jwt"
	mockJwt "github.com/hexley21/fixup/internal/common/auth_jwt/mock"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	userClaims = auth_jwt.NewClaims("1", enum.UserRoleCUSTOMER, true, time.Hour)
)

func setupJWT(t *testing.T) (*gomock.Controller, func(http.Handler) http.Handler, *mockJwt.MockVerifier) {
	mw := setupMiddleware()
	ctrl := gomock.NewController(t)
	JWTVerifierMock := mockJwt.NewMockVerifier(ctrl)

	return ctrl, mw.NewJWT(JWTVerifierMock), JWTVerifierMock
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
		assert.Equal(t, middleware.ErrMissingAuthorizationHeader.Message, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
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
		assert.Equal(t, middleware.ErrMissingBearerToken.Message, errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}
}

func TestJWT_InvalidToken(t *testing.T) {
	ctrl, JWTMiddleware, mockJWTVerifier := setupJWT(t)
	defer ctrl.Finish()

	mockJWTVerifier.EXPECT().Verify(gomock.Any()).Return(auth_jwt.UserClaims{}, rest.NewUnauthorizedError(jwt.ErrInvalidKey))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	rec := httptest.NewRecorder()

	JWTMiddleware(BasicHandler()).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var errResp rest.ErrorResponse
	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
		assert.Equal(t, jwt.ErrInvalidKey.Error(), errResp.Message)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	}
}

func TestJWT_ValidToken(t *testing.T) {
	ctrl, JWTMiddleware, mockJWTVerifier := setupJWT(t)
	defer ctrl.Finish()

	mockJWTVerifier.EXPECT().Verify(gomock.Any()).Return(userClaims, nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer validtoken")
	rec := httptest.NewRecorder()

	JWTMiddleware(BasicHandler()).ServeHTTP(rec, req)

	assert.Equal(t, "ok", rec.Body.String())
	assert.Equal(t, http.StatusOK, rec.Code)
}
