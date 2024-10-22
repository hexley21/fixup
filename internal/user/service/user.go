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

// TODO: move this to config
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

// Get retrieves a user by their userId from the repository.
// It returns ErrUserNotFound if the user does not exist, or the user entity and any other error encountered.
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

// UpdatePersonalInfo updates a user's personal information by their ID and returns domain user personal info.
// If the user is not found, it returns ErrUserNotFound.
// If the update parameters are invalid, it returns ErrUserNotUpdated.
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

// UpdateProfilePicture updates a user's profile picture by uploading a new picture to S3,
// updating the user's picture path in the repository, and deleting the old picture from S3.
// It also invalidates the old picture in the CDN if it exists. It returns an error if any step fails,
// including user not found, S3 upload failure, repository update failure, S3 deletion failure, or CDN invalidation failure.
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

	// if old picture fetched before is invalid, skip the deletion from s3 and cache invalidation
	if picture.String == "" {
		return nil
	}
	err = s.s3Bucket.DeleteObject(ctx, picture.String)
	if err != nil {
		return err
	}

	return s.cdnFileInvalidator.InvalidateFile(ctx, picture.String)
}

// UpdatePassword updates a user's password by verifying the old password and setting a new one.
// It retrieves the user's current password hash from the repository and verifies it against the provided old password.
// If the verification is successful, it hashes the new password and updates the repository with the new hash.
// It returns an error if the user is not found, the old password is incorrect, the new password hashing fails,
// or the repository update is unsuccessful.
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

// Delete removes a user by their userId. It first retrieves the user's picture from the repository.
// If the user has a picture, it deletes the picture from the S3 bucket and invalidates the file in the CDN.
// Finally, it deletes the user from the repository.
// If the user is not found, it returns ErrUserNotFound.
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
