package service

import (
	"context"
	"errors"
	"io"
	"strings"

	"github.com/hexley21/fixup/internal/user/domain"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/infra/cdn"
	"github.com/hexley21/fixup/pkg/infra/s3"
	"github.com/jackc/pgx/v5"
)

// TODO: Add additional log messages to errors
// TODO: Remove reduntant names from structs

var directory = "pfp/"

type UserService interface {
	Get(ctx context.Context, userId int64) (*domain.User, error)
	UpdatePersonalInfo(ctx context.Context, id int64, personalInfo *domain.UserPersonalInfo) (*domain.UserPersonalInfo, error)
	UpdateProfilePicture(ctx context.Context, userId int64, file io.Reader, fileName string, fileSize int64, fileType string) error
	UpdatePassword(ctx context.Context, id int64, oldPassowrd string, newPassword string) error
	Delete(ctx context.Context, userId int64) error
}

type userServiceImpl struct {
	userRepository     repository.UserRepository
	s3Bucket           s3.Bucket
	cdnFileInvalidator cdn.FileInvalidator
	hasher             hasher.Hasher
}

func NewUserService(
	userRepository repository.UserRepository,
	s3Bucket s3.Bucket,
	cdnFileInvalidator cdn.FileInvalidator,
	hasher hasher.Hasher,
) *userServiceImpl {
	return &userServiceImpl{
		userRepository,
		s3Bucket,
		cdnFileInvalidator,
		hasher,
	}
}

func (s *userServiceImpl) Get(ctx context.Context, userId int64) (*domain.User, error) {
	userModel, err := s.userRepository.Get(ctx, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, err
	}

	return MapUserModelToEntity(userModel)
}

func (s *userServiceImpl) UpdatePersonalInfo(ctx context.Context, id int64, personalInfo *domain.UserPersonalInfo) (*domain.UserPersonalInfo, error) {
	user, err := s.userRepository.Update(ctx, id, repository.UpdateUserRow{
		FirstName:   personalInfo.FirstName,
		LastName:    personalInfo.LastName,
		Email:       personalInfo.Email,
		PhoneNumber: personalInfo.PhoneNumber,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		if errors.Is(err, repository.ErrInvalidUpdateParams) {
			return nil, ErrUserNotUpdated
		}

		return nil, err
	}

	return domain.NewUserPersonalInfo(user.Email, user.PhoneNumber, user.FirstName, user.LastName), nil
}

// UpdateProfilePicture uploads new picture to s3, updates record in db, deletes old one from s3 & invalidates cache
func (s *userServiceImpl) UpdateProfilePicture(ctx context.Context, userId int64, file io.Reader, fileName string, fileSize int64, fileType string) error {
	// fetch a picture in advance, also check if user exists
	picture, err := s.userRepository.GetPicture(ctx, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}

		return err
	}

	fileName, err = s.s3Bucket.PutObject(ctx, file, directory, fileName, fileSize, fileType)
	if err != nil {
		return err
	}

	var newPathBuilder strings.Builder
	newPathBuilder.WriteString(directory)
	newPathBuilder.WriteString(fileName)

	ok, err := s.userRepository.UpdatePicture(ctx, userId, newPathBuilder.String())
	if err != nil {
		return err
	}
	if !ok {
		return ErrUserNotUpdated
	}

	// if old picture fetched before invalid, skip the deletion from s3 and cache invalidation
	if picture.String == "" {
		return nil
	}
	err = s.s3Bucket.DeleteObject(ctx, picture.String)
	if err != nil {
		return err
	}

	return s.cdnFileInvalidator.InvalidateFile(ctx, picture.String)
}

func (s *userServiceImpl) UpdatePassword(ctx context.Context, id int64, oldPassowrd string, newPassword string) error {
	oldHash, err := s.userRepository.GetHashById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	err = s.hasher.VerifyPassword(oldPassowrd, oldHash)
	if err != nil {
		return err
	}

	hash, err := s.hasher.HashPassword(newPassword)
	if err != nil {
		if errors.Is(err, hasher.ErrPasswordMismatch) {
			return ErrIncorrectPassword
		}
		return err
	}

	ok, err := s.userRepository.UpdateHash(ctx, id, hash)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUserNotUpdated
	}

	return nil
}

func (s *userServiceImpl) Delete(ctx context.Context, userId int64) error {
	picture, err := s.userRepository.GetPicture(ctx, userId)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	if picture.String != "" {
		if err := s.s3Bucket.DeleteObject(ctx, picture.String); err != nil {
			return err
		}
		if err := s.cdnFileInvalidator.InvalidateFile(ctx, picture.String); err != nil {
			return err
		}
	}

	ok, err := s.userRepository.Delete(ctx, userId)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUserNotFound
	}

	return nil
}
