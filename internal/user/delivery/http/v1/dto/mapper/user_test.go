package mapper_test

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto/mapper"
	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/internal/common/enum"
	mock_cdn "github.com/hexley21/fixup/pkg/infra/cdn/mock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	errSigning = errors.New("error signing photo")
	signedUrl = "http://signed.url/"

	customerEntity = entity.User{
		ID: 1,
		FirstName: "Larry",
		LastName: "Page",
		PhoneNumber: "995555555555",
		Email: "larry@page.com",
		PictureName: pgtype.Text{String: "photograph.png", Valid: true},
		Hash: "abcd",
		Role: enum.UserRoleCUSTOMER,
		UserStatus: pgtype.Bool{Bool: true, Valid: true},
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	providerEntity = entity.User{
		ID: 1,
		FirstName: "Larry",
		LastName: "Page",
		PhoneNumber: "995555555555",
		Email: "larry@page.com",
		PictureName: pgtype.Text{String: "photograph.png", Valid: true},
		Hash: "abcd",
		Role: enum.UserRolePROVIDER,
		UserStatus: pgtype.Bool{Bool: true, Valid: true},
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
)

func TestMapUserToDto_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUrlSigner := mock_cdn.NewMockURLSigner(ctrl)
	mockUrlSigner.EXPECT().SignURL(gomock.Any()).Return(signedUrl, nil)

	dto, err := mapper.MapUserToDto(customerEntity, mockUrlSigner)
	assert.NoError(t, err)

	assert.Equal(t, strconv.FormatInt(customerEntity.ID, 10), dto.ID)
	assert.Equal(t, customerEntity.FirstName, dto.FirstName)
	assert.Equal(t, customerEntity.PhoneNumber, dto.PhoneNumber)
	assert.Equal(t, customerEntity.Email, dto.Email)
	assert.Equal(t, signedUrl, dto.PictureUrl)
	assert.Equal(t, string(customerEntity.Role), dto.Role)
	assert.Equal(t, customerEntity.UserStatus.Bool, dto.UserStatus)
	assert.Equal(t, customerEntity.CreatedAt.Time, dto.CreatedAt)
}

func TestMapUserToDto_SignError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUrlSigner := mock_cdn.NewMockURLSigner(ctrl)
	mockUrlSigner.EXPECT().SignURL(gomock.Any()).Return("", errSigning)

	dto, err := mapper.MapUserToDto(entity.User{PictureName: customerEntity.PictureName}, mockUrlSigner)
	assert.ErrorIs(t, err, errSigning)
	assert.Empty(t, dto)
}


func TestMapCustomerToProfileDto_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUrlSigner := mock_cdn.NewMockURLSigner(ctrl)
	mockUrlSigner.EXPECT().SignURL(gomock.Any()).Return(signedUrl, nil)

	dto, err := mapper.MapUserToProfileDto(customerEntity, mockUrlSigner)
	assert.NoError(t, err)

	assert.Equal(t, strconv.FormatInt(customerEntity.ID, 10), dto.ID)
	assert.Equal(t, customerEntity.FirstName, dto.FirstName)
	assert.Empty(t, dto.PhoneNumber)
	assert.Empty(t, dto.Email)
	assert.Equal(t, signedUrl, dto.PictureUrl)
	assert.Equal(t, string(customerEntity.Role), dto.Role)
	assert.Equal(t, customerEntity.UserStatus.Bool, dto.UserStatus)
	assert.Equal(t, customerEntity.CreatedAt.Time, dto.CreatedAt)
}

func TestMapProviderToProfileDto_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUrlSigner := mock_cdn.NewMockURLSigner(ctrl)
	mockUrlSigner.EXPECT().SignURL(gomock.Any()).Return(signedUrl, nil)

	dto, err := mapper.MapUserToProfileDto(providerEntity, mockUrlSigner)
	assert.NoError(t, err)

	assert.Equal(t, strconv.FormatInt(providerEntity.ID, 10), dto.ID)
	assert.Equal(t, providerEntity.FirstName, dto.FirstName)
	assert.Equal(t, providerEntity.PhoneNumber, dto.PhoneNumber)
	assert.Equal(t, providerEntity.Email, dto.Email)
	assert.Equal(t, signedUrl, dto.PictureUrl)
	assert.Equal(t, string(providerEntity.Role), dto.Role)
	assert.Equal(t, providerEntity.UserStatus.Bool, dto.UserStatus)
	assert.Equal(t, providerEntity.CreatedAt.Time, dto.CreatedAt)
}

func TestMapUserToProfileDto_SignError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUrlSigner := mock_cdn.NewMockURLSigner(ctrl)
	mockUrlSigner.EXPECT().SignURL(gomock.Any()).Return("", errSigning)

	dto, err := mapper.MapUserToProfileDto(entity.User{PictureName: providerEntity.PictureName}, mockUrlSigner)
	assert.ErrorIs(t, err, errSigning)
	assert.Empty(t, dto)
}
