package validator

import (
	"github.com/hexley21/fixup/pkg/http/rest"
)

type Validator interface {
	Validate(i any) *rest.ErrorResponse
}
