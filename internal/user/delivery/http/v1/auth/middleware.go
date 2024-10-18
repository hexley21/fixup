package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/jwt/refresh_jwt"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/http/writer"
)

type ctxKey string

var (
	refreshJwtIdKet       ctxKey              = "refresh_jwt_id"
	ErrRefreshTokenNotSet *rest.ErrorResponse = rest.NewInternalServerError(errors.New("refresh token is not set"))
)

type AuthMiddleware struct {
	writer writer.HTTPErrorWriter
}

func NewAuthMiddleware(writer writer.HTTPErrorWriter) *AuthMiddleware {
	return &AuthMiddleware{
		writer: writer,
	}
}

func (m *AuthMiddleware) RefreshJWT(jwtVerifier refresh_jwt.Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				m.writer.WriteError(w, middleware.ErrMissingAuthorizationHeader)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				m.writer.WriteError(w, middleware.ErrMissingBearerToken)
				return
			}

			claims, err := jwtVerifier.Verify(tokenString)
			if err != nil {
				m.writer.WriteError(w, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), refreshJwtIdKet, claims.ID)))
		})
	}
}
