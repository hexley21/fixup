package middleware

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/common/util/ctx_util"
	"github.com/hexley21/fixup/pkg/http/rest"
)

var (
	ErrInsufficientRights = rest.NewForbiddenError(nil, MsgInsufficientRights)
	ErrUserVerified       = rest.NewForbiddenError(nil, MsgUserIsVerified)
	ErrUserNotVerified    = rest.NewForbiddenError(nil, MsgUserIsNotVerified)
)

func (f *Middleware) NewAllowRoles(roles ...enum.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, err := ctx_util.GetJWTRole(r.Context())
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

func (f *Middleware) NewAllowSelfOrRole(roles ...enum.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idParam := chi.URLParam(r, "id")

			role, err := ctx_util.GetJWTRole(r.Context())
			if err != nil {
				f.writer.WriteError(w, err)
				return
			}

			jwtId, err := ctx_util.GetJWTId(r.Context())
			if err != nil {
				f.writer.WriteError(w, err)
				return
			}

			if idParam == "me" {
				id, err := strconv.ParseInt(jwtId, 10, 64)
				if err != nil {
					f.writer.WriteError(w, rest.NewInternalServerError(err))
					return
				}

				next.ServeHTTP(w, r.WithContext(ctx_util.SetParamId(r.Context(), id)))
				return
			}

			if (idParam == jwtId) || slices.Contains(roles, role) {
				id, err := strconv.ParseInt(idParam, 10, 64)
				if err != nil {
					f.writer.WriteError(w, rest.NewInternalServerError(err))
					return
				}

				r = r.WithContext(ctx_util.SetParamId(r.Context(), id))
				next.ServeHTTP(w, r)
				return
			}

			f.writer.WriteError(w, ErrInsufficientRights)
		})
	}
}

func (f *Middleware) NewAllowVerified(status bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			verified, err := ctx_util.GetJWTUserStatus(r.Context())
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
