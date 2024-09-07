package service

import (
	"context"
	"html/template"

	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto/mapper"
	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/hexley21/fixup/pkg/encryption"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/infra/cdn"
	"github.com/hexley21/fixup/pkg/mailer"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthService interface {
	RegisterCustomer(ctx context.Context, registerDto dto.RegisterUser) (dto.User, error)
	RegisterProvider(ctx context.Context, registerDto dto.RegisterProvider) (dto.User, error)
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
) AuthService {
	return &authServiceImpl{
		userRepository:     userRepository,
		providerRepository: providerRepository,
		dbPool:             dbPool,
		hasher:             hasher,
		encryptor:          encryptor,
		mailer:             mailer,
		emailAddres:        emailAddres,
		cdnUrlSigner:       cdnUrlSigner,
	}
}

func (s *authServiceImpl) sendConfirmationEmail(email string, name string, link string) error {
	t, err := template.ParseFiles("./templates/register_confirm.html")
	if err != nil {
		return err
	}

	return s.mailer.SendHTML(
		s.emailAddres,
		email,
		"Account Confirmation",
		t,
		struct {
			Name string
			Link string
		}{Name: name, Link: link},
	)
}

func (s *authServiceImpl) registerUser(ctx context.Context, dto dto.RegisterUser, tx pgx.Tx) (entity.User, error) {
	var user entity.User
	user, err := s.userRepository.WithTx(tx).CreateUser(ctx, repository.CreateUserParams{
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

	if err := s.sendConfirmationEmail(user.Email, user.FirstName, user.LastName); err != nil {
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

	if err := s.sendConfirmationEmail(user.Email, user.FirstName, user.LastName); err != nil {
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
