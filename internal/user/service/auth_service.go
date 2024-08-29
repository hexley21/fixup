package service

import (
	"context"

	"github.com/hexley21/handy/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/handy/internal/user/delivery/http/v1/dto/mapper"
	"github.com/hexley21/handy/internal/user/enum"
	"github.com/hexley21/handy/internal/user/repository"
	"github.com/hexley21/handy/internal/user/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthService interface {
	RegisterCustomer(ctx context.Context, dto *dto.RegisterUser) (*dto.User, error)
}

type authServiceImpl struct {
	userRepository repository.UserRepository
	emailService   EmailService
	dbPool         *pgxpool.Pool
	hasher         util.Hasher
}

func NewAuthService(userRepository repository.UserRepository, emailService EmailService, dbPool *pgxpool.Pool, hasher util.Hasher) AuthService {
	return &authServiceImpl{
		userRepository: userRepository,
		emailService:   emailService,
		dbPool:         dbPool,
		hasher:         hasher,
	}
}

func (s *authServiceImpl) RegisterCustomer(ctx context.Context, dto *dto.RegisterUser) (*dto.User, error) {
	tx, err := s.dbPool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return nil, err
	}

	user, err := s.userRepository.WithTx(tx).CreateUser(ctx, repository.CreateUserParams{
		FirstName:   dto.FirstName,
		LastName:    dto.LastName,
		PhoneNumber: dto.PhoneNumber,
		Email:       dto.Email,
		Hash:        s.hasher.HashPassword(dto.Password),
		Role:        enum.UserRoleCUSTOMER,
	})
	if err != nil {
		return nil, err
	}

	if err := s.emailService.SendConfirmation(ctx, user.Email); err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return nil, err
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	res, err := mapper.MapUserToDto(&user)
	if err != nil {
		return nil, err
	}
	

	return res, nil
}
