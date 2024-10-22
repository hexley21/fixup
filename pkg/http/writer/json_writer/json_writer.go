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
	jsonSerializer json.Serializer
}

func New(logger logger.Logger, jsonSerializer json.Serializer) *jSONHTTPWriter {
	return &jSONHTTPWriter{logger, jsonSerializer}
}

// WriteData writes the provided data as a JSON response to the http.ResponseWriter.
// It sets the Content-Type header to "application/json" and writes the provided
// HTTP status code. If serialization fails, it writes an internal server error
// message to the response.
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

// WriteError writes the provided ErrorResponse as a JSON response to the http.ResponseWriter.
// It logs the error, sets the Content-Type header to "application/json", and writes the
// HTTP status code from the ErrorResponse. If serialization fails, it writes an internal
// server error message to the response.
func (aw *jSONHTTPWriter) WriteError(w http.ResponseWriter, err *rest.ErrorResponse) {
    aw.logger.Error(err)

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(err.Status)

    if err := aw.jsonSerializer.Serialize(w, err); err != nil {
        http.Error(w, msgErrReturningResult, http.StatusInternalServerError)
    }
}
