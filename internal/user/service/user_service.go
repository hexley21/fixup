package service

import (
	"context"
	"io"

	"github.com/hexley21/handy/internal/user/entity"
	"github.com/hexley21/handy/internal/user/repository"
	"github.com/hexley21/handy/pkg/infra/s3"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserService interface {
	FindUserById(ctx context.Context, userId int64) (entity.User, error)
	SetProfilePicture(ctx context.Context, userId int64, file io.Reader, fileName string, fileSize int64, fileType string) error
}

type userServiceImpl struct {
	userRepository repository.UserRepository
	s3Bucket       s3.Bucket
}

func NewUserService(userRepository repository.UserRepository, s3Bucket s3.Bucket) UserService {
	return &userServiceImpl{
		userRepository: userRepository,
		s3Bucket:       s3Bucket,
	}
}

func (s *userServiceImpl) FindUserById(ctx context.Context, userId int64) (entity.User, error) {
	return s.userRepository.GetById(ctx, userId)
}

func (s *userServiceImpl) SetProfilePicture(ctx context.Context, userId int64, file io.Reader, fileName string, fileSize int64, fileType string) error {
	user, err := s.FindUserById(ctx, userId)
	if err != nil {
		return err
	}

	fileName, err = s.s3Bucket.PutObject(ctx, file, fileName, fileSize, fileType)
	if err != nil {
		return err
	}

	var pictureName pgtype.Text
	if err := pictureName.Scan(fileName); err != nil {
		return err
	}

	err = s.userRepository.UpdateUserPicture(ctx, repository.UpdateUserPictureParams{
		ID:          userId,
		PictureName: pictureName,
	})
	if err != nil {
		return err
	}

	return s.s3Bucket.DeleteObject(ctx, user.PictureName.String)
}
