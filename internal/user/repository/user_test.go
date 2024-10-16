package repository_test

import (
	"context"
	"log"
	"testing"

	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/hexley21/fixup/pkg/infra/postgres/pg_error"
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

func TestGetById_Success(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	userEntity, err := repo.GetById(ctx, insert.ID)
	assert.NoError(t, err)

	assert.Equal(t, insert.ID, userEntity.ID)
	assert.Equal(t, insert.FirstName, userEntity.FirstName)
	assert.Equal(t, insert.LastName, userEntity.LastName)
	assert.Equal(t, insert.PhoneNumber, userEntity.PhoneNumber)
	assert.Equal(t, insert.Email, userEntity.Email)
	assert.Equal(t, insert.Role, userEntity.Role)
	assert.Equal(t, insert.UserStatus, userEntity.UserStatus)
	assert.Equal(t, insert.CreatedAt, userEntity.CreatedAt)
	assert.Equal(t, insert.PictureName, userEntity.PictureName)

	assert.Empty(t, userEntity.Hash)
}

func TestCreate_Success(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	snowflakeNode := getSnowflakeNode()
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, snowflakeNode)

	userEntity, err := repo.CreateUser(ctx, userCreateArgs)
	assert.NoError(t, err)

	assert.Equal(t, userCreateArgs.FirstName, userEntity.FirstName)
	assert.Equal(t, userCreateArgs.LastName, userEntity.LastName)
	assert.Equal(t, userCreateArgs.PhoneNumber, userEntity.PhoneNumber)
	assert.Equal(t, userCreateArgs.Email, userEntity.Email)
	assert.Equal(t, userCreateArgs.Role, userEntity.Role)
	assert.Equal(t, false, userEntity.UserStatus.Bool)

	assert.NotEqual(t, 0, userEntity.ID)
	assert.NotEmpty(t, userEntity.CreatedAt)

	assert.Empty(t, userEntity.PictureName)
	assert.Empty(t, userEntity.Hash)
}

func TestCreate_InvalidArgs(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	snowflakeNode := getSnowflakeNode()
	defer cleanupPostgres(ctx, dbPool)

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
		userEntity, err := repo.CreateUser(ctx, args)
		if !assert.Error(t, err) {
			log.Println("create user:", i)
		}
		assert.Empty(t, userEntity)
		i++
	}
}

func TestGetCredentialsByEmail_Success(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	creds, err := repo.GetCredentialsByEmail(ctx, insert.Email)
	assert.NoError(t, err)

	assert.Equal(t, insert.ID, creds.ID)
	assert.Equal(t, insert.Role, creds.Role)
	assert.Equal(t, insert.Hash, creds.Hash)
	assert.Equal(t, insert.UserStatus, creds.UserStatus)
}

func TestGetCredentialsByEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	creds, err := repo.GetCredentialsByEmail(ctx, "email")
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, creds)
}

func TestGetHashById_Success(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	hash, err := repo.GetHashById(ctx, insert.ID)
	assert.NoError(t, err)

	assert.Equal(t, insert.Hash, hash)
}

func TestGetHash_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	hash, err := repo.GetHashById(ctx, 1)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, hash)
}

func TestUpdate_Success(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
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

	assert.Equal(t, firstName, update.FirstName)
	assert.Equal(t, lastName, update.LastName)
	assert.Equal(t, phoneNumber, update.PhoneNumber)
	assert.Equal(t, email, update.Email)
}

func TestUpdate_PartialArguments(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	firstName := "updated_firstname"
	lastName := "updated_lastname"

	updateArgs := repository.UpdateUserParams{
		ID:        insert.ID,
		FirstName: &firstName,
		LastName:  &lastName,
	}

	update, err := repo.Update(ctx, updateArgs)
	assert.NoError(t, err)

	assert.Equal(t, firstName, update.FirstName)
	assert.Equal(t, lastName, update.LastName)
	assert.Equal(t, userCreateArgs.PhoneNumber, update.PhoneNumber)
	assert.Equal(t, userCreateArgs.Email, update.Email)
}

func TestUpdate_NoArguments(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	_, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	update, err := repo.Update(ctx, repository.UpdateUserParams{ID: 1})
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, update)
}

func TestUpdate_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	update, err := repo.Update(ctx, repository.UpdateUserParams{ID: 1, FirstName: &userCreateArgs.FirstName})
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, update)
}

func TestUpdatePicture_Success(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	pictureArg := pgtype.Text{String: "picture.jpg", Valid: true}
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

	assert.Equal(t, updatedPicture, pictureArg)
}

func TestUpdatePicture_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	err := repo.UpdatePicture(ctx, repository.UpdateUserPictureParams{ID: 1, PictureName: pgtype.Text{}})
	assert.ErrorIs(t, err, pg_error.ErrNotFound)
}

func TestUpdateStatus_Success(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
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

	assert.Equal(t, updatedStatus, statusArg)
}

func TestUpdateStatus_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	err := repo.UpdateStatus(ctx, repository.UpdateUserStatusParams{ID: 1, UserStatus: pgtype.Bool{}})
	assert.ErrorIs(t, err, pg_error.ErrNotFound)
}

func TestUpdateHash_Success(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	args := repository.UpdateUserHashParams{
		ID:   insert.ID,
		Hash: "yT89DfAEtKL6QCB8gMZxzkm0fPt3ObwhQzKAu22bnVYZvVe84GAAh8jFp5Cf47R5YncjKqQCyLakki78isy5899YTeVNjNjxK3N2EwdXGz4RB9YHkILLdfEhx0DNg86z",
	}
	err = repo.UpdateHash(ctx, args)
	assert.NoError(t, err)

	row := dbPool.QueryRow(ctx, "SELECT hash from users where id = $1", insert.ID)
	var updatedHash string
	err = row.Scan(&updatedHash)
	assert.NoError(t, err)

	assert.Equal(t, args.Hash, updatedHash)
}

func TestUpdateHash_InvalidArguments(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	err = repo.UpdateHash(ctx, repository.UpdateUserHashParams{
		ID:   insert.ID,
		Hash: "abc",
	})
	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgerrcode.CheckViolation, pgErr.Code)
	}
}

func TestUpdateHash_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	err := repo.UpdateHash(ctx, repository.UpdateUserHashParams{ID: 1, Hash: "abc"})
	assert.ErrorIs(t, err, pg_error.ErrNotFound)
}

func TestDeleteById_Success(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	err = repo.DeleteById(ctx, insert.ID)
	assert.NoError(t, err)

	row := dbPool.QueryRow(ctx, "SELECT * FROM users WHERE id = $1", insert.ID)
	var userId int64
	err = row.Scan(userId)
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}

func TestDeleteById_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	err := repo.DeleteById(ctx, 1)
	assert.ErrorIs(t, err, pg_error.ErrNotFound)
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
