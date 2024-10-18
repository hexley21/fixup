package middleware

import (
	"errors"
	"net/http"
	"slices"

	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/pkg/http/rest"
)

var (
	ErrUserVerified       = rest.NewForbiddenError(errors.New("user has to be not-verified"))
	ErrUserNotVerified    = rest.NewForbiddenError(errors.New("user is not verified"))
)

func (f *Middleware) NewAllowRoles(roles ...enum.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(auth_jwt.AuthJWTKey).(auth_jwt.UserData)
			if !ok {
				f.writer.WriteError(w, auth_jwt.ErrJWTNotSet)
				return
			}

			if !slices.Contains(roles, claims.Role) {
				f.writer.WriteError(w, rest.ErrInsufficientRights)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (f *Middleware) NewAllowVerified(verified bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(auth_jwt.AuthJWTKey).(auth_jwt.UserData)
			if !ok {
				f.writer.WriteError(w, auth_jwt.ErrJWTNotSet)
				return
			}

			if claims.Verified == verified {
				next.ServeHTTP(w, r)
				return
			}

			if verified {
				f.writer.WriteError(w, ErrUserNotVerified)
				return
			}

			f.writer.WriteError(w, ErrUserVerified)
		})
	}
}
