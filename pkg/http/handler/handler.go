package handler

import (
	"github.com/hexley21/fixup/pkg/http/binder"
	"github.com/hexley21/fixup/pkg/http/writer"
	"github.com/hexley21/fixup/pkg/logger"
	"github.com/hexley21/fixup/pkg/validator"
)

type Components struct {
	Logger    logger.Logger
	Binder    binder.FullBinder
	Validator validator.Validator
	Writer    writer.HTTPWriter
}

func NewComponents(Logger logger.Logger, Binder binder.FullBinder, Validator validator.Validator, Writer writer.HTTPWriter) *Components {
	return &Components{
		Logger:    Logger,
		Binder:    Binder,
		Validator: Validator,
		Writer:    Writer,
	}
}
