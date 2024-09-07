package service

import (
	"context"
	"io"
	"strings"

	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto/mapper"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/hexley21/fixup/pkg/infra/cdn"
	"github.com/hexley21/fixup/pkg/infra/s3"
	"github.com/jackc/pgx/v5/pgtype"
)

var directory = "pfp/"

type UserService interface {
	FindUserById(ctx context.Context, userId int64) (dto.User, error)
	UpdateUserDataById(ctx context.Context, id int64, updateDto dto.UpdateUser) (dto.User, error)
	SetProfilePicture(ctx context.Context, userId int64, file io.Reader, fileName string, fileSize int64, fileType string) error
	DeleteUserById(ctx context.Context, userId int64) error
}

type userServiceImpl struct {
	userRepository     repository.UserRepository
	s3Bucket           s3.Bucket
	cdnFileInvalidator cdn.FileInvalidator
	cdnUrlSigner       cdn.URLSigner
}

func NewUserService(userRepository repository.UserRepository, s3Bucket s3.Bucket, cdnFileInvalidator cdn.FileInvalidator, cdnUrlSigner cdn.URLSigner) UserService {
	return &userServiceImpl{
		userRepository:     userRepository,
		s3Bucket:           s3Bucket,
		cdnFileInvalidator: cdnFileInvalidator,
		cdnUrlSigner:       cdnUrlSigner,
	}
}

func (s *userServiceImpl) FindUserById(ctx context.Context, userId int64) (dto.User, error) {
	var dto dto.User

	entity, err := s.userRepository.GetById(ctx, userId)
	if err != nil {
		return dto, err
	}

	dto, err = mapper.MapUserToDto(entity, s.cdnUrlSigner)
	if err != nil {
		return dto, err
	}

	return dto, nil
}

func (s *userServiceImpl) UpdateUserDataById(ctx context.Context, id int64, updateDto dto.UpdateUser) (dto.User, error) {
	var dto dto.User

	entity, err := s.userRepository.UpdateUserData(ctx, repository.UpdateUserDataParams{
		ID:          id,
		FirstName:   updateDto.FirstName,
		LastName:    updateDto.LastName,
		Email:       updateDto.Email,
		PhoneNumber: updateDto.PhoneNumber,
	})
	if err != nil {
		return dto, err
	}

	return mapper.MapUserToDto(entity, s.cdnUrlSigner)
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

	err = s.userRepository.UpdateUserPicture(ctx, repository.UpdateUserPictureParams{
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

func (s *userServiceImpl) DeleteUserById(ctx context.Context, userId int64) error {
	return s.userRepository.DeleteUserById(ctx, userId)
}