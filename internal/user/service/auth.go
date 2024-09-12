package service

import (
	"context"
	"html/template"
	"strconv"

	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto/mapper"
	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/hexley21/fixup/internal/user/service/verifier"
	"github.com/hexley21/fixup/pkg/encryption"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/infra/cdn"
	"github.com/hexley21/fixup/pkg/mailer"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type templates struct {
	confirmation *template.Template
	verified     *template.Template
}

func newTemplates(
	confirmationPath string,
	verifiedPath string,
) (*templates, error) {
	confirmationTemplate, err := template.ParseFiles(confirmationPath)
	if err != nil {
		return nil, err
	}
	verifiedTemplate, err := template.ParseFiles(verifiedPath)
	if err != nil {
		return nil, err
	}
	return &templates{
		confirmation: confirmationTemplate,
		verified:     verifiedTemplate,
	}, nil
}

type AuthService interface {
	RegisterCustomer(ctx context.Context, registerDto dto.RegisterUser) (dto.User, error)
	RegisterProvider(ctx context.Context, registerDto dto.RegisterProvider) (dto.User, error)
	AuthenticateUser(ctx context.Context, loginDto dto.Login) (dto.User, error)
	VerifyUser(ctx context.Context, id int64, email string) error
}

type authServiceImpl struct {
	userRepository     repository.UserRepository
	providerRepository repository.ProviderRepository
	dbPool             *pgxpool.Pool
	hasher             hasher.Hasher
	encryptor          encryption.Encryptor
	mailer             mailer.Mailer
	cdnUrlSigner       cdn.URLSigner
	emailAddres        string
	templates          *templates
	jwtGenerator       verifier.JwtGenerator
}

func NewAuthService(
	userRepository repository.UserRepository,
	providerRepository repository.ProviderRepository,
	dbPool *pgxpool.Pool,
	hasher hasher.Hasher,
	encryptor encryption.Encryptor,
	mailer mailer.Mailer,
	emailAddres string,
	cdnUrlSigner cdn.URLSigner,
	jwtGenerator verifier.JwtGenerator,
) (AuthService, error) {
	templates, err := newTemplates(
		"./templates/register_confirm.html",
		"./templates/verified_letter.html",
	)
	if err != nil {
		return nil, err
	}

	return &authServiceImpl{
		userRepository:     userRepository,
		providerRepository: providerRepository,
		dbPool:             dbPool,
		hasher:             hasher,
		encryptor:          encryptor,
		mailer:             mailer,
		emailAddres:        emailAddres,
		cdnUrlSigner:       cdnUrlSigner,
		templates:          templates,
		jwtGenerator:       jwtGenerator,
	}, nil
}

func (s *authServiceImpl) sendConfirmationEmail(id int64, email string, name string) error {
	verificationJwt, err := s.jwtGenerator.GenerateToken(strconv.FormatInt(id, 10), email)
	if err != nil {
		return err
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
			Token: verificationJwt,
		},
	)
}

func (s *authServiceImpl) sendVerifiedLetter(email string) error {
	return s.mailer.SendHTML(
		s.emailAddres,
		email,
		"Verification Success",
		s.templates.verified,
		nil,
	)
}

func (s *authServiceImpl) registerUser(ctx context.Context, dto dto.RegisterUser, tx pgx.Tx) (entity.User, error) {
	var user entity.User
	user, err := s.userRepository.WithTx(tx).Create(ctx,
		repository.CreateParams{
			FirstName:   dto.FirstName,
			LastName:    dto.LastName,
			PhoneNumber: dto.PhoneNumber,
			Email:       dto.Email,
			Hash:        s.hasher.HashPassword(dto.Password),
			Role:        enum.UserRoleCUSTOMER,
		})
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return user, err
		}
		return user, err
	}

	return user, nil
}

func (s *authServiceImpl) RegisterCustomer(ctx context.Context, registerDto dto.RegisterUser) (dto.User, error) {
	var dto dto.User

	tx, err := s.dbPool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return dto, err
	}

	user, err := s.registerUser(ctx, registerDto, tx)
	if err != nil {
		return dto, err
	}

	if err := s.sendConfirmationEmail(user.ID, user.Email, user.FirstName); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return dto, err
		}
		return dto, err
	}

	if err := tx.Commit(ctx); err != nil {
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

	tx, err := s.dbPool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return dto, err
	}

	user, err := s.registerUser(ctx, registerDto.RegisterUser, tx)
	if err != nil {
		return dto, err
	}

	enc, err := s.encryptor.Encrypt([]byte(registerDto.PersonalIDNumber))
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return dto, err
		}
		return dto, err
	}

	err = s.providerRepository.WithTx(tx).CreateProvider(ctx, repository.CreateProviderParams{
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

	if err := s.sendConfirmationEmail(user.ID, user.Email, user.FirstName); err != nil {
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

func (s *authServiceImpl) AuthenticateUser(ctx context.Context, loginDto dto.Login) (dto.User, error) {
	var dto dto.User
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

	return dto, nil
}

func (s *authServiceImpl) VerifyUser(ctx context.Context, id int64, email string) error {
	var status pgtype.Bool
	status.Scan(true)

	err := s.userRepository.UpdateStatus(ctx, repository.UpdateStatusParams{
		ID:         id,
		UserStatus: status,
	})
	if err != nil {
		return err
	}

	go func() {
		s.sendVerifiedLetter(email)
	}()

	return nil
}