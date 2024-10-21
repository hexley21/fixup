package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/pkg/http/rest"
)

func (f *Middleware) NewJWT(jwtVerifier auth_jwt.Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				f.writer.WriteError(w, ErrMissingAuthorizationHeader)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				f.writer.WriteError(w, ErrMissingBearerToken)
				return
			}

			claims, err := jwtVerifier.Verify(tokenString)
			if err != nil {
				f.writer.WriteError(w, err)
				return
			}

			if !claims.Data.Role.Valid() {
				f.writer.WriteError(w, rest.NewUnauthorizedError(enum.ErrInvalidRole))
				return
			}

			ctx := context.WithValue(r.Context(), auth_jwt.AuthJWTKey, claims.Data)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
