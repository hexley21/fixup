package service_test

import (
	"context"
	"errors"
	"html/template"
	"strconv"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1/dto"
	"github.com/hexley21/fixup/internal/user/repository"
	mockRepository "github.com/hexley21/fixup/internal/user/repository/mock"
	"github.com/hexley21/fixup/internal/user/service"
	mockEncryption "github.com/hexley21/fixup/pkg/encryption/mock"
	"github.com/hexley21/fixup/pkg/hasher"
	mockHasher "github.com/hexley21/fixup/pkg/hasher/mock"
	mockCdn "github.com/hexley21/fixup/pkg/infra/cdn/mock"
	mockPostgres "github.com/hexley21/fixup/pkg/infra/postgres/mock"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
	mockMailer "github.com/hexley21/fixup/pkg/mailer/mock"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
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
		RegisterUser:     registerUserDto,
		PersonalIDNumber: "1234567890",
	}

	loginDto = dto.Login{
		Email:    "larry@page.com",
		Password: "12345678",
	}

	creds = repository.GetCredentialsByEmailRow{
		ID:         1,
		Role:       enum.UserRoleADMIN,
		Hash:       newHash,
		UserStatus: pgtype.Bool{Bool: true, Valid: true},
	}

	emailAddress = "fixup@gmail.com"
	newHash      = "hash"
	newUrl       = "picture.png?signed=true"

	confiramtionTemplate = template.New("confirmation")
	verifiedTemplate     = template.New("verified")

	verificationTTL = time.Hour
	vrfToken        = "poUnbbjqcnpaDBbK8nQfbxrx0ZJZBbFR"
)

func setupAuth(t *testing.T) (
	ctrl *gomock.Controller,
	svc service.AuthService,
	userRepoMock *mockRepository.MockUserRepository,
	providerRepoMock *mockRepository.MockProviderRepository,
	vrfRepoMock *mockRepository.MockVerificationRepository,
	pgxMock *mockPostgres.MockPGX,
	txMock *mockPostgres.MockTx,
	hasherMock *mockHasher.MockHasher,
	encryptorMock *mockEncryption.MockEncryptor,
	mailerMock *mockMailer.MockMailer,
	urlSignerMock *mockCdn.MockURLSigner,
) {
	ctrl = gomock.NewController(t)

	userRepoMock = mockRepository.NewMockUserRepository(ctrl)
	providerRepoMock = mockRepository.NewMockProviderRepository(ctrl)
	vrfRepoMock = mockRepository.NewMockVerificationRepository(ctrl)
	pgxMock = mockPostgres.NewMockPGX(ctrl)
	txMock = mockPostgres.NewMockTx(ctrl)
	hasherMock = mockHasher.NewMockHasher(ctrl)
	encryptorMock = mockEncryption.NewMockEncryptor(ctrl)
	mailerMock = mockMailer.NewMockMailer(ctrl)
	urlSignerMock = mockCdn.NewMockURLSigner(ctrl)

	s := service.NewAuthService(userRepoMock, providerRepoMock, vrfRepoMock, verificationTTL, pgxMock, hasherMock, encryptorMock, mailerMock, emailAddress, urlSignerMock)
	s.SetTemplates(confiramtionTemplate, verifiedTemplate)

	svc = s
	return
}

func TestRegisterCustomer(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, userRepoMock, _, _, _, _, hasherMock, _, _, urlSignerMock := setupAuth(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().CreateUser(ctx, gomock.Any()).Return(userEntity, nil)
	hasherMock.EXPECT().HashPassword(gomock.Any()).Return(newHash, nil)
	urlSignerMock.EXPECT().SignURL(gomock.Any()).Return(newUrl, nil)

	registerDTO, err := svc.RegisterCustomer(ctx, registerUserDto)
	assert.NoError(t, err)
	assert.Equal(t, registerUserDto.Email, registerDTO.Email)
	assert.Equal(t, registerUserDto.PhoneNumber, registerDTO.PhoneNumber)
	assert.Equal(t, registerUserDto.FirstName, registerDTO.FirstName)
	assert.Equal(t, registerUserDto.LastName, registerDTO.LastName)

	time.Sleep(time.Microsecond)
}

func TestRegisterProvider(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, userRepoMock, providerRepoMock, _, pgxMock, txMock, hasherMock, encryptorMock, _, urlSignerMock := setupAuth(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().WithTx(txMock).Return(userRepoMock)
	userRepoMock.EXPECT().CreateUser(ctx, gomock.Any()).Return(userEntity, nil)
	providerRepoMock.EXPECT().WithTx(txMock).Return(providerRepoMock)
	providerRepoMock.EXPECT().Create(ctx, gomock.Any()).Return(nil)

	pgxMock.EXPECT().BeginTx(ctx, gomock.Any()).Return(txMock, nil)
	txMock.EXPECT().Commit(ctx).Return(nil)

	hasherMock.EXPECT().HashPassword(gomock.Any()).Return(newHash, nil)
	encryptorMock.EXPECT().Encrypt(gomock.Any()).Return([]byte(registerProviderDto.PersonalIDNumber), nil)
	urlSignerMock.EXPECT().SignURL(gomock.Any()).Return(newUrl, nil)

	registerDTO, err := svc.RegisterProvider(ctx, registerProviderDto)
	assert.NoError(t, err)
	assert.Equal(t, registerUserDto.Email, registerDTO.Email)
	assert.Equal(t, registerUserDto.PhoneNumber, registerDTO.PhoneNumber)
	assert.Equal(t, registerUserDto.FirstName, registerDTO.FirstName)
	assert.Equal(t, registerUserDto.LastName, registerDTO.LastName)

	time.Sleep(time.Microsecond)
}

func TestAuthenticateUser_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, userRepoMock, _, _, _, _, hasherMock, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetCredentialsByEmail(ctx, loginDto.Email).Return(creds, nil)
	hasherMock.EXPECT().VerifyPassword(loginDto.Password, creds.Hash).Return(nil)

	credentialsDto, err := svc.AuthenticateUser(ctx, loginDto)
	assert.NoError(t, err)
	assert.Equal(t, strconv.FormatInt(creds.ID, 10), credentialsDto.ID)
	assert.Equal(t, credentialsDto.Role, credentialsDto.Role)
	assert.Equal(t, creds.UserStatus.Bool, credentialsDto.UserStatus)
}

func TestAuthenticateUser_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, userRepoMock, _, _, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetCredentialsByEmail(ctx, loginDto.Email).Return(repository.GetCredentialsByEmailRow{}, pgx.ErrNoRows)

	credentialsDto, err := svc.AuthenticateUser(ctx, loginDto)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, credentialsDto)
}

func TestAuthenticateUser_PasswordMissmatch(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, userRepoMock, _, _, _, _, hasherMock, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	userRepoMock.EXPECT().GetCredentialsByEmail(ctx, loginDto.Email).Return(creds, nil)
	hasherMock.EXPECT().VerifyPassword(loginDto.Password, creds.Hash).Return(hasher.ErrPasswordMismatch)

	credentialsDto, err := svc.AuthenticateUser(ctx, loginDto)
	assert.ErrorIs(t, err, hasher.ErrPasswordMismatch)
	assert.Empty(t, credentialsDto)
}

func TestVerifyUser_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, userRepoMock, _, vrfRepoMock, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	vrfRepoMock.EXPECT().SetTokenUsed(ctx, vrfToken, verificationTTL).Return(nil)
	userRepoMock.EXPECT().UpdateStatus(ctx, gomock.Any()).Return(nil)

	assert.NoError(t, svc.VerifyUser(ctx, vrfToken, verificationTTL, 1))
}

func TestVerifyUser_AlreadySet(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, _, _, vrfRepoMock, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	vrfRepoMock.EXPECT().SetTokenUsed(ctx, vrfToken, verificationTTL).Return(redis.TxFailedErr)

	assert.ErrorIs(t, redis.TxFailedErr, svc.VerifyUser(ctx, vrfToken, verificationTTL, 1))
}

func TestVerifyUser_RepoError(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, userRepoMock, _, vrfRepoMock, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	vrfRepoMock.EXPECT().SetTokenUsed(ctx, vrfToken, verificationTTL).Return(nil)
	userRepoMock.EXPECT().UpdateStatus(ctx, gomock.Any()).Return(errors.New(""))

	assert.Error(t, svc.VerifyUser(ctx, vrfToken, verificationTTL, 1))
}

func TestVerifyUser_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, userRepoMock, _, vrfRepoMock, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	vrfRepoMock.EXPECT().SetTokenUsed(ctx, vrfToken, verificationTTL).Return(nil)
	userRepoMock.EXPECT().UpdateStatus(ctx, gomock.Any()).Return(pg_error.ErrNotFound)

	assert.ErrorIs(t, pg_error.ErrNotFound, svc.VerifyUser(ctx, vrfToken, verificationTTL, 1))
}

func TestGetUserConfirmationDetails_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, userRepoMock, _, _, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	args := repository.GetUserConfirmationDetailsRow{
		ID:         1,
		UserStatus: pgtype.Bool{Bool: false, Valid: true},
		FirstName:  "Larry",
	}

	userRepoMock.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(args, nil)

	confirmationDTO, err := svc.GetUserConfirmationDetails(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, strconv.FormatInt(args.ID, 10), confirmationDTO.ID)
	assert.Equal(t, args.UserStatus.Bool, confirmationDTO.UserStatus)
	assert.Equal(t, args.FirstName, confirmationDTO.Firstname)
}

func TestSendConfirmationLetter_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, _, _, vrfRepoMock, _, _, _, _, mailerMock, _ := setupAuth(t)
	defer ctrl.Finish()

	vrfRepoMock.EXPECT().IsTokenUsed(ctx, gomock.Any()).Return(false, nil)
	mailerMock.EXPECT().SendHTML(emailAddress, gomock.Any(), gomock.Any(), confiramtionTemplate, gomock.Any()).Return(nil)

	assert.NoError(t, svc.SendConfirmationLetter(ctx, "", "", ""))
}

func TestSendConfirmationLetter_AlreadySet(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, _, _, vrfRepoMock, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	vrfRepoMock.EXPECT().IsTokenUsed(ctx, gomock.Any()).Return(true, nil)

	assert.ErrorIs(t, service.ErrAlreadyVerified, svc.SendConfirmationLetter(ctx, "", "", ""))
}

func TestSendConfirmationLetter_RepoError(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, _, _, vrfRepoMock, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	vrfRepoMock.EXPECT().IsTokenUsed(ctx, gomock.Any()).Return(false, errors.New(""))

	assert.Error(t, svc.SendConfirmationLetter(ctx, "", "", ""))
}

func TestSendVerifiedLetter_Success(t *testing.T) {
	ctrl, svc, _, _, _, _, _, _, _, mailerMock, _ := setupAuth(t)
	defer ctrl.Finish()

	mailerMock.EXPECT().SendHTML(emailAddress, gomock.Any(), gomock.Any(), verifiedTemplate, gomock.Nil()).Return(nil)

	assert.NoError(t, svc.SendVerifiedLetter(""))
}
