package service

import (
	"context"
	"io"
	"strings"

	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto/mapper"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/infra/cdn"
	"github.com/hexley21/fixup/pkg/infra/s3"
	"github.com/jackc/pgx/v5/pgtype"
)

var directory = "pfp/"

type UserService interface {
	FindUserById(ctx context.Context, userId int64) (dto.User, error)
	FindUserProfileById(ctx context.Context, userId int64) (dto.Profile, error)
	UpdateUserDataById(ctx context.Context, id int64, updateDTO dto.UpdateUser) (dto.User, error)
	SetProfilePicture(ctx context.Context, userId int64, file io.Reader, fileName string, fileSize int64, fileType string) error
	ChangePassword(ctx context.Context, id int64, updateDTO dto.UpdatePassword) error
	DeleteUserById(ctx context.Context, userId int64) error
}

type userServiceImpl struct {
	userRepository     repository.UserRepository
	s3Bucket           s3.Bucket
	cdnFileInvalidator cdn.FileInvalidator
	cdnUrlSigner       cdn.URLSigner
	hasher             hasher.Hasher
}

func NewUserService(
	userRepository repository.UserRepository,
	s3Bucket s3.Bucket,
	cdnFileInvalidator cdn.FileInvalidator,
	cdnUrlSigner cdn.URLSigner,
	hasher hasher.Hasher,
) *userServiceImpl {
	return &userServiceImpl{
		userRepository,
		s3Bucket,
		cdnFileInvalidator,
		cdnUrlSigner,
		hasher,
	}
}

func (s *userServiceImpl) FindUserById(ctx context.Context, userId int64) (dto.User, error) {
	var userDTO dto.User

	entity, err := s.userRepository.GetById(ctx, userId)
	if err != nil {
		return userDTO, err
	}

	userDTO, err = mapper.MapUserToDTO(entity, s.cdnUrlSigner)
	if err != nil {
		return userDTO, err
	}

	return userDTO, nil
}

func (s *userServiceImpl) FindUserProfileById(ctx context.Context, userId int64) (dto.Profile, error) {
	var profile dto.Profile

	entity, err := s.userRepository.GetById(ctx, userId)
	if err != nil {
		return profile, err
	}

	profile, err = mapper.MapUserToProfileDTO(entity, s.cdnUrlSigner)
	if err != nil {
		return profile, err
	}

	return profile, nil
}

func (s *userServiceImpl) UpdateUserDataById(ctx context.Context, id int64, updateDTO dto.UpdateUser) (dto.User, error) {
	var userDTO dto.User

	entity, err := s.userRepository.Update(ctx, repository.UpdateUserParams{
		ID:          id,
		FirstName:   updateDTO.FirstName,
		LastName:    updateDTO.LastName,
		Email:       updateDTO.Email,
		PhoneNumber: updateDTO.PhoneNumber,
	})
	if err != nil {
		return userDTO, err
	}

	return mapper.MapUserToDTO(entity, s.cdnUrlSigner)
}

func (s *userServiceImpl) SetProfilePicture(ctx context.Context, userId int64, file io.Reader, fileName string, fileSize int64, fileType string) error {
	entity, err := s.userRepository.GetById(ctx, userId)
	if err != nil {
		return err
	}

	fileName, err = s.s3Bucket.PutObject(ctx, file, directory, fileName, fileSize, fileType)
	if err != nil {
		return err
	}

	var newPathBuilder strings.Builder
	newPathBuilder.WriteString(directory)
	newPathBuilder.WriteString(fileName)

	var pictureName pgtype.Text
	if err := pictureName.Scan(newPathBuilder.String()); err != nil {
		return err
	}

	err = s.userRepository.UpdatePicture(ctx, repository.UpdateUserPictureParams{
		ID:          userId,
		PictureName: pictureName,
	})
	if err != nil {
		return err
	}

	if entity.PictureName.String == "" {
		return nil
	}

	err = s.s3Bucket.DeleteObject(ctx, entity.PictureName.String)
	if err != nil {
		return err
	}

	return s.cdnFileInvalidator.InvalidateFile(ctx, entity.PictureName.String)
}

func (s *userServiceImpl) ChangePassword(ctx context.Context, id int64, updateDTO dto.UpdatePassword) error {
	oldHash, err := s.userRepository.GetHashById(ctx, id)
	if err != nil {
		return err
	}

	err = s.hasher.VerifyPassword(updateDTO.OldPassword, oldHash)
	if err != nil {
		return err
	}

	hash, err := s.hasher.HashPassword(updateDTO.NewPassword)
	if err != nil {
		return err
	}

	return s.userRepository.UpdateHash(ctx, repository.UpdateUserHashParams{
		ID:   id,
		Hash: hash,
	})
}

func (s *userServiceImpl) DeleteUserById(ctx context.Context, userId int64) error {
	entity, err := s.userRepository.GetById(ctx, userId)
	if err != nil {
		return err
	}

	if entity.PictureName.String != "" {
		if err := s.s3Bucket.DeleteObject(ctx, entity.PictureName.String); err != nil {
			return err
		}
		if err := s.cdnFileInvalidator.InvalidateFile(ctx, entity.PictureName.String); err != nil {
			return err
		}
	}

	return s.userRepository.DeleteById(ctx, userId)
}
