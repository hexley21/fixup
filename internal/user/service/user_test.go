package service_test

import (
	"bytes"
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/user/domain"
	"github.com/hexley21/fixup/internal/user/repository"
	mockRepository "github.com/hexley21/fixup/internal/user/repository/mock"
	"github.com/hexley21/fixup/internal/user/service"
	mockHasher "github.com/hexley21/fixup/pkg/hasher/mock"
	mockCdn "github.com/hexley21/fixup/pkg/infra/cdn/mock"
	mockS3 "github.com/hexley21/fixup/pkg/infra/s3/mock"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)


// TODO: rewrite tests

var (
	userModel = repository.User{
		ID:          1,
		FirstName:   "Larry",
		LastName:    "Page",
		PhoneNumber: "995111222333",
		Email:       "larry@page.com",
		Picture:     pgtype.Text{String: "larrypage.jpg", Valid: true},
		Hash:        "",
		Role:        string(enum.UserRoleADMIN),
		Verified:      pgtype.Bool{Bool: true, Valid: true},
		CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	userModelWithoutPicture = repository.User{
		ID:          1,
		FirstName:   "Larry",
		LastName:    "Page",
		PhoneNumber: "995111222333",
		Email:       "larry@page.com",
		Picture:     pgtype.Text{String: "", Valid: false},
		Role:        string(enum.UserRoleADMIN),
		Verified:      pgtype.Bool{Bool: true, Valid: true},
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
	s3BucketMock *mockS3.MockBucket,
	fileInvalidatorMock *mockCdn.MockFileInvalidator,
	hasherMock *mockHasher.MockHasher,
) {
	ctx = context.Background()
	ctrl = gomock.NewController(t)
	userRepoMock = mockRepository.NewMockUserRepository(ctrl)
	s3BucketMock = mockS3.NewMockBucket(ctrl)
	fileInvalidatorMock = mockCdn.NewMockFileInvalidator(ctrl)
	hasherMock = mockHasher.NewMockHasher(ctrl)
	svc = service.NewUserService(userRepoMock, s3BucketMock, fileInvalidatorMock, hasherMock)

	return
}

func TestFindUserById_Success(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().Get(ctx, userModel.ID).Return(userModel, nil)

	userEntity, err := svc.Get(ctx, userModel.ID)

	assert.NoError(t, err)

	assert.Equal(t, strconv.FormatInt(userModel.ID, 10), userEntity.ID)
	assert.Equal(t, userModel.FirstName, userEntity.PersonalInfo.FirstName)
	assert.Equal(t, userModel.LastName, userEntity.PersonalInfo.LastName)
	assert.Equal(t, userModel.PhoneNumber, userEntity.PersonalInfo.PhoneNumber)
	assert.Equal(t, userModel.Email, userEntity.PersonalInfo.Email)
	assert.Equal(t, signedPicture, userEntity.Picture)
	assert.Equal(t, string(userModel.Role), userEntity.AccountInfo.Role)
	assert.Equal(t, userModel.Verified.Bool, userEntity.AccountInfo.Verified)
	assert.Equal(t, userModel.CreatedAt.Time, userEntity.CreatedAt)
}

func TestFindUserById_NotFound(t *testing.T) {
	ctx, ctrl, svc, userRepoMock, _, _, _ := setupUser(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().Get(ctx, userModel.ID).Return(domain.User{}, pgx.ErrNoRows)

	userEntity, err := svc.Get(ctx, userModel.ID)

	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, userEntity)
}

// func TestUpdateUserDataById_Success(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, _, _, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	updateuserEntity := dto.UpdateUser{Email: &userModel.Email, PhoneNumber: &userModel.PhoneNumber, FirstName: &userModel.FirstName, LastName: &userModel.LastName}
// 	updateUserParams := repository.UpdateUserRow{FirstName: updateuserEntity.PersonalInfo.FirstName, LastName: updateuserEntity.PersonalInfo.LastName, PhoneNumber: updateuserEntity.PersonalInfo.PhoneNumber, Email: updateuserEntity.PersonalInfo.Email}

// 	userRepoMock.EXPECT().Update(ctx, userModel.ID, updateUserParams).Return(userModel, nil)

// 	userEntity, err := svc.UpdateUserDataById(ctx, userModel.ID, updateuserEntity)

// 	assert.NoError(t, err)

// 	assert.Equal(t, strconv.FormatInt(userModel.ID, 10), userEntity.ID)
// 	assert.Equal(t, userModel.Email, userEntity.PersonalInfo.Email)
// 	assert.Equal(t, userModel.PhoneNumber, userEntity.PersonalInfo.PhoneNumber)
// 	assert.Equal(t, userModel.FirstName, userEntity.PersonalInfo.FirstName)
// 	assert.Equal(t, userModel.LastName, userEntity.PersonalInfo.LastName)
// }

// func TestUpdateUserDataById_NotFound(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, _, _, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	updateuserEntity := dto.UpdateUser{Email: &userModel.Email, PhoneNumber: &userModel.PhoneNumber, FirstName: &userModel.FirstName, LastName: &userModel.LastName}
// 	updateUserParams := repository.UpdateUserParams{ID: userModel.ID, FirstName: updateuserEntity.PersonalInfo.FirstName, LastName: updateuserEntity.PersonalInfo.LastName, PhoneNumber: updateuserEntity.PersonalInfo.PhoneNumber, Email: updateuserEntity.PersonalInfo.Email}

// 	userRepoMock.EXPECT().Update(ctx, updateUserParams).Return(domain.User{}, pgx.ErrNoRows)

// 	userEntity, err := svc.UpdateUserDataById(ctx, userModel.ID, updateuserEntity)

// 	assert.ErrorIs(t, err, pgx.ErrNoRows)
// 	assert.Empty(t, userEntity)
// }

// func TestSetProfilePicture_Success(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, s3BucketMock, fileInvalidatorMock, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().Get(ctx, userModel.ID).Return(userModel, nil)
// 	userRepoMock.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(nil)
// 	s3BucketMock.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)
// 	s3BucketMock.EXPECT().DeleteObject(ctx, userModel.PictureName.String).Return(nil)
// 	fileInvalidatorMock.EXPECT().InvalidateFile(ctx, userModel.PictureName.String).Return(nil)

// 	err := svc.SetProfilePicture(ctx, userModel.ID, file, fileName, file.Size(), fileType)
// 	assert.NoError(t, err)
// }

// func TestSetProfilePicture_WithoutPicture(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, s3BucketMock, _, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(nil)
// 	userRepoMock.EXPECT().Get(ctx, userModelWithoutPicture.ID).Return(userModelWithoutPicture, nil)
// 	s3BucketMock.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)

// 	err := svc.SetProfilePicture(ctx, userModelWithoutPicture.ID, file, fileName, file.Size(), fileType)
// 	assert.NoError(t, err)
// }

// func TestSetProfilePicture_NotFound(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, _, _, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().Get(ctx, userModel.ID).Return(userModel, pgx.ErrNoRows)

// 	err := svc.SetProfilePicture(ctx, userModel.ID, file, fileName, file.Size(), fileType)
// 	assert.ErrorIs(t, err, pgx.ErrNoRows)
// }

// func TestSetProfilePicture_PutObjectError(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, s3BucketMock, _, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().Get(ctx, userModel.ID).Return(userModel, nil)
// 	s3BucketMock.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return("", errS3PutObject)

// 	err := svc.SetProfilePicture(ctx, userModel.ID, file, fileName, file.Size(), fileType)
// 	assert.ErrorIs(t, err, errS3PutObject)
// }

// func TestSetProfilePicture_UpdatePictureError(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, s3BucketMock, _, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().Get(ctx, userModel.ID).Return(userModel, nil)
// 	userRepoMock.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(pgx.ErrNoRows)
// 	s3BucketMock.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)

// 	err := svc.SetProfilePicture(ctx, userModel.ID, file, fileName, file.Size(), fileType)
// 	assert.ErrorIs(t, err, pgx.ErrNoRows)
// }

// func TestSetProfilePicture_DeleteObjectError(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, s3BucketMock, _, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().Get(ctx, userModel.ID).Return(userModel, nil)
// 	userRepoMock.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(nil)
// 	s3BucketMock.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)
// 	s3BucketMock.EXPECT().DeleteObject(ctx, userModel.PictureName.String).Return(errS3DeleteObject)

// 	err := svc.SetProfilePicture(ctx, userModel.ID, file, fileName, file.Size(), fileType)
// 	assert.ErrorIs(t, err, errS3DeleteObject)
// }

// func TestSetProfilePicture_InvalidationError(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, s3BucketMock, fileInvalidatorMock, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().Get(ctx, userModel.ID).Return(userModel, nil)
// 	userRepoMock.EXPECT().UpdatePicture(ctx, gomock.Any()).Return(nil)
// 	s3BucketMock.EXPECT().PutObject(ctx, file, gomock.Any(), fileName, file.Size(), fileType).Return(randomFilename, nil)
// 	s3BucketMock.EXPECT().DeleteObject(ctx, userModel.PictureName.String).Return(nil)
// 	fileInvalidatorMock.EXPECT().InvalidateFile(ctx, userModel.PictureName.String).Return(errCdnFileInvalidation)

// 	err := svc.SetProfilePicture(ctx, userModel.ID, file, fileName, file.Size(), fileType)
// 	assert.ErrorIs(t, err, errCdnFileInvalidation)
// }

// func TestChangePassword_Success(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, _, _, _, hasherMock := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().GetHashById(ctx, userModel.ID).Return(userModel.Hash, nil)
// 	userRepoMock.EXPECT().UpdateHash(ctx, gomock.Any()).Return(nil)
// 	hasherMock.EXPECT().VerifyPassword(gomock.Any(), userModel.Hash).Return(nil)
// 	hasherMock.EXPECT().HashPassword(gomock.Any()).Return("", nil)

// 	assert.NoError(t, svc.ChangePassword(ctx, userModel.ID, dto.UpdatePassword{}))
// }

// func TestChangePassword_NotFound(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, _, _, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().GetHashById(ctx, userModel.ID).Return("", pgx.ErrNoRows)

// 	assert.ErrorIs(t, svc.ChangePassword(ctx, userModel.ID, dto.UpdatePassword{}), pgx.ErrNoRows)
// }

// func TestChangePassword_IncorrectPassword(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, _, _, _, hasherMock := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().GetHashById(ctx, userModel.ID).Return(userModel.Hash, nil)
// 	hasherMock.EXPECT().VerifyPassword(gomock.Any(), userModel.Hash).Return(hasher.ErrPasswordMismatch)

// 	assert.ErrorIs(t, svc.ChangePassword(ctx, userModel.ID, dto.UpdatePassword{}), hasher.ErrPasswordMismatch)
// }

// func TestChangePassword_UpdateError(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, _, _, _, hasherMock := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().GetHashById(ctx, userModel.ID).Return(userModel.Hash, nil)
// 	userRepoMock.EXPECT().UpdateHash(ctx, gomock.Any()).Return(errUpdateError)
// 	hasherMock.EXPECT().VerifyPassword(gomock.Any(), userModel.Hash).Return(nil)
// 	hasherMock.EXPECT().HashPassword(gomock.Any()).Return("", nil)

// 	assert.ErrorIs(t, svc.ChangePassword(ctx, userModel.ID, dto.UpdatePassword{}), errUpdateError)
// }

// func TestDeleteUserById_Success(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, s3BucketMock, fileInvalidatorMock, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().Get(ctx, userModel.ID).Return(userModel, nil)
// 	userRepoMock.EXPECT().DeleteById(ctx, userModel.ID).Return(nil)
// 	s3BucketMock.EXPECT().DeleteObject(ctx, userModel.PictureName.String).Return(nil)
// 	fileInvalidatorMock.EXPECT().InvalidateFile(ctx, userModel.PictureName.String).Return(nil)

// 	assert.NoError(t, svc.DeleteUserById(ctx, userModel.ID))
// }

// func TestDeleteUserById_EmptyPicture(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, _, _, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().Get(ctx, userModelWithoutPicture.ID).Return(userModelWithoutPicture, nil)
// 	userRepoMock.EXPECT().DeleteById(ctx, userModelWithoutPicture.ID).Return(nil)

// 	assert.NoError(t, svc.DeleteUserById(ctx, userModel.ID))
// }

// func TestDeleteUserById_NotFound(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, _, _, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().Get(ctx, userModel.ID).Return(domain.User{}, pgx.ErrNoRows)

// 	assert.ErrorIs(t, svc.DeleteUserById(ctx, userModel.ID), pgx.ErrNoRows)
// }

// func TestDeleteUserById_DeleteObjectError(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, s3BucketMock, _, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().Get(ctx, userModel.ID).Return(userModel, nil)
// 	s3BucketMock.EXPECT().DeleteObject(ctx, userModel.PictureName.String).Return(errS3DeleteObject)

// 	assert.ErrorIs(t, svc.DeleteUserById(ctx, userModel.ID), errS3DeleteObject)
// }

// func TestDeleteUserById_InvalidationError(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, s3BucketMock, fileInvalidatorMock, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().Get(ctx, userModel.ID).Return(userModel, nil)
// 	s3BucketMock.EXPECT().DeleteObject(ctx, userModel.PictureName.String).Return(nil)
// 	fileInvalidatorMock.EXPECT().InvalidateFile(ctx, userModel.PictureName.String).Return(errCdnFileInvalidation)

// 	assert.ErrorIs(t, svc.DeleteUserById(ctx, userModel.ID), errCdnFileInvalidation)
// }

// func TestDeleteUserById_RowDeletionError(t *testing.T) {
// 	ctx, ctrl, svc, userRepoMock, s3BucketMock, fileInvalidatorMock, _ := setupUser(t)
// 	defer ctrl.Finish()

// 	userRepoMock.EXPECT().Get(ctx, userModel.ID).Return(userModel, nil)
// 	userRepoMock.EXPECT().DeleteById(ctx, userModel.ID).Return(errDeleteError)
// 	s3BucketMock.EXPECT().DeleteObject(ctx, userModel.PictureName.String).Return(nil)
// 	fileInvalidatorMock.EXPECT().InvalidateFile(ctx, userModel.PictureName.String).Return(nil)

// 	assert.ErrorIs(t, svc.DeleteUserById(ctx, userModel.ID), errDeleteError)
// }
