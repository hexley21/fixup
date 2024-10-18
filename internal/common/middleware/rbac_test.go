package middleware_test

// TODO: rewrite rbac middleware tests
// import (
// 	"encoding/json"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/hexley21/fixup/internal/common/enum"
// 	"github.com/hexley21/fixup/internal/common/middleware"
// 	"github.com/hexley21/fixup/internal/common/util/ctx_util"
// 	"github.com/hexley21/fixup/pkg/http/rest"
// 	"github.com/stretchr/testify/assert"
// )

// func ParamTestHandlerFunc(t *testing.T, value int64) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		id, err := ctx_util.GetParamId(r.Context())
// 		assert.Equal(t, value, id)
// 		assert.Nil(t, err)
// 		w.WriteHeader(http.StatusOK)
// 	}
// }

// func TestAllowRoles_Success(t *testing.T) {
// 	req := httptest.NewRequest(http.MethodGet, "/", nil)
// 	rec := httptest.NewRecorder()

// 	req = req.WithContext(ctx_util.SetJWTRole(req.Context(), enum.UserRoleCUSTOMER))

// 	mw.NewAllowRoles(enum.UserRoleCUSTOMER, enum.UserRoleMODERATOR, enum.UserRoleADMIN)(BasicHandler()).ServeHTTP(rec, req)

// 	assert.Equal(t, "ok", rec.Body.String())
// 	assert.Equal(t, http.StatusOK, rec.Code)
// }

// func TestAllowRoles_InsufficientRights(t *testing.T) {
// 	req := httptest.NewRequest(http.MethodGet, "/", nil)
// 	rec := httptest.NewRecorder()

// 	req = req.WithContext(ctx_util.SetJWTRole(req.Context(), enum.UserRoleCUSTOMER))

// 	mw.NewAllowRoles(enum.UserRoleADMIN, enum.UserRoleMODERATOR)(BasicHandler()).ServeHTTP(rec, req)

// 	assert.Equal(t, http.StatusForbidden, rec.Code)

// 	var errResp rest.ErrorResponse
// 	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
// 		assert.Equal(t, middleware.ErrInsufficientRights.Message, errResp.Message)
// 		assert.Equal(t, http.StatusForbidden, rec.Code)
// 	}
// }

// func TestAllowRole_JwtNotImplemented(t *testing.T) {
// 	req := httptest.NewRequest(http.MethodGet, "/", nil)
// 	rec := httptest.NewRecorder()

// 	mw.NewAllowSelfOrRole(enum.UserRoleMODERATOR, enum.UserRoleADMIN)(BasicHandler()).ServeHTTP(rec, req)

// 	assert.Equal(t, http.StatusInternalServerError, rec.Code)

// 	var errResp rest.ErrorResponse
// 	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
// 		assert.Equal(t, rest.MsgInternalServerError, errResp.Message)
// 		assert.Equal(t, http.StatusInternalServerError, rec.Code)
// 	}
// }

