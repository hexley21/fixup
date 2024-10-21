package user_test

// TODO: Refactor allow self or role mdidleware tests

// func TestAllowSelfOrRole_SelfMe(t *testing.T) {
// 	r := chi.NewRouter()
// 	r.Group(func(r chi.Router) {
// 		r.Use(mw.NewAllowSelfOrRole(enum.UserRoleMODERATOR, enum.UserRoleADMIN))
// 		r.Get("/{id}", ParamTestHandlerFunc(t, int64(123)))
// 	})

// 	req := httptest.NewRequest(http.MethodGet, "/me", nil)
// 	rec := httptest.NewRecorder()

// 	ctx := ctx_util.SetJWTRole(req.Context(), enum.UserRoleCUSTOMER)
// 	ctx = ctx_util.SetJWTId(ctx, "123")
// 	ctx = ctx_util.SetJWTUserStatus(ctx, true)
// 	req = req.WithContext(ctx)

// 	r.ServeHTTP(rec, req)
// }

// func TestAllowSelfOrRole_SelfId(t *testing.T) {
// 	r := chi.NewRouter()
// 	r.Group(func(r chi.Router) {
// 		r.Use(mw.NewAllowSelfOrRole(enum.UserRoleMODERATOR, enum.UserRoleADMIN))
// 		r.Get("/{id}", ParamTestHandlerFunc(t, int64(123)))
// 	})

// 	req := httptest.NewRequest(http.MethodGet, "/123", nil)
// 	rec := httptest.NewRecorder()

// 	ctx := ctx_util.SetJWTRole(req.Context(), enum.UserRoleCUSTOMER)
// 	ctx = ctx_util.SetJWTId(ctx, "123")
// 	ctx = ctx_util.SetJWTUserStatus(ctx, true)
// 	req = req.WithContext(ctx)

// 	r.ServeHTTP(rec, req)
// }

// func TestAllowSelfOrRole_AllowedRole(t *testing.T) {
// 	r := chi.NewRouter()
// 	r.Group(func(r chi.Router) {
// 		r.Use(mw.NewAllowSelfOrRole(enum.UserRoleMODERATOR, enum.UserRoleADMIN))
// 		r.Get("/{id}", ParamTestHandlerFunc(t, int64(321)))
// 	})

// 	req := httptest.NewRequest(http.MethodGet, "/321", nil)
// 	rec := httptest.NewRecorder()

// 	ctx := ctx_util.SetJWTRole(req.Context(), enum.UserRoleADMIN)
// 	ctx = ctx_util.SetJWTId(ctx, "123")
// 	ctx = ctx_util.SetJWTUserStatus(ctx, true)

// 	r.ServeHTTP(rec, req.WithContext(ctx))
// }

// func TestAllowSelfOrRole_InsufficientRights(t *testing.T) {
// 	r := chi.NewRouter()
// 	r.Group(func(r chi.Router) {
// 		r.Use(mw.NewAllowSelfOrRole(enum.UserRoleMODERATOR, enum.UserRoleADMIN))
// 		r.Get("/{id}", BasicHandlerFunc)
// 	})

// 	req := httptest.NewRequest(http.MethodGet, "/321", nil)
// 	rec := httptest.NewRecorder()

// 	ctx := ctx_util.SetJWTRole(req.Context(), enum.UserRoleCUSTOMER)
// 	ctx = ctx_util.SetJWTId(ctx, "123")
// 	ctx = ctx_util.SetJWTUserStatus(ctx, true)

// 	r.ServeHTTP(rec, req.WithContext(ctx))

// 	var errResp rest.ErrorResponse
// 	if assert.NoError(t, json.NewDecoder(rec.Body).Decode(&errResp)) {
// 		assert.Equal(t, middleware.ErrInsufficientRights.Message, errResp.Message)
// 		assert.Equal(t, http.StatusForbidden, rec.Code)
// 	}
// }