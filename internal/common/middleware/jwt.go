package middleware

import (
	"net/http"
	"strings"

	"github.com/hexley21/fixup/internal/common/app_error"
	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/util/ctx_util"
	"github.com/hexley21/fixup/pkg/http/rest"
)

func (f *MiddlewareFactory) NewJWT(jwtVerifier auth_jwt.JWTVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				f.writer.WriteError(w, rest.NewUnauthorizedError(nil, MsgMissingAuthorizationHeader))
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				f.writer.WriteError(w, rest.NewUnauthorizedError(nil, MsgMissingBearerToken))
				return
			}

			claims, err := jwtVerifier.VerifyJWT(tokenString)
			if err != nil {
				f.writer.WriteError(w, err)
				return
			}

			if !claims.Role.Valid() {
				f.writer.WriteError(w, rest.NewUnauthorizedError(nil, app_error.MsgInvalidToken))
				return
			}

			ctx := ctx_util.SetJWTId(r.Context(), claims.ID)
			ctx = ctx_util.SetJWTRole(ctx, claims.Role)
			ctx = ctx_util.SetJWTUserStatus(ctx, claims.Verified)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
