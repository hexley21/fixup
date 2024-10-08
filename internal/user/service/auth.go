package service

import (
	"context"
	"errors"
	"html/template"
	"strconv"
	"time"

	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto/mapper"
	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/hexley21/fixup/pkg/encryption"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/infra/cdn"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/mailer"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	ErrAlreadyVerified = errors.New("user is already activated")
)

type UserConfirmationDetails struct {
	ID         string
	UserStatus bool
	Firstname  string
}

type UserIdentity struct {
	ID         string
	Role       string
	UserStatus bool
}

type UserRoleAndStatus struct {
	Role       string
	UserStatus bool
}

type templates struct {
	confirmation *template.Template
	verified     *template.Template
}

func NewTemplates(confirmation *template.Template, verified *template.Template) *templates {
	return &templates{confirmation: confirmation, verified: verified}
}

type AuthService interface {
	RegisterCustomer(ctx context.Context, registerDto dto.RegisterUser) (dto.User, error)
	RegisterProvider(ctx context.Context, registerDto dto.RegisterProvider) (dto.User, error)
	AuthenticateUser(ctx context.Context, loginDto dto.Login) (UserIdentity, error)
	VerifyUser(ctx context.Context, token string, ttl time.Duration, id int64, email string) error
	GetUserConfirmationDetails(ctx context.Context, email string) (UserConfirmationDetails, error)
	GetUserRoleAndStatus(ctx context.Context, id int64) (UserRoleAndStatus, error)
	SendConfirmationLetter(ctx context.Context, token string, email string, name string) error
	SendVerifiedLetter(email string) error
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
	emailAddres            string
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
	emailAddres string,
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
		emailAddres:            emailAddres,
		cdnUrlSigner:           cdnUrlSigner,
	}
}

func (s *authServiceImpl) ParseTemplates() error {
	confirmationTemplate, err := template.ParseFiles("./templates/register_confirm.html")
	if err != nil {
		return err
	}
	verifiedTemplate, err := template.ParseFiles("./templates/verified_letter.html")
	if err != nil {
		return err
	}

	s.templates = NewTemplates(confirmationTemplate, verifiedTemplate)
	return nil
}

func (s *authServiceImpl) SetTemplates(confirmation *template.Template, verified *template.Template) {
	s.templates = NewTemplates(confirmation, verified)
}

func (s *authServiceImpl) RegisterCustomer(ctx context.Context, registerDto dto.RegisterUser) (dto.User, error) {
	var dto dto.User

	user, err := s.userRepository.CreateUser(ctx,
		repository.CreateUserParams{
			FirstName:   registerDto.FirstName,
			LastName:    registerDto.LastName,
			PhoneNumber: registerDto.PhoneNumber,
			Email:       registerDto.Email,
			Hash:        s.hasher.HashPassword(registerDto.Password),
			Role:        enum.UserRoleCUSTOMER,
		})
	if err != nil {
		return dto, err
	}

	dto, err = mapper.MapUserToDto(user, s.cdnUrlSigner)
	if err != nil {
		return dto, err
	}

	return dto, nil
}

func (s *authServiceImpl) RegisterProvider(ctx context.Context, registerDto dto.RegisterProvider) (dto.User, error) {
	var dto dto.User

	tx, err := s.pgx.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return dto, err
	}

	user, err := s.userRepository.WithTx(tx).CreateUser(ctx,
		repository.CreateUserParams{
			FirstName:   registerDto.FirstName,
			LastName:    registerDto.LastName,
			PhoneNumber: registerDto.PhoneNumber,
			Email:       registerDto.Email,
			Hash:        s.hasher.HashPassword(registerDto.Password),
			Role:        enum.UserRoleCUSTOMER,
		},
	)
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return dto, err
		}
		return dto, err
	}

	enc, err := s.encryptor.Encrypt([]byte(registerDto.PersonalIDNumber))
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return dto, err
		}
		return dto, err
	}

	err = s.providerRepository.WithTx(tx).Create(ctx, repository.CreateProviderParams{
		PersonalIDNumber:  enc,
		PersonalIDPreview: registerDto.PersonalIDNumber[len(registerDto.PersonalIDNumber)-5:],
		UserID:            user.ID,
	})
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return dto, err
		}
		return dto, err
	}

	if err := tx.Commit(ctx); err != nil {
		return dto, err
	}

	res, err := mapper.MapUserToDto(user, s.cdnUrlSigner)
	if err != nil {
		return dto, err
	}

	return res, nil
}

func (s *authServiceImpl) AuthenticateUser(ctx context.Context, loginDto dto.Login) (UserIdentity, error) {
	var dto UserIdentity
	creds, err := s.userRepository.GetCredentialsByEmail(ctx, loginDto.Email)
	if err != nil {
		return dto, err
	}

	err = s.hasher.VerifyPassword(loginDto.Password, creds.Hash)
	if err != nil {
		return dto, err
	}

	dto.ID = strconv.FormatInt(creds.ID, 10)
	dto.Role = string(creds.Role)
	dto.UserStatus = creds.UserStatus.Bool

	return dto, nil
}

func (s *authServiceImpl) VerifyUser(ctx context.Context, token string, ttl time.Duration, id int64, email string) error {
	err := s.verificationRepository.SetTokenUsed(ctx, token, ttl)
	if err != nil {
		return err
	}

	return s.userRepository.UpdateStatus(ctx, repository.UpdateUserStatusParams{
		ID:         id,
		UserStatus: pgtype.Bool{Bool: true, Valid: true},
	})
}

func (s *authServiceImpl) GetUserConfirmationDetails(ctx context.Context, email string) (UserConfirmationDetails, error) {
	var dto UserConfirmationDetails
	res, err := s.userRepository.GetUserConfirmationDetails(ctx, email)
	if err != nil {
		return dto, err
	}

	dto.ID = strconv.FormatInt(res.ID, 10)
	dto.UserStatus = res.UserStatus.Bool
	dto.Firstname = res.FirstName

	return dto, nil
}

func (s *authServiceImpl) GetUserRoleAndStatus(ctx context.Context, id int64) (dto UserRoleAndStatus, err error) {
	res, err := s.userRepository.GetUserRoleAndStatus(ctx, id)
	if err != nil {
		return dto, err
	}

	dto.Role = string(res.Role)
	dto.UserStatus = res.UserStatus.Bool

	return dto, err
}

func (s *authServiceImpl) SendConfirmationLetter(ctx context.Context, token string, email string, name string) error {
	isUsed, err := s.verificationRepository.IsTokenUsed(ctx, token)
	if err != nil {
		return err
	}
	if isUsed {
		return ErrAlreadyVerified
	}

	return s.mailer.SendHTML(
		s.emailAddres,
		email,
		"Account Confirmation",
		s.templates.confirmation,
		struct {
			Name  string
			Token string
		}{
			Name:  name,
			Token: token,
		},
	)
}

func (s *authServiceImpl) SendVerifiedLetter(email string) error {
	return s.mailer.SendHTML(
		s.emailAddres,
		email,
		"Verification Success",
		s.templates.verified,
		nil,
	)
}
