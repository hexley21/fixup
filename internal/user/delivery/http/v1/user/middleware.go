package user

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/http/writer"
)

type ctxKey string

const paramIdKey ctxKey = "param_id"

var ErrParamIdNotSet = rest.NewInternalServerError(errors.New("param id is not set"))

type UserMiddleware struct {
	writer writer.HTTPErrorWriter
}

func NewUserMiddleware(writer writer.HTTPErrorWriter) *UserMiddleware {
	return &UserMiddleware{
		writer: writer,
	}
}

func (m *UserMiddleware) AllowSelfOrRole(roles ...enum.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			idParam := chi.URLParam(r, "id")

			claims, ok := r.Context().Value(auth_jwt.AuthJWTKey).(auth_jwt.UserData)
			if !ok {
				m.writer.WriteError(w, auth_jwt.ErrJWTNotSet)
				return
			}

			if idParam == "me" {
				id, err := strconv.ParseInt(claims.ID, 10, 64)
				if err != nil {
					m.writer.WriteError(w, rest.NewInternalServerError(err))
					return
				}

				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), paramIdKey, id)))
				return
			}

			if (idParam == claims.ID) || slices.Contains(roles, claims.Role) {
				id, err := strconv.ParseInt(idParam, 10, 64)
				if err != nil {
					m.writer.WriteError(w, rest.NewInternalServerError(err))
					return
				}

				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), paramIdKey, id)))
				return
			}

			m.writer.WriteError(w, rest.ErrInsufficientRights)
		})
	}
}
