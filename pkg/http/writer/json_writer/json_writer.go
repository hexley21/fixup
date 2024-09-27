package json_writer

import (
	"net/http"

	"github.com/hexley21/fixup/pkg/http/json"
	"github.com/hexley21/fixup/pkg/http/rest"
	"github.com/hexley21/fixup/pkg/logger"
)

var msgErrReturningResult = "Error returning result"

type jSONHTTPWriter struct {
	logger         logger.Logger
	jsonSerializer json.JSONSerializer
}

func New(logger logger.Logger, jsonSerializer json.JSONSerializer) *jSONHTTPWriter {
	return &jSONHTTPWriter{logger, jsonSerializer}
}

func (aw *jSONHTTPWriter) WriteData(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := aw.jsonSerializer.Serialize(w, rest.NewApiResponse(data)); err != nil {
		http.Error(w, msgErrReturningResult, http.StatusInternalServerError)
	}
}

func (aw *jSONHTTPWriter) WriteNoContent(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}

func (aw *jSONHTTPWriter) WriteError(w http.ResponseWriter, err *rest.ErrorResponse) {
	aw.logger.Error(err)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.Status)

	if err := aw.jsonSerializer.Serialize(w, err); err != nil {
		http.Error(w, msgErrReturningResult, http.StatusInternalServerError)
	}
}
