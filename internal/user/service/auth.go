package service

import (
	"context"
	"errors"
	"html/template"
	"time"

	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/user/domain"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/encryption"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/mailer"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
)

type templates struct {
	verification        *template.Template
	verificationSuccess *template.Template
}

func NewTemplates(verification *template.Template, verificationSuccess *template.Template) *templates {
	return &templates{verification: verification, verificationSuccess: verificationSuccess}
}

type AuthService interface {
	RegisterCustomer(ctx context.Context, password string, personalInfo *domain.UserPersonalInfo) (*domain.User, error)
	RegisterProvider(ctx context.Context, password string, personalIdNumber string, personalInfo *domain.UserPersonalInfo) (*domain.User, error)
	AuthenticateUser(ctx context.Context, email string, password string) (domain.UserIdentity, error)
	RefreshUserToken(ctx context.Context, id int64, tokenFunc func(role enum.UserRole, verified bool) (string, error)) (string, error)
	VerifyUser(ctx context.Context, token string, ttl time.Duration, id int64) error
	ResendVerificationLetter(ctx context.Context, tokenFunc func(id int64) (string, error), email string) error
	SendVerificationLetter(ctx context.Context, token string, email string, name string) error
	SendVerificationSuccessLetter(email string) error
}

type authServiceImpl struct {
	userRepository         repository.UserRepository
	providerRepository     repository.ProviderRepository
	verificationRepository repository.VerificationRepository
	verificationTokenTTL   time.Duration
	pgx                    postgres.PGX
	hasher                 hasher.Hasher
	encryptor              encryption.Encryptor
	mailer                 mailer.Mailer
	emailAddress           string
	templates              *templates
}

func NewAuthService(
	userRepository repository.UserRepository,
	providerRepository repository.ProviderRepository,
	verificationRepository repository.VerificationRepository,
	verificationTokenTTL time.Duration,
	pgx postgres.PGX,
	hasher hasher.Hasher,
	encryptor encryption.Encryptor,
	mailer mailer.Mailer,
	emailAddress string,
) *authServiceImpl {
	return &authServiceImpl{
		userRepository:         userRepository,
		providerRepository:     providerRepository,
		verificationRepository: verificationRepository,
		verificationTokenTTL:   verificationTokenTTL,
		pgx:                    pgx,
		hasher:                 hasher,
		encryptor:              encryptor,
		mailer:                 mailer,
		emailAddress:           emailAddress,
	}
}

// ParseTemplates parses the email templates from the provided configuration paths.
// It is called once to ensure the templates are available for use at any time.
// It returns an error if any template fails to parse.
// This was done so test would not need to read files.
func (s *authServiceImpl) ParseTemplates(cfg config.Templates) error {
	verificationTemplate, err := template.ParseFiles(cfg.VerificationPath)
	if err != nil {
		return err
	}
	verificationSuccessTemplate, err := template.ParseFiles(cfg.VerificationSuccessPath)
	if err != nil {
		return err
	}

	s.templates = NewTemplates(verificationTemplate, verificationSuccessTemplate)
	return nil
}

func (s *authServiceImpl) SetTemplates(verificationTemplate *template.Template, verificationSuccessTemplate *template.Template) {
	s.templates = NewTemplates(verificationTemplate, verificationSuccessTemplate)
}

// RegisterProvider writes user record to a database, returns domain user result.
func (s *authServiceImpl) RegisterCustomer(ctx context.Context, password string, personalInfo *domain.UserPersonalInfo) (*domain.User, error) {
	hash, err := s.hasher.HashPassword(password)
	if err != nil {
		return nil, err
	}

	userModel, err := s.userRepository.Create(ctx,
		repository.CreateUserParams{
			FirstName:   personalInfo.FirstName,
			LastName:    personalInfo.LastName,
			PhoneNumber: personalInfo.PhoneNumber,
			Email:       personalInfo.Email,
			Hash:        hash,
			Role:        string(enum.UserRoleCUSTOMER),
		})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotRegistered
		}
		return nil, err
	}

	return MapUserModelToEntity(userModel)
}

// RegisterProvider writes user and provider records to a database, returns domain user result.
func (s *authServiceImpl) RegisterProvider(ctx context.Context, password string, personalIdNumber string, personalInfo *domain.UserPersonalInfo) (*domain.User, error) {
	// hash a password first
	hash, err := s.hasher.HashPassword(password)
	if err != nil {
		return nil, err
	}

	tx, err := s.pgx.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}

	// insert user first & check for errors or if email is taken
	userModel, err := s.userRepository.WithTx(tx).Create(ctx,
		repository.CreateUserParams{
			FirstName:   personalInfo.FirstName,
			LastName:    personalInfo.LastName,
			PhoneNumber: personalInfo.PhoneNumber,
			Email:       personalInfo.Email,
			Hash:        hash,
			Role:        string(enum.UserRoleCUSTOMER),
		},
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, postgres.Rollback(tx, ctx, ErrUserNotRegistered)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return nil, postgres.Rollback(tx, ctx, ErrUserEmailTaken)
			}
		}
		return nil, postgres.Rollback(tx, ctx, err)
	}

	enc, err := s.encryptor.Encrypt([]byte(personalIdNumber))
	if err != nil {
		return nil, postgres.Rollback(tx, ctx, err)
	}

	// insert provider & check for errors
	ok, err := s.providerRepository.WithTx(tx).Create(ctx, repository.CreateProviderParams{
		PersonalIDNumber:  enc,
		PersonalIDPreview: personalIdNumber[len(personalIdNumber)-5:],
		UserID:            userModel.ID,
	})
	if err != nil {
		return nil, postgres.Rollback(tx, ctx, err)
	}
	if !ok {
		return nil, postgres.Rollback(tx, ctx, ErrProviderNotRegistered)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return MapUserModelToEntity(userModel)
}

// AuthenticateUser authenticates a user by verifying their email and password.
// It returns error if password is incorrect
func (s *authServiceImpl) AuthenticateUser(ctx context.Context, email string, password string) (domain.UserIdentity, error) {
	authInfo, err := s.userRepository.GetAuthInfoByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserIdentity{}, ErrUserNotFound
		}
		return domain.UserIdentity{}, err
	}

	err = s.hasher.VerifyPassword(password, authInfo.Hash)
	if err != nil {
		if errors.Is(err, hasher.ErrPasswordMismatch) {
			return domain.UserIdentity{}, ErrIncorrectEmailOrPassword
		}

		return domain.UserIdentity{}, err
	}

	return MapUserIdentity(authInfo.ID, authInfo.Role, authInfo.Verified)
}

// RefreshUserToken retrieves user's current accout information and returns a new access token.
// It returns an error if the user is not found or if any other error occurs during the process.
func (s *authServiceImpl) RefreshUserToken(ctx context.Context, id int64, tokenFunc func(role enum.UserRole, verified bool) (string, error)) (string, error) {
	accountInfo, err := s.userRepository.GetAccountInfo(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrUserNotFound
		}

		return "", err
	}


	return tokenFunc(enum.UserRole(accountInfo.Role), accountInfo.Verified.Bool)
}

// VerifyUser verifies a user by setting the token as used and updating the user's verification status.
// It returns an error if the token has already been used, the user update fails, or any other error occurs.
func (s *authServiceImpl) VerifyUser(ctx context.Context, token string, ttl time.Duration, id int64) error {
	err := s.verificationRepository.SetTokenUsed(ctx, token, ttl)
	if err != nil {
		if errors.Is(err, redis.TxFailedErr) {
			return ErrVerificationTokenUsed
		}
		return err
	}

	ok, err := s.userRepository.UpdateVerification(ctx, id, true)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUserNotUpdated
	}

	return nil
}

// ResendVerificationLetter resends a verification email to the specified address.
// It retrieves the user's verification info from the repository and checks if the user is already verified.
// If the user is not verified, it generates a new token and sends a verification email.
func (s *authServiceImpl) ResendVerificationLetter(ctx context.Context, tokenFunc func(id int64) (string, error), email string) error {
	verificationInfo, err := s.userRepository.GetVerificationInfo(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}

		return err
	}
	if !verificationInfo.Verified.Bool {
		return ErrUserVerified
	}

	token, err := tokenFunc(verificationInfo.ID)
	if err != nil {
		return err
	}

	return s.SendVerificationLetter(ctx, token, email, verificationInfo.FirstName)
}

// SendVerificationLetter sends an account verification email to the specified address.
// It checks if the token has already been used and returns an error if it has.
func (s *authServiceImpl) SendVerificationLetter(ctx context.Context, token string, email string, name string) error {
	isUsed, err := s.verificationRepository.IsTokenUsed(ctx, token)
	if err != nil {
		return err
	}
	if isUsed {
		return ErrUserVerified
	}

	return s.mailer.SendHTML(
		s.emailAddress,
		email,
		"Account verification",
		s.templates.verification,
		struct {
			Name  string
			Token string
		}{
			Name:  name,
			Token: token,
		},
	)
}

// SendVerificationSuccessLetter sends a verification success email to the specified address.
func (s *authServiceImpl) SendVerificationSuccessLetter(email string) error {
	return s.mailer.SendHTML(
		s.emailAddress,
		email,
		"Verification Success",
		s.templates.verificationSuccess,
		nil,
	)
}
