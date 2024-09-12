package mapper_test

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto/mapper"
	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/internal/user/enum"
	mock_cdn "github.com/hexley21/fixup/pkg/infra/cdn/mock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	errSigning = errors.New("error signing photo")
	picture = "photograph.png"
	signedUrl = "http://signed.url/"
)

func TestMapUserToDto(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUrlSigner := mock_cdn.NewMockURLSigner(ctrl)
	mockUrlSigner.EXPECT().SignURL(picture).Return(signedUrl, nil)

	var pictureName pgtype.Text
	pictureName.Scan(picture)

	var userStatus pgtype.Bool
	userStatus.Scan(true)

	var createdAt pgtype.Timestamp
	createdAt.Scan(time.Now())

	entity := entity.User{
		ID: 1,
		FirstName: "Larry",
		LastName: "Page",
		PhoneNumber: "995555555555",
		Email: "larry@page.com",
		PictureName: pictureName,
		Hash: "abcd",
		Role: enum.UserRoleADMIN,
		UserStatus: userStatus,
		CreatedAt: createdAt,
	}

	dto, err := mapper.MapUserToDto(entity, mockUrlSigner)
	assert.NoError(t, err)

	assert.Equal(t, dto.ID, strconv.FormatInt(entity.ID, 10))
	assert.Equal(t, dto.FirstName, entity.FirstName)
	assert.Equal(t, dto.PhoneNumber, entity.PhoneNumber)
	assert.Equal(t, dto.Email, entity.Email)
	assert.Equal(t, dto.PictureUrl, signedUrl)
	assert.Equal(t, dto.Role, string(entity.Role))
	assert.Equal(t, dto.UserStatus, entity.UserStatus.Bool)
	assert.Equal(t, dto.CreatedAt, entity.CreatedAt.Time)
}

func TestMapUserToDtoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUrlSigner := mock_cdn.NewMockURLSigner(ctrl)
	mockUrlSigner.EXPECT().SignURL(picture).Return("", errSigning)

	var pictureName pgtype.Text
	pictureName.Scan(picture)

	dto, err := mapper.MapUserToDto(entity.User{PictureName: pictureName}, mockUrlSigner)
	assert.ErrorIs(t, err, errSigning)
	assert.Empty(t, dto)
}
