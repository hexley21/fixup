package mapper_test

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/mapper"
	"github.com/hexley21/fixup/internal/user/domain"
	mock_cdn "github.com/hexley21/fixup/pkg/infra/cdn/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	errSigning = errors.New("error signing photo")
	signedUrl  = "http://signed.url/"

	customerEntity = domain.NewUser(
		1,
		"photograph.png",
		domain.NewUserPersonalInfo("larry@page.com", "995555555555", "Larry", "Page"),
		domain.NewUserAccountInfo(enum.UserRoleCUSTOMER, true),
		time.Now(),
	)
)

func setup(t *testing.T) (
	ctrl *gomock.Controller,
	mockUrlSigner *mock_cdn.MockURLSigner,
) {
	ctrl = gomock.NewController(t)
	mockUrlSigner = mock_cdn.NewMockURLSigner(ctrl)

	return
}

func TestMapUserEntityToDTO(t *testing.T) {
	ctrl, mockUrlSigner := setup(t)
	defer ctrl.Finish()

	tests := []struct {
		data          *domain.User
		expectedError error
		setupFunc     func()
	}{
		{
			data:      customerEntity,
			setupFunc: func() {
				mockUrlSigner.EXPECT().SignURL(customerEntity.Picture).Return(signedUrl, nil)
			},
		},
		{
			data:      customerEntity,
			expectedError: errSigning,
			setupFunc: func() {
				mockUrlSigner.EXPECT().SignURL(customerEntity.Picture).Return("", errSigning)
			},
		},
		{
			expectedError: mapper.ErrNilEntity,
			setupFunc: func() {},
		},
	}

	for _, tt := range tests {
		tt.setupFunc()

		userDTO, err := mapper.MapUserEntityToDTO(tt.data, mockUrlSigner)
		
		if tt.expectedError != nil {
			assert.ErrorIs(t, err, tt.expectedError)
			continue
		}

		assert.Equal(t, strconv.FormatInt(tt.data.ID, 10), userDTO.ID)
		assert.Equal(t, tt.data.PersonalInfo.FirstName, userDTO.FirstName)
		assert.Equal(t, tt.data.PersonalInfo.PhoneNumber, userDTO.PhoneNumber)
		assert.Equal(t, tt.data.PersonalInfo.Email, userDTO.Email)
		assert.Equal(t, signedUrl, userDTO.PictureUrl)
		assert.Equal(t, tt.data.AccountInfo.Role, userDTO.Role)
		assert.Equal(t, tt.data.AccountInfo.Active, userDTO.Active)
		assert.Equal(t, tt.data.CreatedAt, userDTO.CreatedAt)
	}
}
