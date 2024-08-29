package service

import (
	"context"

	"github.com/hexley21/handy/pkg/logger"
)

type EmailService interface {
	SendConfirmation(ctx context.Context, email string) error
}

type emailServiceImpl struct {
	logger logger.Logger
}

func NewEmailService(logger logger.Logger) EmailService {
	return &emailServiceImpl{
		logger,
	}
}

func (s *emailServiceImpl) SendConfirmation(ctx context.Context, email string) error {
	s.logger.Warn("NOT IMPLEMENTED")
	return nil
}
