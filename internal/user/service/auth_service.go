package service

import (
	"context"
	"html/template"

	"github.com/hexley21/handy/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/handy/internal/user/delivery/http/v1/dto/mapper"
	"github.com/hexley21/handy/internal/user/entity"
	"github.com/hexley21/handy/internal/user/enum"
	"github.com/hexley21/handy/internal/user/repository"
	"github.com/hexley21/handy/pkg/encryption"
	"github.com/hexley21/handy/pkg/hasher"
	"github.com/hexley21/handy/pkg/mailer"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthService interface {
	RegisterCustomer(ctx context.Context, dto *dto.RegisterUser) (*dto.User, error)
	RegisterProvider(ctx context.Context, dto *dto.RegisterProvider) (*dto.User, error)
}

type authServiceImpl struct {
	userRepository     repository.UserRepository
	providerRepository repository.ProviderRepository
	dbPool             *pgxpool.Pool
	hasher             hasher.Hasher
	encryptor          encryption.Encryptor
	mailer             mailer.Mailer
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
) AuthService {
	return &authServiceImpl{
		userRepository:     userRepository,
		providerRepository: providerRepository,
		dbPool:             dbPool,
		hasher:             hasher,
		encryptor:          encryptor,
		mailer:             mailer,
		emailAddres:        emailAddres,
	}
}

func (s *authServiceImpl) sendConfirmationEmail(email string, name string, link string) error {
	t, err := template.ParseFiles("./templates/register_confirm.templ")
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

func (s *authServiceImpl) registerUser(ctx context.Context, dto *dto.RegisterUser, tx pgx.Tx) (*entity.User, error) {
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
			return nil, err
		}
		return nil, err
	}

	return &user, nil
}

func (s *authServiceImpl) RegisterCustomer(ctx context.Context, dto *dto.RegisterUser) (*dto.User, error) {
	tx, err := s.dbPool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}

	user, err := s.registerUser(ctx, dto, tx)
	if err != nil {
		return nil, err
	}

	if err := s.sendConfirmationEmail(user.Email, user.FirstName, user.LastName); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, err
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	res, err := mapper.MapUserToDto(user)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s *authServiceImpl) RegisterProvider(ctx context.Context, dto *dto.RegisterProvider) (*dto.User, error) {
	tx, err := s.dbPool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}

	user, err := s.registerUser(ctx, &dto.RegisterUser, tx)
	if err != nil {
		return nil, err
	}

	enc, err := s.encryptor.Encrypt([]byte(dto.PersonalIDNumber))
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, err
		}
		return nil, err
	}

	err = s.providerRepository.WithTx(tx).CreateProvider(ctx, repository.CreateProviderParams{
		PersonalIDNumber:  enc,
		PersonalIDPreview: dto.PersonalIDNumber[len(dto.PersonalIDNumber)-5:],
		UserID:            user.ID,
	})
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, err
		}
		return nil, err
	}

	if err := s.sendConfirmationEmail(user.Email, user.FirstName, user.LastName); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, err
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	res, err := mapper.MapUserToDto(user)
	if err != nil {
		return nil, err
	}

	return res, nil
}
