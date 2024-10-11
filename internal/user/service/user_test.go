package service_test

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/internal/user/repository"
	mockRepository "github.com/hexley21/fixup/internal/user/repository/mock"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/pkg/hasher"
	mockHasher "github.com/hexley21/fixup/pkg/hasher/mock"
	mockCdn "github.com/hexley21/fixup/pkg/infra/cdn/mock"
	mockS3 "github.com/hexley21/fixup/pkg/infra/s3/mock"
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

func setupUser(t *testing.T) (
	ctx context.Context,
	ctrl *gomock.Controller,
	svc service.UserService,
	userRepoMock *mockRepository.MockUserRepository,
	cdnUrlSignerMock *mockCdn.MockURLSigner,
	s3BucketMock *mockS3.MockBucket,
	fileInvalidatorMock *mockCdn.MockFileInvalidator,
	hasherMock *mockHasher.MockHasher,
) {
	ctx = context.Background()
	ctrl = gomock.NewController(t)
	userRepoMock = mockRepository.NewMockUserRepository(ctrl)
	cdnUrlSignerMock = mockCdn.NewMockURLSigner(ctrl)
	s3BucketMock = mockS3.NewMockBucket(ctrl)
	fileInvalidatorMock = mockCdn.NewMockFileInvalidator(ctrl)
	hasherMock = mockHasher.NewMockHasher(ctrl)
	svc = service.NewUserService(userRepoMock, s3BucketMock, fileInvalidatorMock, cdnUrlSignerMock, hasherMock)

	return
}

func TestFindUserById_Success(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, cdnUrlSignerMock, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	cdnUrlSignerMock.EXPECT().SignURL(userEntity.PictureName.String).Return(signedPicture, nil)

	userDTO, err := svc.FindUserById(ctx, userEntity.ID)

	assert.NoError(t, err)

	assert.Equal(t, strconv.FormatInt(userEntity.ID, 10), userDTO.ID)
	assert.Equal(t, userEntity.FirstName, userDTO.FirstName)
	assert.Equal(t, userEntity.LastName, userDTO.LastName)
	assert.Equal(t, userEntity.PhoneNumber, userDTO.PhoneNumber)
	assert.Equal(t, userEntity.Email, userDTO.Email)
	assert.Equal(t, signedPicture, userDTO.PictureUrl)
	assert.Equal(t, string(userEntity.Role), userDTO.Role)
	assert.Equal(t, userEntity.UserStatus.Bool, userDTO.UserStatus)
	assert.Equal(t, userEntity.CreatedAt.Time, userDTO.CreatedAt)
}

func TestFindUserById_NotFound(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(entity.User{}, pgx.ErrNoRows)

	userDTO, err := svc.FindUserById(ctx, userEntity.ID)

	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, userDTO)
}

func TestFindUserById_MapperError(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, cdnUrlSignerMock, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	cdnUrlSignerMock.EXPECT().SignURL(userEntity.PictureName.String).Return(signedPicture, errSigningUrl)

	userDTO, err := svc.FindUserById(ctx, userEntity.ID)

	assert.ErrorIs(t, err, errSigningUrl)
	assert.Empty(t, userDTO)
}

func TestFindUserProfileById_Success(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, cdnUrlSignerMock, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	cdnUrlSignerMock.EXPECT().SignURL(userEntity.PictureName.String).Return(signedPicture, nil)

	profileDTO, err := svc.FindUserProfileById(ctx, userEntity.ID)

	assert.NoError(t, err)

	assert.Equal(t, strconv.FormatInt(userEntity.ID, 10), profileDTO.ID)
	assert.Equal(t, userEntity.FirstName, profileDTO.FirstName)
	assert.Equal(t, userEntity.LastName, profileDTO.LastName)
	assert.Equal(t, signedPicture, profileDTO.PictureUrl)
	assert.Equal(t, string(userEntity.Role), profileDTO.Role)
	assert.Equal(t, userEntity.UserStatus.Bool, profileDTO.UserStatus)
	assert.Equal(t, userEntity.CreatedAt.Time, profileDTO.CreatedAt)
}

func TestFindUserProfileById_NotFound(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(entity.User{}, pgx.ErrNoRows)

	profileDTO, err := svc.FindUserProfileById(ctx, userEntity.ID)

	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, profileDTO)
}

func TestFindUserProfileById_MapperError(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, cdnUrlSignerMock, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	cdnUrlSignerMock.EXPECT().SignURL(userEntity.PictureName.String).Return(signedPicture, errSigningUrl)

	profileDTO, err := svc.FindUserProfileById(ctx, userEntity.ID)

	assert.ErrorIs(t, err, errSigningUrl)
	assert.Empty(t, profileDTO)
}

func TestUpdateUserDataById_Success(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, cdnUrlSignerMock, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	updateUserDTO := dto.UpdateUser{Email: &userEntity.Email, PhoneNumber: &userEntity.PhoneNumber, FirstName: &userEntity.FirstName, LastName: &userEntity.LastName}
	updateUserParams := repository.UpdateUserParams{ID: userEntity.ID, FirstName: updateUserDTO.FirstName, LastName: updateUserDTO.LastName, PhoneNumber: updateUserDTO.PhoneNumber, Email: updateUserDTO.Email}

	userRepoMock.EXPECT().Update(ctx, updateUserParams).Return(userEntity, nil)
	cdnUrlSignerMock.EXPECT().SignURL(userEntity.PictureName.String).Return(signedPicture, nil)

	userDTO, err := svc.UpdateUserDataById(ctx, userEntity.ID, updateUserDTO)

	assert.NoError(t, err)

	assert.Equal(t, strconv.FormatInt(userEntity.ID, 10), userDTO.ID)
	assert.Equal(t, userEntity.Email, userDTO.Email)
	assert.Equal(t, userEntity.PhoneNumber, userDTO.PhoneNumber)
	assert.Equal(t, userEntity.FirstName, userDTO.FirstName)
	assert.Equal(t, userEntity.LastName, userDTO.LastName)
}

func TestUpdateUserDataById_NotFound(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	updateUserDTO := dto.UpdateUser{Email: &userEntity.Email, PhoneNumber: &userEntity.PhoneNumber, FirstName: &userEntity.FirstName, LastName: &userEntity.LastName}
	updateUserParams := repository.UpdateUserParams{ID: userEntity.ID, FirstName: updateUserDTO.FirstName, LastName: updateUserDTO.LastName, PhoneNumber: updateUserDTO.PhoneNumber, Email: updateUserDTO.Email}

	userRepoMock.EXPECT().Update(ctx, updateUserParams).Return(entity.User{}, pgx.ErrNoRows)

	userDTO, err := svc.UpdateUserDataById(ctx, userEntity.ID, updateUserDTO)

	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, userDTO)
}

func TestUpdateUserDataById_SignerError(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, cdnUrlSignerMock, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	updateUserDTO := dto.UpdateUser{Email: &userEntity.Email, PhoneNumber: &userEntity.PhoneNumber, FirstName: &userEntity.FirstName, LastName: &userEntity.LastName}
	updateUserParams := repository.UpdateUserParams{ID: userEntity.ID, FirstName: updateUserDTO.FirstName, LastName: updateUserDTO.LastName, PhoneNumber: updateUserDTO.PhoneNumber, Email: updateUserDTO.Email}

	userRepoMock.EXPECT().Update(ctx, updateUserParams).Return(userEntity, nil)
	cdnUrlSignerMock.EXPECT().SignURL(userEntity.PictureName.String).Return("", errSigningUrl)

	userDTO, err := svc.UpdateUserDataById(ctx, userEntity.ID, updateUserDTO)

	assert.ErrorIs(t, err, errSigningUrl)
	assert.Empty(t, userDTO)
}

func TestSetProfilePicture_Success(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, s3BucketMock, fileInvalidatorMock, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	userRepoMock.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(nil)
	s3BucketMock.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)
	s3BucketMock.EXPECT().DeleteObject(ctx, userEntity.PictureName.String).Return(nil)
	fileInvalidatorMock.EXPECT().InvalidateFile(ctx, userEntity.PictureName.String).Return(nil)

	err := svc.SetProfilePicture(ctx, userEntity.ID, file, fileName, file.Size(), fileType)
	assert.NoError(t, err)
}

func TestSetProfilePicture_WithoutPicture(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, s3BucketMock, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(nil)
	userRepoMock.EXPECT().GetById(ctx, userEntityWithoutPicture.ID).Return(userEntityWithoutPicture, nil)
	s3BucketMock.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)

	err := svc.SetProfilePicture(ctx, userEntityWithoutPicture.ID, file, fileName, file.Size(), fileType)
	assert.NoError(t, err)
}

func TestSetProfilePicture_NotFound(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, pgx.ErrNoRows)

	err := svc.SetProfilePicture(ctx, userEntity.ID, file, fileName, file.Size(), fileType)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestSetProfilePicture_PutObjectError(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, s3BucketMock, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	s3BucketMock.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return("", errS3PutObject)

	err := svc.SetProfilePicture(ctx, userEntity.ID, file, fileName, file.Size(), fileType)
	assert.ErrorIs(t, err, errS3PutObject)
}

func TestSetProfilePicture_UpdatePictureError(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, s3BucketMock, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	userRepoMock.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(pgx.ErrNoRows)
	s3BucketMock.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)

	err := svc.SetProfilePicture(ctx, userEntity.ID, file, fileName, file.Size(), fileType)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestSetProfilePicture_DeleteObjectError(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, s3BucketMock, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	userRepoMock.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(nil)
	s3BucketMock.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)
	s3BucketMock.EXPECT().DeleteObject(ctx, userEntity.PictureName.String).Return(errS3DeleteObject)

	err := svc.SetProfilePicture(ctx, userEntity.ID, file, fileName, file.Size(), fileType)
	assert.ErrorIs(t, err, errS3DeleteObject)
}

func TestSetProfilePicture_InvalidationError(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, s3BucketMock, fileInvalidatorMock, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	userRepoMock.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(nil)
	s3BucketMock.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)
	s3BucketMock.EXPECT().DeleteObject(ctx, userEntity.PictureName.String).Return(nil)
	fileInvalidatorMock.EXPECT().InvalidateFile(ctx, userEntity.PictureName.String).Return(errCdnFileInvalidation)

	err := svc.SetProfilePicture(ctx, userEntity.ID, file, fileName, file.Size(), fileType)
	assert.ErrorIs(t, err, errCdnFileInvalidation)
}

func TestChangePassword_Success(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, _, _, hasherMock := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetHashById(ctx, userEntity.ID).Return(userEntity.Hash, nil)
	userRepoMock.EXPECT().UpdateHash(ctx, gomock.Any()).Return(nil)
	hasherMock.EXPECT().VerifyPassword(gomock.Any(), userEntity.Hash).Return(nil)
	hasherMock.EXPECT().HashPassword(gomock.Any()).Return("", nil)

	assert.NoError(t, svc.ChangePassword(ctx, userEntity.ID, dto.UpdatePassword{}))
}

func TestChangePassword_NotFound(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetHashById(ctx, userEntity.ID).Return("", pgx.ErrNoRows)

	assert.ErrorIs(t, svc.ChangePassword(ctx, userEntity.ID, dto.UpdatePassword{}), pgx.ErrNoRows)
}

func TestChangePassword_IncorrectPassword(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, _, _, hasherMock := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetHashById(ctx, userEntity.ID).Return(userEntity.Hash, nil)
	hasherMock.EXPECT().VerifyPassword(gomock.Any(), userEntity.Hash).Return(hasher.ErrPasswordMismatch)

	assert.ErrorIs(t, svc.ChangePassword(ctx, userEntity.ID, dto.UpdatePassword{}), hasher.ErrPasswordMismatch)
}

func TestChangePassword_UpdateError(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, _, _, hasherMock := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetHashById(ctx, userEntity.ID).Return(userEntity.Hash, nil)
	userRepoMock.EXPECT().UpdateHash(ctx, gomock.Any()).Return(errUpdateError)
	hasherMock.EXPECT().VerifyPassword(gomock.Any(), userEntity.Hash).Return(nil)
	hasherMock.EXPECT().HashPassword(gomock.Any()).Return("", nil)

	assert.ErrorIs(t, svc.ChangePassword(ctx, userEntity.ID, dto.UpdatePassword{}), errUpdateError)
}

func TestDeleteUserById_Success(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, s3BucketMock, fileInvalidatorMock, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	userRepoMock.EXPECT().DeleteById(ctx, userEntity.ID).Return(nil)
	s3BucketMock.EXPECT().DeleteObject(ctx, userEntity.PictureName.String).Return(nil)
	fileInvalidatorMock.EXPECT().InvalidateFile(ctx, userEntity.PictureName.String).Return(nil)

	assert.NoError(t, svc.DeleteUserById(ctx, userEntity.ID))
}

func TestDeleteUserById_EmptyPicture(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntityWithoutPicture.ID).Return(userEntityWithoutPicture, nil)
	userRepoMock.EXPECT().DeleteById(ctx, userEntityWithoutPicture.ID).Return(nil)

	assert.NoError(t, svc.DeleteUserById(ctx, userEntity.ID))
}

func TestDeleteUserById_NotFound(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(entity.User{}, pgx.ErrNoRows)

	assert.ErrorIs(t, svc.DeleteUserById(ctx, userEntity.ID), pgx.ErrNoRows)
}

func TestDeleteUserById_DeleteObjectError(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, s3BucketMock, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	s3BucketMock.EXPECT().DeleteObject(ctx, userEntity.PictureName.String).Return(errS3DeleteObject)

	assert.ErrorIs(t, svc.DeleteUserById(ctx, userEntity.ID), errS3DeleteObject)
}

func TestDeleteUserById_InvalidationError(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, s3BucketMock, fileInvalidatorMock, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	s3BucketMock.EXPECT().DeleteObject(ctx, userEntity.PictureName.String).Return(nil)
	fileInvalidatorMock.EXPECT().InvalidateFile(ctx, userEntity.PictureName.String).Return(errCdnFileInvalidation)

	assert.ErrorIs(t, svc.DeleteUserById(ctx, userEntity.ID), errCdnFileInvalidation)
}

func TestDeleteUserById_RowDeletionError(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, s3BucketMock, fileInvalidatorMock, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetById(ctx, userEntity.ID).Return(userEntity, nil)
	userRepoMock.EXPECT().DeleteById(ctx, userEntity.ID).Return(errDeleteError)
	s3BucketMock.EXPECT().DeleteObject(ctx, userEntity.PictureName.String).Return(nil)
	fileInvalidatorMock.EXPECT().InvalidateFile(ctx, userEntity.PictureName.String).Return(nil)

	assert.ErrorIs(t, svc.DeleteUserById(ctx, userEntity.ID), errDeleteError)
}
