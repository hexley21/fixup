package service_test

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/hexley21/fixup/internal/user/repository"
	mock_repository "github.com/hexley21/fixup/internal/user/repository/mock"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/pkg/hasher"
	mock_hasher "github.com/hexley21/fixup/pkg/hasher/mock"
	mock_cdn "github.com/hexley21/fixup/pkg/infra/cdn/mock"
	mock_s3 "github.com/hexley21/fixup/pkg/infra/s3/mock"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	userEntity = entity.User{
		ID:          1,
		FirstName:   "Larry",
		LastName:    "Page",
		PhoneNumber: "995111222333",
		Email:       "larry@page.com",
		PictureName: pgtype.Text{String: "larrypage.jpg", Valid: true},
		Hash:        "",
		Role:        enum.UserRoleADMIN,
		UserStatus:  pgtype.Bool{Bool: true, Valid: true},
		CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	userEntityWithoutPicture = entity.User{
		ID:          1,
		FirstName:   "Larry",
		LastName:    "Page",
		PhoneNumber: "995111222333",
		Email:       "larry@page.com",
		PictureName: pgtype.Text{String: "", Valid: false},
		Role:        enum.UserRoleADMIN,
		UserStatus:  pgtype.Bool{Bool: true, Valid: true},
		CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	signedPicture = "larrypage.jpg?signed=true"

	file           = bytes.NewReader([]byte("file"))
	fileName       = "file.jpg"
	fileType       = "image/jpeg"
	randomFilename = "zhGrapTRABowkxyhjqYjmybYbWWgZY1B"

	errSigningUrl          = errors.New("invalid URL, missing scheme and domain/path")
	errS3PutObject         = errors.New("failed to put object")
	errS3DeleteObject      = errors.New("failed to delete object")
	errCdnFileInvalidation = errors.New("failed to invaldate file")
	errUpdateError         = errors.New("failed to update row")
	errDeleteError         = errors.New("failed to delete row")
)

func TestFindUserById_Success(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockCdnUrlSigner := mock_cdn.NewMockURLSigner(ctrl)

	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	mockCdnUrlSigner.EXPECT().SignURL(userEntity.PictureName.String).Return(signedPicture, nil)

	service := service.NewUserService(mockUserRepo, nil, nil, mockCdnUrlSigner, nil)
	dto, err := service.FindUserById(ctx, userEntity.ID)

	assert.NoError(t, err)

	assert.Equal(t, strconv.FormatInt(userEntity.ID, 10), dto.ID)
	assert.Equal(t, userEntity.FirstName, dto.FirstName)
	assert.Equal(t, userEntity.LastName, dto.LastName)
	assert.Equal(t, userEntity.PhoneNumber, dto.PhoneNumber)
	assert.Equal(t, userEntity.Email, dto.Email)
	assert.Equal(t, signedPicture, dto.PictureUrl)
	assert.Equal(t, string(userEntity.Role), dto.Role)
	assert.Equal(t, userEntity.UserStatus.Bool, dto.UserStatus)
	assert.Equal(t, userEntity.CreatedAt.Time, dto.CreatedAt)
}

func TestFindUserById_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(entity.User{}, pgx.ErrNoRows)

	service := service.NewUserService(mockUserRepo, nil, nil, nil, nil)
	dto, err := service.FindUserById(ctx, userEntity.ID)

	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, dto)
}

func TestFindUserById_MapperError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockCdnUrlSigner := mock_cdn.NewMockURLSigner(ctrl)

	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	mockCdnUrlSigner.EXPECT().SignURL(userEntity.PictureName.String).Return(signedPicture, errSigningUrl)

	service := service.NewUserService(mockUserRepo, nil, nil, mockCdnUrlSigner, nil)
	dto, err := service.FindUserById(ctx, userEntity.ID)

	assert.ErrorIs(t, err, errSigningUrl)
	assert.Empty(t, dto)
}

func TestFindUserProfileById_Success(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockCdnUrlSigner := mock_cdn.NewMockURLSigner(ctrl)

	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	mockCdnUrlSigner.EXPECT().SignURL(userEntity.PictureName.String).Return(signedPicture, nil)

	service := service.NewUserService(mockUserRepo, nil, nil, mockCdnUrlSigner, nil)
	dto, err := service.FindUserProfileById(ctx, userEntity.ID)

	assert.NoError(t, err)

	assert.Equal(t, strconv.FormatInt(userEntity.ID, 10), dto.ID)
	assert.Equal(t, userEntity.FirstName, dto.FirstName)
	assert.Equal(t, userEntity.LastName, dto.LastName)
	assert.Equal(t, signedPicture, dto.PictureUrl)
	assert.Equal(t, string(userEntity.Role), dto.Role)
	assert.Equal(t, userEntity.UserStatus.Bool, dto.UserStatus)
	assert.Equal(t, userEntity.CreatedAt.Time, dto.CreatedAt)
}

func TestFindUserProfileById_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(entity.User{}, pgx.ErrNoRows)

	service := service.NewUserService(mockUserRepo, nil, nil, nil, nil)
	dto, err := service.FindUserProfileById(ctx, userEntity.ID)

	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, dto)
}

func TestFindUserProfileById_MapperError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockCdnUrlSigner := mock_cdn.NewMockURLSigner(ctrl)

	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	mockCdnUrlSigner.EXPECT().SignURL(userEntity.PictureName.String).Return(signedPicture, errSigningUrl)

	service := service.NewUserService(mockUserRepo, nil, nil, mockCdnUrlSigner, nil)
	dto, err := service.FindUserProfileById(ctx, userEntity.ID)

	assert.ErrorIs(t, err, errSigningUrl)
	assert.Empty(t, dto)
}

func TestUpdateUserDataById_Success(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockCdnUrlSigner := mock_cdn.NewMockURLSigner(ctrl)

	updateUserDto := dto.UpdateUser{Email: &userEntity.Email, PhoneNumber: &userEntity.PhoneNumber, FirstName: &userEntity.FirstName, LastName: &userEntity.LastName}
	updateUserParams := repository.UpdateUserParams{ID: userEntity.ID, FirstName: updateUserDto.FirstName, LastName: updateUserDto.LastName, PhoneNumber: updateUserDto.PhoneNumber, Email: updateUserDto.Email}

	mockUserRepo.EXPECT().Update(ctx, updateUserParams).Return(userEntity, nil)
	mockCdnUrlSigner.EXPECT().SignURL(userEntity.PictureName.String).Return(signedPicture, nil)

	service := service.NewUserService(mockUserRepo, nil, nil, mockCdnUrlSigner, nil)
	dto, err := service.UpdateUserDataById(ctx, userEntity.ID, updateUserDto)

	assert.NoError(t, err)

	assert.Equal(t, strconv.FormatInt(userEntity.ID, 10), dto.ID)
	assert.Equal(t, userEntity.Email, dto.Email)
	assert.Equal(t, userEntity.PhoneNumber, dto.PhoneNumber)
	assert.Equal(t, userEntity.FirstName, dto.FirstName)
	assert.Equal(t, userEntity.LastName, dto.LastName)
}

func TestUpdateUserDataById_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)

	updateUserDto := dto.UpdateUser{Email: &userEntity.Email, PhoneNumber: &userEntity.PhoneNumber, FirstName: &userEntity.FirstName, LastName: &userEntity.LastName}
	updateUserParams := repository.UpdateUserParams{ID: userEntity.ID, FirstName: updateUserDto.FirstName, LastName: updateUserDto.LastName, PhoneNumber: updateUserDto.PhoneNumber, Email: updateUserDto.Email}

	mockUserRepo.EXPECT().Update(ctx, updateUserParams).Return(entity.User{}, pgx.ErrNoRows)

	service := service.NewUserService(mockUserRepo, nil, nil, nil, nil)
	dto, err := service.UpdateUserDataById(ctx, userEntity.ID, updateUserDto)

	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, dto)
}

func TestUpdateUserDataById_SignerError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockCdnUrlSigner := mock_cdn.NewMockURLSigner(ctrl)

	updateUserDto := dto.UpdateUser{Email: &userEntity.Email, PhoneNumber: &userEntity.PhoneNumber, FirstName: &userEntity.FirstName, LastName: &userEntity.LastName}
	updateUserParams := repository.UpdateUserParams{ID: userEntity.ID, FirstName: updateUserDto.FirstName, LastName: updateUserDto.LastName, PhoneNumber: updateUserDto.PhoneNumber, Email: updateUserDto.Email}

	mockUserRepo.EXPECT().Update(ctx, updateUserParams).Return(userEntity, nil)
	mockCdnUrlSigner.EXPECT().SignURL(userEntity.PictureName.String).Return("", errSigningUrl)

	service := service.NewUserService(mockUserRepo, nil, nil, mockCdnUrlSigner, nil)
	dto, err := service.UpdateUserDataById(ctx, userEntity.ID, updateUserDto)

	assert.ErrorIs(t, err, errSigningUrl)
	assert.Empty(t, dto)
}

func TestSetProfilePicture_Success(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockS3Bucket := mock_s3.NewMockBucket(ctrl)
	mockFileInvalidator := mock_cdn.NewMockFileInvalidator(ctrl)

	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	mockUserRepo.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(nil)
	mockS3Bucket.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)
	mockS3Bucket.EXPECT().DeleteObject(ctx, userEntity.PictureName.String).Return(nil)
	mockFileInvalidator.EXPECT().InvalidateFile(ctx, userEntity.PictureName.String).Return(nil)

	service := service.NewUserService(mockUserRepo, mockS3Bucket, mockFileInvalidator, nil, nil)

	err := service.SetProfilePicture(ctx, userEntity.ID, file, fileName, file.Size(), fileType)
	assert.NoError(t, err)
}

func TestSetProfilePicture_WithoutPicture(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockS3Bucket := mock_s3.NewMockBucket(ctrl)

	mockUserRepo.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(nil)
	mockUserRepo.EXPECT().GetById(ctx, userEntityWithoutPicture.ID).Return(userEntityWithoutPicture, nil)
	mockS3Bucket.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)

	service := service.NewUserService(mockUserRepo, mockS3Bucket, nil, nil, nil)

	err := service.SetProfilePicture(ctx, userEntityWithoutPicture.ID, file, fileName, file.Size(), fileType)
	assert.NoError(t, err)
}

func TestSetProfilePicture_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, pgx.ErrNoRows)

	service := service.NewUserService(mockUserRepo, nil, nil, nil, nil)

	err := service.SetProfilePicture(ctx, userEntity.ID, file, fileName, file.Size(), fileType)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestSetProfilePicture_PutObjectError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockS3Bucket := mock_s3.NewMockBucket(ctrl)

	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	mockS3Bucket.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return("", errS3PutObject)

	service := service.NewUserService(mockUserRepo, mockS3Bucket, nil, nil, nil)

	err := service.SetProfilePicture(ctx, userEntity.ID, file, fileName, file.Size(), fileType)
	assert.ErrorIs(t, err, errS3PutObject)
}

func TestSetProfilePicture_UpdatePictureError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockS3Bucket := mock_s3.NewMockBucket(ctrl)

	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	mockUserRepo.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(pgx.ErrNoRows)
	mockS3Bucket.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)

	service := service.NewUserService(mockUserRepo, mockS3Bucket, nil, nil, nil)

	err := service.SetProfilePicture(ctx, userEntity.ID, file, fileName, file.Size(), fileType)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestSetProfilePicture_DeleteObjectError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockS3Bucket := mock_s3.NewMockBucket(ctrl)

	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	mockUserRepo.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(nil)
	mockS3Bucket.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)
	mockS3Bucket.EXPECT().DeleteObject(ctx, userEntity.PictureName.String).Return(errS3DeleteObject)

	service := service.NewUserService(mockUserRepo, mockS3Bucket, nil, nil, nil)

	err := service.SetProfilePicture(ctx, userEntity.ID, file, fileName, file.Size(), fileType)
	assert.ErrorIs(t, err, errS3DeleteObject)
}

func TestSetProfilePicture_InvalidationError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockS3Bucket := mock_s3.NewMockBucket(ctrl)
	mockFileInvalidator := mock_cdn.NewMockFileInvalidator(ctrl)

	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	mockUserRepo.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(nil)
	mockS3Bucket.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)
	mockS3Bucket.EXPECT().DeleteObject(ctx, userEntity.PictureName.String).Return(nil)
	mockFileInvalidator.EXPECT().InvalidateFile(ctx, userEntity.PictureName.String).Return(errCdnFileInvalidation)

	service := service.NewUserService(mockUserRepo, mockS3Bucket, mockFileInvalidator, nil, nil)

	err := service.SetProfilePicture(ctx, userEntity.ID, file, fileName, file.Size(), fileType)
	assert.ErrorIs(t, err, errCdnFileInvalidation)
}

func TestChangePassword_Success(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockHasher := mock_hasher.NewMockHasher(ctrl)

	mockUserRepo.EXPECT().GetHashById(ctx, userEntity.ID).Return(userEntity.Hash, nil)
	mockUserRepo.EXPECT().UpdateHash(ctx, gomock.Any()).Return(nil)
	mockHasher.EXPECT().VerifyPassword(gomock.Any(), userEntity.Hash).Return(nil)
	mockHasher.EXPECT().HashPassword(gomock.Any()).Return("")

	service := service.NewUserService(mockUserRepo, nil, nil, nil, mockHasher)
	assert.NoError(t, service.ChangePassword(ctx, userEntity.ID, dto.UpdatePassword{}))
}

func TestChangePassword_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetHashById(ctx, userEntity.ID).Return("", pgx.ErrNoRows)

	service := service.NewUserService(mockUserRepo, nil, nil, nil, nil)
	assert.ErrorIs(t, service.ChangePassword(ctx, userEntity.ID, dto.UpdatePassword{}), pgx.ErrNoRows)
}

func TestChangePassword_IncorrectPassword(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockHasher := mock_hasher.NewMockHasher(ctrl)

	mockUserRepo.EXPECT().GetHashById(ctx, userEntity.ID).Return(userEntity.Hash, nil)
	mockHasher.EXPECT().VerifyPassword(gomock.Any(), userEntity.Hash).Return(hasher.ErrPasswordMismatch)

	service := service.NewUserService(mockUserRepo, nil, nil, nil, mockHasher)
	assert.ErrorIs(t, service.ChangePassword(ctx, userEntity.ID, dto.UpdatePassword{}), hasher.ErrPasswordMismatch)
}

func TestChangePassword_UpdateError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockHasher := mock_hasher.NewMockHasher(ctrl)

	mockUserRepo.EXPECT().GetHashById(ctx, userEntity.ID).Return(userEntity.Hash, nil)
	mockUserRepo.EXPECT().UpdateHash(ctx, gomock.Any()).Return(errUpdateError)
	mockHasher.EXPECT().VerifyPassword(gomock.Any(), userEntity.Hash).Return(nil)
	mockHasher.EXPECT().HashPassword(gomock.Any()).Return("")

	service := service.NewUserService(mockUserRepo, nil, nil, nil, mockHasher)
	assert.ErrorIs(t, service.ChangePassword(ctx, userEntity.ID, dto.UpdatePassword{}), errUpdateError)
}

func TestDeleteUserById_Success(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockS3Bucket := mock_s3.NewMockBucket(ctrl)
	mockFileInvalidator := mock_cdn.NewMockFileInvalidator(ctrl)

	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	mockUserRepo.EXPECT().DeleteById(ctx, userEntity.ID).Return(nil)
	mockS3Bucket.EXPECT().DeleteObject(ctx, userEntity.PictureName.String).Return(nil)
	mockFileInvalidator.EXPECT().InvalidateFile(ctx, userEntity.PictureName.String).Return(nil)

	service := service.NewUserService(mockUserRepo, mockS3Bucket, mockFileInvalidator, nil, nil)
	assert.NoError(t, service.DeleteUserById(ctx, userEntity.ID))
}

func TestDeleteUserById_EmptyPicture(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetById(ctx, userEntityWithoutPicture.ID).Return(userEntityWithoutPicture, nil)
	mockUserRepo.EXPECT().DeleteById(ctx, userEntityWithoutPicture.ID).Return(nil)

	service := service.NewUserService(mockUserRepo, nil, nil, nil, nil)
	assert.NoError(t, service.DeleteUserById(ctx, userEntity.ID))
}

func TestDeleteUserById_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(entity.User{}, pgx.ErrNoRows)

	service := service.NewUserService(mockUserRepo, nil, nil, nil, nil)
	assert.ErrorIs(t, service.DeleteUserById(ctx, userEntity.ID), pgx.ErrNoRows)
}

func TestDeleteUserById_DeleteObjectError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockS3Bucket := mock_s3.NewMockBucket(ctrl)

	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	mockS3Bucket.EXPECT().DeleteObject(ctx, userEntity.PictureName.String).Return(errS3DeleteObject)

	service := service.NewUserService(mockUserRepo, mockS3Bucket, nil, nil, nil)
	assert.ErrorIs(t, service.DeleteUserById(ctx, userEntity.ID), errS3DeleteObject)
}

func TestDeleteUserById_InvalidationError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockS3Bucket := mock_s3.NewMockBucket(ctrl)
	mockFileInvalidator := mock_cdn.NewMockFileInvalidator(ctrl)

	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	mockS3Bucket.EXPECT().DeleteObject(ctx, userEntity.PictureName.String).Return(nil)
	mockFileInvalidator.EXPECT().InvalidateFile(ctx, userEntity.PictureName.String).Return(errCdnFileInvalidation)

	service := service.NewUserService(mockUserRepo, mockS3Bucket, mockFileInvalidator, nil, nil)
	assert.ErrorIs(t, service.DeleteUserById(ctx, userEntity.ID), errCdnFileInvalidation)
}

func TestDeleteUserById_RowDeletionError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockS3Bucket := mock_s3.NewMockBucket(ctrl)
	mockFileInvalidator := mock_cdn.NewMockFileInvalidator(ctrl)

	mockUserRepo.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	mockUserRepo.EXPECT().DeleteById(ctx, userEntity.ID).Return(errDeleteError)
	mockS3Bucket.EXPECT().DeleteObject(ctx, userEntity.PictureName.String).Return(nil)
	mockFileInvalidator.EXPECT().InvalidateFile(ctx, userEntity.PictureName.String).Return(nil)

	service := service.NewUserService(mockUserRepo, mockS3Bucket, mockFileInvalidator, nil, nil)
	assert.ErrorIs(t, service.DeleteUserById(ctx, userEntity.ID), errDeleteError)
}
