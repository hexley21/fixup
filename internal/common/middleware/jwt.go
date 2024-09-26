package middleware

import (
	"net/http"
	"strings"

	"github.com/hexley21/fixup/internal/common/app_error"
	authjwt "github.com/hexley21/fixup/internal/common/jwt"
	"github.com/hexley21/fixup/internal/common/util/ctxutil"
	"github.com/hexley21/fixup/pkg/http/rest"
)

func (f *MiddlewareFactory) NewJWT(jwtVerifier authjwt.JwtVerifier) func(http.Handler) http.Handler {
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

			ctx := ctxutil.SetJwtId(r.Context(), claims.ID)
			ctx = ctxutil.SetJwtRole(ctx, claims.Role)
			ctx = ctxutil.SetJwtUserStatus(ctx, claims.Verified)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
