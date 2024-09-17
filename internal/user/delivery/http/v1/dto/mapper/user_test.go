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
	mockUrlSigner.EXPECT().SignURL(gomock.Any()).Return(signedUrl, nil)

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

	assert.Equal(t, strconv.FormatInt(entity.ID, 10), dto.ID)
	assert.Equal(t, entity.FirstName, dto.FirstName)
	assert.Equal(t, entity.PhoneNumber, dto.PhoneNumber)
	assert.Equal(t, entity.Email, dto.Email)
	assert.Equal(t, signedUrl, dto.PictureUrl)
	assert.Equal(t, string(entity.Role), dto.Role)
	assert.Equal(t, entity.UserStatus.Bool, dto.UserStatus)
	assert.Equal(t, entity.CreatedAt.Time, dto.CreatedAt)
}

func TestMapUserToDtoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockUrlSigner := mock_cdn.NewMockURLSigner(ctrl)
	mockUrlSigner.EXPECT().SignURL(gomock.Any()).Return("", errSigning)

	var pictureName pgtype.Text
	pictureName.Scan(picture)

	dto, err := mapper.MapUserToDto(entity.User{PictureName: pictureName}, mockUrlSigner)
	assert.ErrorIs(t, err, errSigning)
	assert.Empty(t, dto)
}
