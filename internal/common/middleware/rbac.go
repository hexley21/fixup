package middleware

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/internal/common/util/ctxutil"
	"github.com/hexley21/fixup/internal/user/enum"
)

var (
	ErrInsufficientRights = rest.NewForbiddenError(nil, MsgInsufficientRights)
	ErrUserVerified       = rest.NewForbiddenError(nil, MsgUserIsVerified)
	ErrUserNotVerified    = rest.NewForbiddenError(nil, MsgUserIsNotVerified)
)

func (f *MiddlewareFactory) NewAllowRoles(roles ...enum.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, err := ctxutil.GetJwtRole(r.Context())
			if err != nil {
				f.writer.WriteError(w, err)
				return
			}

			if !slices.Contains(roles, role) {
				f.writer.WriteError(w, ErrInsufficientRights)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (f *MiddlewareFactory) NewAllowSelfOrRole(roles ...enum.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idParam := chi.URLParam(r, "id")

			role, err := ctxutil.GetJwtRole(r.Context())
			if err != nil {
				f.writer.WriteError(w, err)
				return
			}

			jwtId, err := ctxutil.GetJwtId(r.Context())
			if err != nil {
				f.writer.WriteError(w, err)
				return
			}

			if idParam == "me" {
				userId, err := strconv.ParseInt(jwtId, 10, 64)
				if err != nil {
					f.writer.WriteError(w, rest.NewInternalServerError(err))
					return
				}

				next.ServeHTTP(w, r.WithContext(ctxutil.SetParamId(r.Context(), userId)))
				return
			}

			if (idParam == jwtId) || slices.Contains(roles, role) {
				userId, err := strconv.ParseInt(idParam, 10, 64)
				if err != nil {
					f.writer.WriteError(w, rest.NewInternalServerError(err))
					return
				}

				r = r.WithContext(ctxutil.SetParamId(r.Context(), userId))
				next.ServeHTTP(w, r)
				return
			}

			f.writer.WriteError(w, ErrInsufficientRights)
		})
	}
}

func (f *MiddlewareFactory) NewAllowVerified(status bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			verified, err := ctxutil.GetJwtUserStatus(r.Context())
			if err != nil {
				f.writer.WriteError(w, err)
				return
			}

			if verified == status {
				next.ServeHTTP(w, r)
				return
			}

			if status {
				f.writer.WriteError(w, ErrUserNotVerified)
				return
			}

			f.writer.WriteError(w, ErrUserVerified)
		})
	}
}