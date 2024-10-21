package repository_test

import (
	"context"
	"log"
	"testing"

	"github.com/hexley21/fixup/internal/common/enum"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

// TODO: Convert tests to table driven & test get user picture 

var (
	hashArg        = "Ehx0DNg86zL6QCB8gMZxzkm0fPt3ObwhQzKAu22bnVYZvVe84GAAh8jFp5Cf47R5YncjKqQCyLakki78isy5899YTeVNjNjxK3N2EwdXGz4RB9YHkILLdfyT89DfAEtK"
	userCreateArgs = repository.CreateUserParams{
		FirstName:   "test",
		LastName:    "test",
		PhoneNumber: "995555555555",
		Email:       "test@email.com",
		Hash:        hashArg,
		Role:        string(enum.UserRoleCUSTOMER),
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

	user, err := repo.Get(ctx, insert.ID)
	assert.NoError(t, err)

	assert.Equal(t, insert.ID, user.ID)
	assert.Equal(t, insert.FirstName, user.FirstName)
	assert.Equal(t, insert.LastName, user.LastName)
	assert.Equal(t, insert.PhoneNumber, user.PhoneNumber)
	assert.Equal(t, insert.Email, user.Email)
	assert.Equal(t, insert.Role, user.Role)
	assert.Equal(t, insert.Verified, user.Verified)
	assert.Equal(t, insert.CreatedAt, user.CreatedAt)
	assert.Equal(t, insert.Picture, user.Picture)
	assert.Empty(t, user.Hash)
}

func TestCreate_Success(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	snowflakeNode := getSnowflakeNode()
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, snowflakeNode)

	user, err := repo.Create(ctx, userCreateArgs)
	assert.NoError(t, err)

	assert.Equal(t, userCreateArgs.FirstName, user.FirstName)
	assert.Equal(t, userCreateArgs.LastName, user.LastName)
	assert.Equal(t, userCreateArgs.PhoneNumber, user.PhoneNumber)
	assert.Equal(t, userCreateArgs.Email, user.Email)
	assert.Equal(t, userCreateArgs.Role, user.Role)
	assert.False(t, user.Verified.Bool)
	assert.True(t, user.Verified.Valid)

	assert.NotEqual(t, 0, user.ID)
	assert.NotEmpty(t, user.CreatedAt)

	assert.False(t, user.Picture.Valid)
	assert.Empty(t, user.Hash)
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
		{FirstName: userCreateArgs.FirstName, LastName: userCreateArgs.LastName, PhoneNumber: userCreateArgs.PhoneNumber, Hash: userCreateArgs.Hash, Role: invalidValue},
	}

	i := 0
	for _, args := range invalidArgs {
		user, err := repo.Create(ctx, args)
		if !assert.Error(t, err) {
			log.Println("create user:", i)
		}
		assert.Empty(t, user)
		i++
	}
}

func TestGetAuthInfo_Success(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	authInfo, err := repo.GetAuthInfoByEmail(ctx, insert.Email)
	assert.NoError(t, err)

	assert.Equal(t, insert.ID, authInfo.ID)
	assert.Equal(t, insert.Role, authInfo.Role)
	assert.Equal(t, insert.Hash, authInfo.Hash)
	assert.Equal(t, insert.Verified, authInfo.Verified)
}

func TestGetAuthInfoByEmail_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	authInfo, err := repo.GetAuthInfoByEmail(ctx, "email")
	assert.ErrorIs(t, err, pgx.ErrNoRows)
	assert.Empty(t, authInfo)
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

	updateArgs := repository.UpdateUserRow{
		FirstName:   firstName,
		LastName:    lastName,
		PhoneNumber: phoneNumber,
		Email:       email,
	}

	update, err := repo.Update(ctx, insert.ID, updateArgs)
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

	updateArgs := repository.UpdateUserRow{
		FirstName: firstName,
		LastName:  lastName,
	}

	update, err := repo.Update(ctx, insert.ID, updateArgs)
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

	update, err := repo.Update(ctx, 1, repository.UpdateUserRow{})
	assert.ErrorIs(t, err, repository.ErrInvalidUpdateParams)
	assert.Empty(t, update)
}

func TestUpdate_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	update, err := repo.Update(ctx, 1, repository.UpdateUserRow{FirstName: userCreateArgs.FirstName})
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

	ok, err := repo.UpdatePicture(ctx, insert.ID, "picture.jpg")
	assert.NoError(t, err)
	assert.True(t, ok)

	row := dbPool.QueryRow(ctx, "SELECT picture from users where id = $1", insert.ID)
	var updatedPicture pgtype.Text
	err = row.Scan(&updatedPicture)
	assert.NoError(t, err)

	assert.Equal(t, updatedPicture.String, "picture.jpg")
}

func TestUpdatePicture_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	ok, err := repo.UpdatePicture(ctx, 1, "")
	assert.NoError(t, err)
	assert.False(t, ok)
}

func TestUpdateVerification_Success(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	insert, err := insertUser(dbPool, ctx, userCreateArgs, 1)
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	ok, err := repo.UpdateVerification(ctx, insert.ID, true)
	assert.NoError(t, err)
	assert.True(t, ok)

	row := dbPool.QueryRow(ctx, "SELECT verified from users where id = $1", insert.ID)
	var updatedStatus pgtype.Bool
	err = row.Scan(&updatedStatus)
	assert.NoError(t, err)

	assert.True(t, updatedStatus.Bool)
}

func TestUpdateVerification_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	ok, err := repo.UpdateVerification(ctx, 1, true)
	assert.NoError(t, err)
	assert.False(t, ok)
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

	ok, err := repo.UpdateHash(ctx, insert.ID, hashArg)
	assert.NoError(t, err)
	assert.True(t, ok)

	row := dbPool.QueryRow(ctx, "SELECT hash from users where id = $1", insert.ID)
	var updatedHash string
	err = row.Scan(&updatedHash)
	assert.NoError(t, err)

	assert.Equal(t, hashArg, updatedHash)
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

	ok, err := repo.UpdateHash(ctx, insert.ID, "abc")
	var pgErr *pgconn.PgError
	if assert.ErrorAs(t, err, &pgErr) {
		assert.Equal(t, pgerrcode.CheckViolation, pgErr.Code)
	}
	assert.False(t, ok)

}

func TestUpdateHash_NotFound(t *testing.T) {
	ctx := context.Background()
	dbPool := getPgPool(ctx)
	defer cleanupPostgres(ctx, dbPool)

	repo := repository.NewUserRepository(dbPool, nil)

	ok, err := repo.UpdateHash(ctx, 1, "abc")
	assert.NoError(t, err)
	assert.False(t, ok)
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

	ok, err := repo.Delete(ctx, insert.ID)
	assert.NoError(t, err)
	assert.True(t, ok)

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

	ok, err := repo.Delete(ctx, 1)
	assert.NoError(t, err)
	assert.False(t, ok)
}

func insertUser(dbPool *pgxpool.Pool, ctx context.Context, args repository.CreateUserParams, id int64) (repository.User, error) {
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
	var i repository.User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.PhoneNumber,
		&i.Email,
		&i.Picture,
		&i.Hash,
		&i.Role,
		&i.Verified,
		&i.CreatedAt,
	)
	return i, err
}
