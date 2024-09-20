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
	mock_encryption "github.com/hexley21/fixup/pkg/encryption/mock"
	"github.com/hexley21/fixup/pkg/hasher"
	mock_hasher "github.com/hexley21/fixup/pkg/hasher/mock"
	mock_cdn "github.com/hexley21/fixup/pkg/infra/cdn/mock"
	mock_postgres "github.com/hexley21/fixup/pkg/infra/postgres/mock"
	mock_mailer "github.com/hexley21/fixup/pkg/mailer/mock"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

var (
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

	creds = repository.GetCredentialsByEmailRow{
		ID: 1,
		Role: enum.UserRoleADMIN,
		Hash: newHash,
		UserStatus: pgtype.Bool{Bool: true, Valid: true},
	}

	mockEmailAddress = "fixup@gmail.com"
	newHash          = "hash"
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
	mockHasher := mock_hasher.NewMockHasher(ctrl)
	mockUrlSigner := mock_cdn.NewMockURLSigner(ctrl)

	mockUserRepo.EXPECT().CreateUser(ctx, gomock.Any()).Return(userEntity, nil)
	mockHasher.EXPECT().HashPassword(gomock.Any()).Return(newHash)
	mockUrlSigner.EXPECT().SignURL(gomock.Any()).Return(newUrl, nil)

	service := service.NewAuthService(mockUserRepo, nil, mockPgx, mockHasher, nil, nil, mockEmailAddress, mockUrlSigner)
	service.SetTemplates(confiramtionTemplate, nil)

	dto, err := service.RegisterCustomer(ctx, registerUserDto)
	assert.NoError(t, err)
	assert.Equal(t, registerUserDto.Email, dto.Email)
	assert.Equal(t, registerUserDto.PhoneNumber, dto.PhoneNumber)
	assert.Equal(t, registerUserDto.FirstName, dto.FirstName)
	assert.Equal(t, registerUserDto.LastName, dto.LastName)

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
	mockUrlSigner := mock_cdn.NewMockURLSigner(ctrl)

	mockUserRepo.EXPECT().WithTx(mockTx).Return(mockUserRepo)
	mockUserRepo.EXPECT().CreateUser(ctx, gomock.Any()).Return(userEntity, nil)
	mockProviderRepo.EXPECT().WithTx(mockTx).Return(mockProviderRepo)
	mockProviderRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

	mockPgx.EXPECT().BeginTx(ctx, gomock.Any()).Return(mockTx, nil)
	mockTx.EXPECT().Commit(ctx).Return(nil)

	mockHasher.EXPECT().HashPassword(gomock.Any()).Return(newHash)
	mockEncryptor.EXPECT().Encrypt(gomock.Any()).Return([]byte(registerProviderDto.PersonalIDNumber), nil)
	mockUrlSigner.EXPECT().SignURL(gomock.Any()).Return(newUrl, nil)

	service := service.NewAuthService(mockUserRepo, mockProviderRepo, mockPgx, mockHasher, mockEncryptor, nil, mockEmailAddress, mockUrlSigner)
	service.SetTemplates(confiramtionTemplate, nil)

	dto, err := service.RegisterProvider(ctx, registerProviderDto)
	assert.NoError(t, err)
	assert.Equal(t, registerUserDto.Email, dto.Email)
	assert.Equal(t, registerUserDto.PhoneNumber, dto.PhoneNumber)
	assert.Equal(t, registerUserDto.FirstName, dto.FirstName)
	assert.Equal(t, registerUserDto.LastName, dto.LastName)

	time.Sleep(100 * time.Millisecond)
}

func TestAuthenticateUser_Success(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockHasher := mock_hasher.NewMockHasher(ctrl)

	mockUserRepo.EXPECT().GetCredentialsByEmail(ctx, loginDto.Email).Return(creds, nil)
	mockHasher.EXPECT().VerifyPassword(loginDto.Password, creds.Hash).Return(nil)

	service := service.NewAuthService(mockUserRepo, nil, nil, mockHasher, nil, nil, mockEmailAddress, nil)

	credentialsDto, err := service.AuthenticateUser(ctx, loginDto)
	assert.NoError(t, err)
	assert.Equal(t, strconv.FormatInt(creds.ID, 10), credentialsDto.ID)
	assert.Equal(t, string(credentialsDto.Role), credentialsDto.Role)
	assert.Equal(t, creds.UserStatus.Bool, credentialsDto.UserStatus)
}

func TestAuthenticateUser_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().GetCredentialsByEmail(ctx, loginDto.Email).Return(repository.GetCredentialsByEmailRow{}, pgx.ErrNoRows)

	svc := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockEmailAddress, nil)

	credentialsDto, err := svc.AuthenticateUser(ctx, loginDto)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, credentialsDto)
}

func TestAuthenticateUser_PasswordMissmatch(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockHasher := mock_hasher.NewMockHasher(ctrl)

	mockUserRepo.EXPECT().GetCredentialsByEmail(ctx, loginDto.Email).Return(creds, nil)
	mockHasher.EXPECT().VerifyPassword(loginDto.Password, creds.Hash).Return(hasher.ErrPasswordMismatch)

	svc := service.NewAuthService(mockUserRepo, nil, nil, mockHasher, nil, nil, mockEmailAddress, nil)

	credentialsDto, err := svc.AuthenticateUser(ctx, loginDto)
	assert.ErrorIs(t, err, hasher.ErrPasswordMismatch)
	assert.Empty(t, credentialsDto)
}

func TestVerifyUser(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)
	mockUserRepo.EXPECT().UpdateStatus(ctx, gomock.Any()).Return(nil)

	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockEmailAddress, nil)
	service.SetTemplates(nil, verifiedTemplate)

	assert.NoError(t, service.VerifyUser(ctx, 1, ""))

	time.Sleep(100 * time.Millisecond)
}

func TestGetUserConfirmationDetails_Success(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)

	args := repository.GetUserConfirmationDetailsRow{
		ID: 1,
		UserStatus: pgtype.Bool{Bool: false, Valid: true},
		FirstName: "Larry",
	}

	mockUserRepo.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(args, nil)

	service := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockEmailAddress, nil)

	dto, err := service.GetUserConfirmationDetails(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, strconv.FormatInt(args.ID, 10), dto.ID)
	assert.Equal(t, args.UserStatus.Bool, dto.UserStatus)
	assert.Equal(t, args.FirstName, dto.Firstname)
}

func TestGetUserConfirmationDetails_ActiveUserError(t *testing.T) {
	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUserRepo := mock_repository.NewMockUserRepository(ctrl)

	args := repository.GetUserConfirmationDetailsRow{
		ID: 1,
		UserStatus: pgtype.Bool{Bool: true, Valid: true},
		FirstName: "Larry",
	}

	mockUserRepo.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(args, nil)

	svc := service.NewAuthService(mockUserRepo, nil, nil, nil, nil, nil, mockEmailAddress, nil)

	dto, err := svc.GetUserConfirmationDetails(ctx, "")
	assert.ErrorIs(t, err, service.ErrUserAlreadyActive)
	assert.NotEmpty(t, dto)
}

func TestSendConfirmationLetter_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMailer := mock_mailer.NewMockMailer(ctrl)
	mockMailer.EXPECT().SendHTML(mockEmailAddress, gomock.Any(), gomock.Any(), confiramtionTemplate, gomock.Any()).Return(nil)

	service := service.NewAuthService(nil, nil, nil, nil, nil, mockMailer, mockEmailAddress, nil)
	service.SetTemplates(confiramtionTemplate, nil)

	assert.NoError(t, service.SendConfirmationLetter("", "", ""))
}

func TestSendVerifiedLetter_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMailer := mock_mailer.NewMockMailer(ctrl)
	mockMailer.EXPECT().SendHTML(mockEmailAddress, gomock.Any(), gomock.Any(), verifiedTemplate, gomock.Nil()).Return(nil)

	service := service.NewAuthService(nil, nil, nil, nil, nil, mockMailer, mockEmailAddress, nil)
	service.SetTemplates(nil, verifiedTemplate)

	assert.NoError(t, service.SendVerifiedLetter(""))
}
