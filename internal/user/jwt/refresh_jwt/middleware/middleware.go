package middleware

import (
	"net/http"
	"strings"

	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/common/util/ctx_util"
	"github.com/hexley21/fixup/internal/user/jwt/refresh_jwt"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/http/writer"
)

func NewJWT(writer writer.HTTPErrorWriter, jwtVerifier refresh_jwt.JWTVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writer.WriteError(w, rest.NewUnauthorizedError(nil, middleware.MsgMissingAuthorizationHeader))
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				writer.WriteError(w, rest.NewUnauthorizedError(nil, middleware.MsgMissingBearerToken))
				return
			}

			claims, err := jwtVerifier.VerifyJWT(tokenString)
			if err != nil {
				writer.WriteError(w, err)
				return
			}

			ctx := ctx_util.SetJWTId(r.Context(), claims.ID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
