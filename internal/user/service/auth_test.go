package service_test

import (
	"context"
	"html/template"
	"strconv"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/hexley21/fixup/internal/user/repository"
	mock_repository "github.com/hexley21/fixup/internal/user/repository/mock"
	"github.com/hexley21/fixup/internal/user/service"
	mock_verifier "github.com/hexley21/fixup/internal/user/service/verifier/mock"
	mock_encryption "github.com/hexley21/fixup/pkg/encryption/mock"
	mock_hasher "github.com/hexley21/fixup/pkg/hasher/mock"
	mock_cdn "github.com/hexley21/fixup/pkg/infra/cdn/mock"
	mock_postgres "github.com/hexley21/fixup/pkg/infra/postgres/mock"
	mock_mailer "github.com/hexley21/fixup/pkg/mailer/mock"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
	userDto = dto.User{
		ID:          "1",
		FirstName:   "Larry",
		LastName:    "Page",
		PhoneNumber: "995111222333",
		Email:       "larry@page.com",
		PictureUrl:  "larrypage.png",
		Role:        string(enum.UserRoleADMIN),
		UserStatus:  true,
		CreatedAt:   time.Now(),
	}

	registerUserDto = dto.RegisterUser{
		Email:       "larry@page.com",
		PhoneNumber: "995111222333",
		FirstName:   "Larry",
		LastName:    "Page",
		Password:    "12345678",
	}

	registerProviderDto = dto.RegisterProvider{
		RegisterUser: registerUserDto,
		PersonalIDNumber: "1234567890",
	}

	loginDto = dto.Login{
		Email: "larry@page.com",
		Password: "12345678",
	}

	mockEmailAddress = "fixup@gmail.com"
	newHash          = "hash"
	newToken         = "VerificationToken"
	newUrl           = "picture.png?signed=true"

	confiramtionTemplate = template.New("confirmation")
	verifiedTemplate     = template.New("verified")
)

func TestRegisterCustomer(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockPgx := mock_postgres.NewMockPGX(ctrl)
	mockTx := mock_postgres.NewMockTx(ctrl)
	mockHasher := mock_hasher.NewMockHasher(ctrl)
	mockJwtGenerator := mock_verifier.NewMockJwtGenerator(ctrl)
	mockMailer := mock_mailer.NewMockMailer(ctrl)
	mockUrlSigner := mock_cdn.NewMockURLSigner(ctrl)

	mockUserRepo.EXPECT().WithTx(mockTx).Return(mockUserRepo)
	mockUserRepo.EXPECT().Create(ctx, gomock.Any()).Return(userEntity, nil)

	mockPgx.EXPECT().BeginTx(ctx, gomock.Any()).Return(mockTx, nil)
	mockTx.EXPECT().Commit(ctx).Return(nil)

	mockHasher.EXPECT().HashPassword(gomock.Any()).Return(newHash)
	mockJwtGenerator.EXPECT().GenerateToken(gomock.Any(), gomock.Any()).Return(newToken, nil)
	mockMailer.EXPECT().SendHTML(mockEmailAddress, gomock.Any(), gomock.Any(), confiramtionTemplate, gomock.Any()).Return(nil)
	mockUrlSigner.EXPECT().SignURL(gomock.Any()).Return(newUrl, nil)

	service := service.NewAuthService(mockUserRepo, nil, mockPgx, mockHasher, nil, mockMailer, mockEmailAddress, mockUrlSigner, mockJwtGenerator)
	service.SetTemplates(confiramtionTemplate, nil)

	dto, err := service.RegisterCustomer(ctx, registerUserDto)
	assert.NoError(t, err)
	assert.Equal(t, dto.Email, registerUserDto.Email)
	assert.Equal(t, dto.PhoneNumber, registerUserDto.PhoneNumber)
	assert.Equal(t, dto.FirstName, registerUserDto.FirstName)
	assert.Equal(t, dto.LastName, registerUserDto.LastName)

	time.Sleep(100 * time.Millisecond)
}

func TestRegisterProvider(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockProviderRepo := mock_repository.NewMockProviderRepository(ctrl)
	mockPgx := mock_postgres.NewMockPGX(ctrl)
	mockTx := mock_postgres.NewMockTx(ctrl)
	mockHasher := mock_hasher.NewMockHasher(ctrl)
	mockEncryptor := mock_encryption.NewMockEncryptor(ctrl)
	mockMailer := mock_mailer.NewMockMailer(ctrl)
	mockUrlSigner := mock_cdn.NewMockURLSigner(ctrl)
	mockJwtGenerator := mock_verifier.NewMockJwtGenerator(ctrl)

	mockUserRepo.EXPECT().WithTx(mockTx).Return(mockUserRepo)
	mockUserRepo.EXPECT().Create(ctx, gomock.Any()).Return(userEntity, nil)
	mockProviderRepo.EXPECT().WithTx(mockTx).Return(mockProviderRepo)
	mockProviderRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

	mockPgx.EXPECT().BeginTx(ctx, gomock.Any()).Return(mockTx, nil)
	mockTx.EXPECT().Commit(ctx).Return(nil)

	mockHasher.EXPECT().HashPassword(gomock.Any()).Return(newHash)
	mockEncryptor.EXPECT().Encrypt(gomock.Any()).Return([]byte(registerProviderDto.PersonalIDNumber), nil)
	mockJwtGenerator.EXPECT().GenerateToken(gomock.Any(), gomock.Any()).Return(newToken, nil)
	mockMailer.EXPECT().SendHTML(mockEmailAddress, gomock.Any(), gomock.Any(), confiramtionTemplate, gomock.Any()).Return(nil)
	mockUrlSigner.EXPECT().SignURL(gomock.Any()).Return(newUrl, nil)

	service := service.NewAuthService(mockUserRepo, mockProviderRepo, mockPgx, mockHasher, mockEncryptor, mockMailer, mockEmailAddress, mockUrlSigner, mockJwtGenerator)
	service.SetTemplates(confiramtionTemplate, nil)

	dto, err := service.RegisterProvider(ctx, registerProviderDto)
	assert.NoError(t, err)
	assert.Equal(t, dto.Email, registerUserDto.Email)
	assert.Equal(t, dto.PhoneNumber, registerUserDto.PhoneNumber)
	assert.Equal(t, dto.FirstName, registerUserDto.FirstName)
	assert.Equal(t, dto.LastName, registerUserDto.LastName)

	time.Sleep(100 * time.Millisecond)
}

func TestAuthenticateUser(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockHasher := mock_hasher.NewMockHasher(ctrl)

	creds := repository.GetCredentialsByEmailRow{
		ID: 1,
		Role: enum.UserRoleADMIN,
		Hash: newHash,
		UserStatus: pgtype.Bool{Bool: true, Valid: true},
	}

	mockUserRepo.EXPECT().GetCredentialsByEmail(ctx, loginDto.Email).Return(creds, nil)
	mockHasher.EXPECT().VerifyPassword(loginDto.Password, creds.Hash).Return(nil)

	service := service.NewAuthService(mockUserRepo, nil, nil, mockHasher, nil, nil, mockEmailAddress, nil, nil)

	credentialsDto, err := service.AuthenticateUser(ctx, loginDto)
	assert.NoError(t, err)
	assert.Equal(t, credentialsDto.ID, strconv.FormatInt(creds.ID, 10))
	assert.Equal(t, credentialsDto.Role, string(credentialsDto.Role))
	assert.Equal(t, credentialsDto.UserStatus, creds.UserStatus.Bool)
}

func TestVerifyUser(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockMailer := mock_mailer.NewMockMailer(ctrl)

	mockUserRepo.EXPECT().UpdateStatus(ctx, gomock.Any()).Return(nil)
	mockMailer.EXPECT().SendHTML(mockEmailAddress, gomock.Any(), gomock.Any(), verifiedTemplate, gomock.Any()).Return(nil)

	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, mockMailer, mockEmailAddress, nil, nil)
	service.SetTemplates(nil, verifiedTemplate)

	err := service.VerifyUser(ctx, 1, "")
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)
}
