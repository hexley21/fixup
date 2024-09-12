package repository_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

var (
	userCreateArgs = repository.CreateParams{
		FirstName:   "test",
		LastName:    "test",
		PhoneNumber: "995555555555",
		Email:       "test@email.com",
		Hash:        "Ehx0DNg86zL6QCB8gMZxzkm0fPt3ObwhQzKAu22bnVYZvVe84GAAh8jFp5Cf47R5YncjKqQCyLakki78isy5899YTeVNjNjxK3N2EwdXGz4RB9YHkILLdfyT89DfAEtK",
		Role:        enum.UserRoleCUSTOMER,
	}
)

func TestCreate(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	snowflakeNode := getSnowflakeNode()
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, snowflakeNode)

	entity, err := repo.Create(ctx, userCreateArgs)
	assert.NoError(t, err)

	assert.Equal(t, entity.FirstName, userCreateArgs.FirstName)
	assert.Equal(t, entity.LastName, userCreateArgs.LastName)
	assert.Equal(t, entity.PhoneNumber, userCreateArgs.PhoneNumber)
	assert.Equal(t, entity.Email, userCreateArgs.Email)
	assert.Equal(t, entity.Role, userCreateArgs.Role)
	assert.Equal(t, entity.UserStatus.Bool, false)

	assert.NotEqual(t, entity.ID, 0)
	assert.NotEmpty(t, entity.CreatedAt)

	assert.Empty(t, entity.PictureName)
	assert.Empty(t, entity.Hash)
}

func TestGetById(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	snowflakeNode := getSnowflakeNode()
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, snowflakeNode)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	entity, err := repo.GetById(ctx, insert.ID)
	assert.NoError(t, err)

	assert.Equal(t, entity.ID, insert.ID)
	assert.Equal(t, entity.FirstName, insert.FirstName)
	assert.Equal(t, entity.LastName, insert.LastName)
	assert.Equal(t, entity.PhoneNumber, insert.PhoneNumber)
	assert.Equal(t, entity.Email, insert.Email)
	assert.Equal(t, entity.Role, insert.Role)
	assert.Equal(t, entity.UserStatus, insert.UserStatus)
	assert.Equal(t, entity.CreatedAt, insert.CreatedAt)
	assert.Equal(t, entity.PictureName, insert.PictureName)

	assert.Empty(t, entity.Hash)
}

func TestGetCredentialsByEmail(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	snowflakeNode := getSnowflakeNode()
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, snowflakeNode)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	credentials, err := repo.GetCredentialsByEmail(ctx, insert.Email)
	assert.NoError(t, err)

	assert.Equal(t, credentials.Hash, insert.Hash)
	assert.Equal(t, credentials.ID, insert.ID)
	assert.Equal(t, credentials.Role, insert.Role)
}

func TestGetPasswordHashById(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	snowflakeNode := getSnowflakeNode()
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, snowflakeNode)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	hash, err := repo.GetPasswordHashById(ctx, insert.ID)
	assert.NoError(t, err)

	assert.Equal(t, hash, insert.Hash)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	snowflakeNode := getSnowflakeNode()
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, snowflakeNode)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	firstName := "updated_firstname"
	lastName := "updated_lastname"
	phoneNumber := "995111111111"
	email := "updated@email.com"

	updateArgs := repository.UpdateParams{
		ID:          insert.ID,
		FirstName:   &firstName,
		LastName:    &lastName,
		PhoneNumber: &phoneNumber,
		Email:       &email,
	}

	update, err := repo.Update(ctx, updateArgs)
	assert.NoError(t, err)

	assert.Equal(t, update.FirstName, firstName)
	assert.Equal(t, update.LastName, lastName)
	assert.Equal(t, update.PhoneNumber, phoneNumber)
	assert.Equal(t, update.Email, email)
}

func TestUpdatePicture(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	snowflakeNode := getSnowflakeNode()
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, snowflakeNode)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	var pictureArg pgtype.Text
	pictureArg.Scan("picture.jpg")
	args := repository.UpdatePictureParams{
		ID:          insert.ID,
		PictureName: pictureArg,
	}
	err = repo.UpdatePicture(ctx, args)
	assert.NoError(t, err)

	row := dbPool.QueryRow(ctx, "SELECT picture_name from users where id = $1", insert.ID)
	var updatedPicture pgtype.Text
	err = row.Scan(&updatedPicture)
	assert.NoError(t, err)

	assert.Equal(t, pictureArg, updatedPicture)
}

func TestUpdateStatus(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	snowflakeNode := getSnowflakeNode()
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, snowflakeNode)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	var statusArg pgtype.Bool
	statusArg.Scan(true)
	args := repository.UpdateStatusParams{
		ID:         insert.ID,
		UserStatus: statusArg,
	}
	err = repo.UpdateStatus(ctx, args)
	assert.NoError(t, err)

	row := dbPool.QueryRow(ctx, "SELECT user_status from users where id = $1", insert.ID)
	var updatedStatus pgtype.Bool
	err = row.Scan(&updatedStatus)
	assert.NoError(t, err)

	assert.Equal(t, statusArg, updatedStatus)
}

func TestUpdatePassword(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	snowflakeNode := getSnowflakeNode()
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, snowflakeNode)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	args := repository.UpdatePasswordParams{
		ID:   insert.ID,
		Hash: "yT89DfAEtKL6QCB8gMZxzkm0fPt3ObwhQzKAu22bnVYZvVe84GAAh8jFp5Cf47R5YncjKqQCyLakki78isy5899YTeVNjNjxK3N2EwdXGz4RB9YHkILLdfEhx0DNg86z",
	}
	err = repo.UpdatePassword(ctx, args)
	assert.NoError(t, err)

	row := dbPool.QueryRow(ctx, "SELECT hash from users where id = $1", insert.ID)
	var updatedPassword string
	err = row.Scan(&updatedPassword)
	assert.NoError(t, err)

	assert.Equal(t, args.Hash, updatedPassword)
}

func TestDeleteById(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	snowflakeNode := getSnowflakeNode()
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, snowflakeNode)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	err = repo.DeleteById(ctx, insert.ID)
	assert.NoError(t, err)

	row := dbPool.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", insert.ID)
	var userId int64
	err = row.Scan(userId)
	assert.Error(t, err, pgx.ErrNoRows)
}

func insertUser(dbPool *pgxpool.Pool, ctx context.Context, args repository.CreateParams, id int64) (entity.User, error) {
	row := dbPool.QueryRow(
		ctx,
		"INSERT INTO users (id, first_name, last_name, phone_number, email, hash, role) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *",
		id,
		args.FirstName,
		args.LastName,
		args.PhoneNumber,
		args.Email,
		args.Hash,
		args.Role,
	)
	var i entity.User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.PhoneNumber,
		&i.Email,
		&i.PictureName,
		&i.Hash,
		&i.Role,
		&i.UserStatus,
		&i.CreatedAt,
	)
	return i, err
}
