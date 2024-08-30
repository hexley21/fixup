package handler

import (
	"encoding/json"
	"net/http"

	"github.com/hexley21/handy/pkg/logger"
	"github.com/hexley21/handy/pkg/rest"
)

type ErrorHandlerFunc func(w http.ResponseWriter, r *http.Request) error
type ErrorHandler interface {
	ServeHttpError(w http.ResponseWriter, r *http.Request) error
}

type errorHandler struct {
	logger logger.Logger
	h      ErrorHandlerFunc
}

type HandlerFactory struct {
	logger logger.Logger
}

func NewErrorHandlerFactory(logger logger.Logger) *HandlerFactory {
	return &HandlerFactory{logger: logger}
}

func (f *HandlerFactory) NewHandlerFunc(h ErrorHandlerFunc) http.HandlerFunc {
	return f.newErrorHandler(h).ServeHTTP
}

func (f *HandlerFactory) NewHandler(h ErrorHandlerFunc) http.Handler {
	return f.newErrorHandler(h)
}

func (f *HandlerFactory) newErrorHandler(h ErrorHandlerFunc) *errorHandler {
	return &errorHandler{logger: f.logger, h: h}
}

func writeError(w http.ResponseWriter, err rest.ErrorResponse) error {
	if e := json.NewEncoder(w).Encode(err); e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return e
	}

	return err
}

func (h *errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.h(w, r)
	if err != nil {
		switch e := err.(type) {
		case rest.ErrorResponse:
			h.logger.ErrorCause(e, err)
			writeError(w, e)
		default:
			h.logger.ErrorCause(e, err)
			writeError(w, rest.InternalServerError)
		}
	}
}

func (h *errorHandler) ServeHttpError(w http.ResponseWriter, r *http.Request) error {
	return h.h(w, r)
}
