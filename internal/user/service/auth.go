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
	"github.com/hexley21/fixup/pkg/infra/cdn"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/mailer"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
)

var (
	ErrAlreadyVerified = errors.New("user is already verified")
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
	GetAccountInfo(ctx context.Context, id int64) (domain.UserAccountInfo, error)
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
	cdnUrlSigner           cdn.URLSigner
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
	cdnUrlSigner cdn.URLSigner,
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
		cdnUrlSigner:           cdnUrlSigner,
	}
}

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

// RegisterProvider inserts records in user & provider tables
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

func (s *authServiceImpl) GetAccountInfo(ctx context.Context, id int64) (domain.UserAccountInfo, error) {
	accountInfo, err := s.userRepository.GetAccountInfo(ctx, id)
	if err != nil {
		return domain.UserAccountInfo{}, err
	}

	return MapUserAccountInfo(accountInfo.Role, accountInfo.Verified)
}

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

func (s *authServiceImpl) SendVerificationLetter(ctx context.Context, token string, email string, name string) error {
	isUsed, err := s.verificationRepository.IsTokenUsed(ctx, token)
	if err != nil {
		return err
	}
	if isUsed {
		return ErrAlreadyVerified
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

func (s *authServiceImpl) SendVerificationSuccessLetter(email string) error {
	return s.mailer.SendHTML(
		s.emailAddress,
		email,
		"Verification Success",
		s.templates.verificationSuccess,
		nil,
	)
}
