package middleware_test

import (
	"net/http"

	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/pkg/http/binder/std_binder"
	"github.com/hexley21/fixup/pkg/http/json/std_json"
	"github.com/hexley21/fixup/pkg/http/writer/json_writer"
	"github.com/hexley21/fixup/pkg/logger/std_logger"
)

func BasicHandler() http.Handler {
	return http.HandlerFunc(BasicHandlerFunc)
}

func BasicHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func setupMiddleware() *middleware.Middleware {
	logger := std_logger.New()
	json := std_json.New()

	return middleware.NewMiddleware(
		std_binder.New(json),
		json_writer.New(logger, json),
	)

}

var (
	mw = setupMiddleware()
)