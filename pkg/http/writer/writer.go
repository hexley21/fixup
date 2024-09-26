package writer

import (
	"net/http"

	"github.com/hexley21/fixup/pkg/http/rest"
)


type HTTPWriter interface {
	HTTPDataWriter
	HTTPErrorWriter
}

type HTTPDataWriter interface {
	WriteData(w http.ResponseWriter, data any, code int)
}

type HTTPErrorWriter interface {
	WriteError(w http.ResponseWriter, err *rest.ErrorResponse)
}
