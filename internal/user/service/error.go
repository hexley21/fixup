package service

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUserNotUpdated = errors.New("user not updated")
	ErrUserEmailTaken = errors.New("user email is taken")
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrIncorrectEmailOrPassword = errors.New("incorrect email or password")

	ErrUserVerified          = errors.New("user is already verified")
	ErrUserNotRegistered     = errors.New("could not register user")
	ErrProviderNotRegistered = errors.New("could not register provider")
	ErrVerificationTokenUsed = errors.New("user verification token already used")
)

