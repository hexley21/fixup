package service

import (
	"context"
	"errors"
	"html/template"
	"strconv"
	"time"

	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto/mapper"
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
	RegisterCustomer(ctx context.Context, registerDTO dto.RegisterUser) (dto.User, error)
	RegisterProvider(ctx context.Context, registerDTO dto.RegisterProvider) (dto.User, error)
	AuthenticateUser(ctx context.Context, loginDTO dto.Login) (UserIdentity, error)
	VerifyUser(ctx context.Context, token string, ttl time.Duration, id int64) error
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

func (s *authServiceImpl) RegisterCustomer(ctx context.Context, registerDTO dto.RegisterUser) (dto.User, error) {
	var userDTO dto.User

	hash, err := s.hasher.HashPassword(registerDTO.Password)
	if err != nil {
		return userDTO, err
	}

	user, err := s.userRepository.CreateUser(ctx,
		repository.CreateUserParams{
			FirstName:   registerDTO.FirstName,
			LastName:    registerDTO.LastName,
			PhoneNumber: registerDTO.PhoneNumber,
			Email:       registerDTO.Email,
			Hash:        hash,
			Role:        enum.UserRoleCUSTOMER,
		})
	if err != nil {
		return userDTO, err
	}

	userDTO, err = mapper.MapUserToDTO(user, s.cdnUrlSigner)
	if err != nil {
		return userDTO, err
	}

	return userDTO, nil
}

func (s *authServiceImpl) RegisterProvider(ctx context.Context, registerDTO dto.RegisterProvider) (dto.User, error) {
	var userDTO dto.User

	hash, err := s.hasher.HashPassword(registerDTO.Password)
	if err != nil {
		return userDTO, err
	}

	tx, err := s.pgx.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return userDTO, err
	}

	user, err := s.userRepository.WithTx(tx).CreateUser(ctx,
		repository.CreateUserParams{
			FirstName:   registerDTO.FirstName,
			LastName:    registerDTO.LastName,
			PhoneNumber: registerDTO.PhoneNumber,
			Email:       registerDTO.Email,
			Hash:        hash,
			Role:        enum.UserRoleCUSTOMER,
		},
	)
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return userDTO, err
		}
		return userDTO, err
	}

	enc, err := s.encryptor.Encrypt([]byte(registerDTO.PersonalIDNumber))
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return userDTO, err
		}
		return userDTO, err
	}

	err = s.providerRepository.WithTx(tx).Create(ctx, repository.CreateProviderParams{
		PersonalIDNumber:  enc,
		PersonalIDPreview: registerDTO.PersonalIDNumber[len(registerDTO.PersonalIDNumber)-5:],
		UserID:            user.ID,
	})
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return userDTO, err
		}
		return userDTO, err
	}

	if err := tx.Commit(ctx); err != nil {
		return userDTO, err
	}

	res, err := mapper.MapUserToDTO(user, s.cdnUrlSigner)
	if err != nil {
		return userDTO, err
	}

	return res, nil
}

func (s *authServiceImpl) AuthenticateUser(ctx context.Context, loginDTO dto.Login) (UserIdentity, error) {
	var identityDTO UserIdentity
	creds, err := s.userRepository.GetCredentialsByEmail(ctx, loginDTO.Email)
	if err != nil {
		return identityDTO, err
	}

	err = s.hasher.VerifyPassword(loginDTO.Password, creds.Hash)
	if err != nil {
		return identityDTO, err
	}

	identityDTO.ID = strconv.FormatInt(creds.ID, 10)
	identityDTO.Role = string(creds.Role)
	identityDTO.UserStatus = creds.UserStatus.Bool

	return identityDTO, nil
}

func (s *authServiceImpl) VerifyUser(ctx context.Context, token string, ttl time.Duration, id int64) error {
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
	var detailsDTO UserConfirmationDetails
	res, err := s.userRepository.GetUserConfirmationDetails(ctx, email)
	if err != nil {
		return detailsDTO, err
	}

	detailsDTO.ID = strconv.FormatInt(res.ID, 10)
	detailsDTO.UserStatus = res.UserStatus.Bool
	detailsDTO.Firstname = res.FirstName

	return detailsDTO, nil
}

func (s *authServiceImpl) GetUserRoleAndStatus(ctx context.Context, id int64) (UserRoleAndStatus, error) {
	var roleAndStatus UserRoleAndStatus

	res, err := s.userRepository.GetUserRoleAndStatus(ctx, id)
	if err != nil {
		return roleAndStatus, err
	}

	roleAndStatus.Role = string(res.Role)
	roleAndStatus.UserStatus = res.UserStatus.Bool

	return roleAndStatus, err
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
		s.emailAddress,
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
		s.emailAddress,
		email,
		"Verification Success",
		s.templates.verified,
		nil,
	)
}
