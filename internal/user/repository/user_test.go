package repository_test

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

var (
	userCreateArgs = repository.CreateUserParams{
		FirstName:   "test",
		LastName:    "test",
		PhoneNumber: "995555555555",
		Email:       "test@email.com",
		Hash:        "Ehx0DNg86zL6QCB8gMZxzkm0fPt3ObwhQzKAu22bnVYZvVe84GAAh8jFp5Cf47R5YncjKqQCyLakki78isy5899YTeVNjNjxK3N2EwdXGz4RB9YHkILLdfyT89DfAEtK",
		Role:        enum.UserRoleCUSTOMER,
	}

	invalidValue = "uwox71YgdFn6SuR4x971KjxrUaSoUdax9k0DkCt1WnzEHcdG9lpqEkF7RHw0SWUL"
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

func TestCreateWithInvalidArgs(t *testing.T) {
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
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

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

func TestCreateInvalidArgs(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	snowflakeNode := getSnowflakeNode()
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, snowflakeNode)

	invalidArgs := []repository.CreateUserParams{
		{FirstName: invalidValue, LastName: userCreateArgs.LastName, PhoneNumber: userCreateArgs.PhoneNumber, Email: userCreateArgs.Email, Hash: userCreateArgs.Hash, Role: userCreateArgs.Role},
		{FirstName: userCreateArgs.FirstName, LastName: invalidValue, PhoneNumber: userCreateArgs.PhoneNumber, Email: userCreateArgs.Email, Hash: userCreateArgs.Hash, Role: userCreateArgs.Role},
		{FirstName: userCreateArgs.FirstName, LastName: userCreateArgs.LastName, PhoneNumber: invalidValue, Email: userCreateArgs.Email, Hash: userCreateArgs.Hash, Role: userCreateArgs.Role},
		{FirstName: userCreateArgs.FirstName, LastName: userCreateArgs.LastName, PhoneNumber: userCreateArgs.PhoneNumber, Email: invalidValue, Hash: userCreateArgs.Hash, Role: userCreateArgs.Role},
		{FirstName: userCreateArgs.FirstName, LastName: userCreateArgs.LastName, PhoneNumber: userCreateArgs.PhoneNumber, Hash: invalidValue, Role: userCreateArgs.Role},
		{FirstName: userCreateArgs.FirstName, LastName: userCreateArgs.LastName, PhoneNumber: userCreateArgs.PhoneNumber, Hash: userCreateArgs.Hash, Role: enum.UserRole(invalidValue)},
	}

	i := 0
	for _, args := range invalidArgs {
		entity, err := repo.Create(ctx, args)
		if !assert.Error(t, err, i) {
			log.Println("create user:", i)
		}
		assert.Empty(t, entity)
		i++
	}
}

func TestGetHashById(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	hash, err := repo.GetHashById(ctx, insert.ID)
	assert.NoError(t, err)

	assert.Equal(t, hash, insert.Hash)
}

func TestGetPasswordHashFromNonexistentUser(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	hash, err := repo.GetHashById(ctx, 1)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, hash)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	firstName := "updated_firstname"
	lastName := "updated_lastname"
	phoneNumber := "995111111111"
	email := "updated@email.com"

	updateArgs := repository.UpdateUserParams{
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

func TestUpdateWithPartialArguments(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	firstName := "updated_firstname"
	lastName := "updated_lastname"

	updateArgs := repository.UpdateUserParams{
		ID:          insert.ID,
		FirstName:   &firstName,
		LastName:    &lastName,
	}

	update, err := repo.Update(ctx, updateArgs)
	assert.NoError(t, err)

	assert.Equal(t, update.FirstName, firstName)
	assert.Equal(t, update.LastName, lastName)
	assert.Equal(t, update.PhoneNumber, userCreateArgs.PhoneNumber)
	assert.Equal(t, update.Email, userCreateArgs.Email)
}

func TestUpdateWithNoArguments(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	_, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	update, err := repo.Update(ctx, repository.UpdateUserParams{ID: 1})
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, update)
}

func TestUpdateForNonexistentUser(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	update, err := repo.Update(ctx, repository.UpdateUserParams{ID: 1, FirstName: &userCreateArgs.FirstName})
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, update)
}

func TestUpdatePicture(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	var pictureArg pgtype.Text
	pictureArg.Scan("picture.jpg")
	args := repository.UpdateUserPictureParams{
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

func TestUpdatePictureForNonexistentUser(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	err := repo.UpdatePicture(ctx, repository.UpdateUserPictureParams{ID: 1, PictureName: pgtype.Text{}})
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestUpdateStatus(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	var statusArg pgtype.Bool
	statusArg.Scan(true)
	args := repository.UpdateUserStatusParams{
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

func TestUpdateStatusForNonexistentUser(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	err := repo.UpdateStatus(ctx, repository.UpdateUserStatusParams{ID: 1, UserStatus: pgtype.Bool{}})
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestUpdateHash(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	args := repository.UpdateUserHashParams{
		ID:   insert.ID,
		Hash: "yT89DfAEtKL6QCB8gMZxzkm0fPt3ObwhQzKAu22bnVYZvVe84GAAh8jFp5Cf47R5YncjKqQCyLakki78isy5899YTeVNjNjxK3N2EwdXGz4RB9YHkILLdfEhx0DNg86z",
	}
	err = repo.UpdateHash(ctx, args)
	assert.NoError(t, err)

	row := dbPool.QueryRow(ctx, "SELECT hash from users where id = $1", insert.ID)
	var updatedPassword string
	err = row.Scan(&updatedPassword)
	assert.NoError(t, err)

	assert.Equal(t, args.Hash, updatedPassword)
}

func TestUpdateHashWithInvalidArgs(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to insert user: %v", err))
	}

	err = repo.UpdateHash(ctx, repository.UpdateUserHashParams{
		ID:   insert.ID,
		Hash: "abc",
	})
	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgErr.Code, pgerrcode.CheckViolation)
	}
}

func TestUpdateHashForNonexistentUser(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	err := repo.UpdateHash(ctx, repository.UpdateUserHashParams{ID: 1, Hash: "abc"})
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestDeleteById(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

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

func TestDeleteByIdWithNonexistendUser(t *testing.T) {
	ctx := context.Background()
	dbPool := getDbPool(ctx)
	setupDatabaseCleanup(t, ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	err := repo.DeleteById(ctx, 1)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func insertUser(dbPool *pgxpool.Pool, ctx context.Context, args repository.CreateUserParams, id int64) (entity.User, error) {
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
