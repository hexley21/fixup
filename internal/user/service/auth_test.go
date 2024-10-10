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
	mock_repository "github.com/hexley21/fixup/internal/user/repository/mock"
	"github.com/hexley21/fixup/internal/user/service"
	mock_encryption "github.com/hexley21/fixup/pkg/encryption/mock"
	"github.com/hexley21/fixup/pkg/hasher"
	mock_hasher "github.com/hexley21/fixup/pkg/hasher/mock"
	mock_cdn "github.com/hexley21/fixup/pkg/infra/cdn/mock"
	mock_postgres "github.com/hexley21/fixup/pkg/infra/postgres/mock"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
	mock_mailer "github.com/hexley21/fixup/pkg/mailer/mock"
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
	mockUserRepo *mock_repository.MockUserRepository,
	mockProviderRepo *mock_repository.MockProviderRepository,
	mockVerificationRepo *mock_repository.MockVerificationRepository,
	mockPgx *mock_postgres.MockPGX,
	mockTx *mock_postgres.MockTx,
	mockHasher *mock_hasher.MockHasher,
	mockEncryptor *mock_encryption.MockEncryptor,
	mockMailer *mock_mailer.MockMailer,
	mockUrlSigner *mock_cdn.MockURLSigner,
) {
	ctrl = gomock.NewController(t)

	mockUserRepo = mock_repository.NewMockUserRepository(ctrl)
	mockProviderRepo = mock_repository.NewMockProviderRepository(ctrl)
	mockVerificationRepo = mock_repository.NewMockVerificationRepository(ctrl)
	mockPgx = mock_postgres.NewMockPGX(ctrl)
	mockTx = mock_postgres.NewMockTx(ctrl)
	mockHasher = mock_hasher.NewMockHasher(ctrl)
	mockEncryptor = mock_encryption.NewMockEncryptor(ctrl)
	mockMailer = mock_mailer.NewMockMailer(ctrl)
	mockUrlSigner = mock_cdn.NewMockURLSigner(ctrl)

	s := service.NewAuthService(mockUserRepo, mockProviderRepo, mockVerificationRepo, verificationTTL, mockPgx, mockHasher, mockEncryptor, mockMailer, emailAddress, mockUrlSigner)
	s.SetTemplates(confiramtionTemplate, verifiedTemplate)

	svc = s
	return
}

func TestRegisterCustomer(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, mockUserRepo, _, _, _, _, mockHasher, _, _, mockUrlSigner := setupAuth(t)
	defer ctrl.Finish()

	mockUserRepo.EXPECT().CreateUser(ctx, gomock.Any()).Return(userEntity, nil)
	mockHasher.EXPECT().HashPassword(gomock.Any()).Return(newHash, nil)
	mockUrlSigner.EXPECT().SignURL(gomock.Any()).Return(newUrl, nil)

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

	ctrl, svc, mockUserRepo, mockProviderRepo, _, mockPgx, mockTx, mockHasher, mockEncryptor, _, mockUrlSigner := setupAuth(t)
	defer ctrl.Finish()

	mockUserRepo.EXPECT().WithTx(mockTx).Return(mockUserRepo)
	mockUserRepo.EXPECT().CreateUser(ctx, gomock.Any()).Return(userEntity, nil)
	mockProviderRepo.EXPECT().WithTx(mockTx).Return(mockProviderRepo)
	mockProviderRepo.EXPECT().Create(ctx, gomock.Any()).Return(nil)

	mockPgx.EXPECT().BeginTx(ctx, gomock.Any()).Return(mockTx, nil)
	mockTx.EXPECT().Commit(ctx).Return(nil)

	mockHasher.EXPECT().HashPassword(gomock.Any()).Return(newHash, nil)
	mockEncryptor.EXPECT().Encrypt(gomock.Any()).Return([]byte(registerProviderDto.PersonalIDNumber), nil)
	mockUrlSigner.EXPECT().SignURL(gomock.Any()).Return(newUrl, nil)

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

	ctrl, svc, mockUserRepo, _, _, _, _, mockHasher, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	mockUserRepo.EXPECT().GetCredentialsByEmail(ctx, loginDto.Email).Return(creds, nil)
	mockHasher.EXPECT().VerifyPassword(loginDto.Password, creds.Hash).Return(nil)

	//svc := svc.NewAuthService(mockUserRepo, nil, nil, time.Hour, nil, mockHasher, nil, nil, emailAddress, nil)

	credentialsDto, err := svc.AuthenticateUser(ctx, loginDto)
	assert.NoError(t, err)
	assert.Equal(t, strconv.FormatInt(creds.ID, 10), credentialsDto.ID)
	assert.Equal(t, credentialsDto.Role, credentialsDto.Role)
	assert.Equal(t, creds.UserStatus.Bool, credentialsDto.UserStatus)
}

func TestAuthenticateUser_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, mockUserRepo, _, _, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	mockUserRepo.EXPECT().GetCredentialsByEmail(ctx, loginDto.Email).Return(repository.GetCredentialsByEmailRow{}, pgx.ErrNoRows)

	credentialsDto, err := svc.AuthenticateUser(ctx, loginDto)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, credentialsDto)
}

func TestAuthenticateUser_PasswordMissmatch(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, mockUserRepo, _, _, _, _, mockHasher, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	mockUserRepo.EXPECT().GetCredentialsByEmail(ctx, loginDto.Email).Return(creds, nil)
	mockHasher.EXPECT().VerifyPassword(loginDto.Password, creds.Hash).Return(hasher.ErrPasswordMismatch)

	credentialsDto, err := svc.AuthenticateUser(ctx, loginDto)
	assert.ErrorIs(t, err, hasher.ErrPasswordMismatch)
	assert.Empty(t, credentialsDto)
}

func TestVerifyUser_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, mockUserRepo, _, mockVerificationRepo, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	mockVerificationRepo.EXPECT().SetTokenUsed(ctx, vrfToken, verificationTTL).Return(nil)
	mockUserRepo.EXPECT().UpdateStatus(ctx, gomock.Any()).Return(nil)

	assert.NoError(t, svc.VerifyUser(ctx, vrfToken, verificationTTL, 1))
}

func TestVerifyUser_AlreadySet(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, _, _, mockVerificationRepo, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	mockVerificationRepo.EXPECT().SetTokenUsed(ctx, vrfToken, verificationTTL).Return(redis.TxFailedErr)

	assert.ErrorIs(t, redis.TxFailedErr, svc.VerifyUser(ctx, vrfToken, verificationTTL, 1))
}

func TestVerifyUser_RepoError(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, mockUserRepo, _, mockVerificationRepo, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	mockVerificationRepo.EXPECT().SetTokenUsed(ctx, vrfToken, verificationTTL).Return(nil)
	mockUserRepo.EXPECT().UpdateStatus(ctx, gomock.Any()).Return(errors.New(""))

	assert.Error(t, svc.VerifyUser(ctx, vrfToken, verificationTTL, 1))
}

func TestVerifyUser_NotFound(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, mockUserRepo, _, mockVerificationRepo, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	mockVerificationRepo.EXPECT().SetTokenUsed(ctx, vrfToken, verificationTTL).Return(nil)
	mockUserRepo.EXPECT().UpdateStatus(ctx, gomock.Any()).Return(pg_error.ErrNotFound)

	assert.ErrorIs(t, pg_error.ErrNotFound, svc.VerifyUser(ctx, vrfToken, verificationTTL, 1))
}

func TestGetUserConfirmationDetails_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, mockUserRepo, _, _, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	args := repository.GetUserConfirmationDetailsRow{
		ID:         1,
		UserStatus: pgtype.Bool{Bool: false, Valid: true},
		FirstName:  "Larry",
	}

	mockUserRepo.EXPECT().GetUserConfirmationDetails(ctx, gomock.Any()).Return(args, nil)

	confirmationDTO, err := svc.GetUserConfirmationDetails(ctx, "")
	assert.NoError(t, err)
	assert.Equal(t, strconv.FormatInt(args.ID, 10), confirmationDTO.ID)
	assert.Equal(t, args.UserStatus.Bool, confirmationDTO.UserStatus)
	assert.Equal(t, args.FirstName, confirmationDTO.Firstname)
}

func TestSendConfirmationLetter_Success(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, _, _, mockVerificationRepo, _, _, _, _, mockMailer, _ := setupAuth(t)
	defer ctrl.Finish()

	mockVerificationRepo.EXPECT().IsTokenUsed(ctx, gomock.Any()).Return(false, nil)
	mockMailer.EXPECT().SendHTML(emailAddress, gomock.Any(), gomock.Any(), confiramtionTemplate, gomock.Any()).Return(nil)

	assert.NoError(t, svc.SendConfirmationLetter(ctx, "", "", ""))
}

func TestSendConfirmationLetter_AlreadySet(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, _, _, mockVerificationRepo, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	mockVerificationRepo.EXPECT().IsTokenUsed(ctx, gomock.Any()).Return(true, nil)

	assert.ErrorIs(t, service.ErrAlreadyVerified, svc.SendConfirmationLetter(ctx, "", "", ""))
}

func TestSendConfirmationLetter_RepoError(t *testing.T) {
	ctx := context.Background()

	ctrl, svc, _, _, mockVerificationRepo, _, _, _, _, _, _ := setupAuth(t)
	defer ctrl.Finish()

	mockVerificationRepo.EXPECT().IsTokenUsed(ctx, gomock.Any()).Return(false, errors.New(""))

	assert.Error(t, svc.SendConfirmationLetter(ctx, "", "", ""))
}

func TestSendVerifiedLetter_Success(t *testing.T) {
	ctrl, svc, _, _, _, _, _, _, _, mockMailer, _ := setupAuth(t)
	defer ctrl.Finish()

	mockMailer.EXPECT().SendHTML(emailAddress, gomock.Any(), gomock.Any(), verifiedTemplate, gomock.Nil()).Return(nil)

	assert.NoError(t, svc.SendVerifiedLetter(""))
}
