package service

import (
	"context"
	"strconv"

	"github.com/hexley21/handy/internal/user/entity"
	"github.com/hexley21/handy/internal/user/repository"
)

type UserService interface {
	FindUserById(ctx context.Context, id string) (entity.User, error)
}

type userServiceImpl struct {
	userRepository repository.UserRepository
}

func NewUserService(userRepository repository.UserRepository) UserService {
	return &userServiceImpl{
		userRepository: userRepository,
	}
}

func (s *userServiceImpl) FindUserById(ctx context.Context, id string) (entity.User, error) {
	_id, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return entity.User{}, err
	}
	
	return s.userRepository.GetById(ctx, _id)
}
