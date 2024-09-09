package repository

import (
	"context"
	"strconv"
	"strings"

	"github.com/bwmarrin/snowflake"
	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	postgres.Repository[UserRepository]
	Create(ctx context.Context, arg CreateParams) (entity.User, error)
	GetById(ctx context.Context, id int64) (entity.User, error)
	GetCredentialsByEmail(ctx context.Context, email string) (GetCredentialsByEmailRow, error)
	GetPasswordHashById(ctx context.Context, id int64) (string, error)
	Update(ctx context.Context, arg UpdateParams) (entity.User, error)
	UpdatePicture(ctx context.Context, arg UpdatePictureParams) error
	UpdateStatus(ctx context.Context, arg UpdateStatusParams) error
	UpdatePassword(ctx context.Context, arg UpdatePasswordParams) error
	DeleteById(ctx context.Context, id int64) error
}

type userRepositoryImpl struct {
	db        postgres.DBTX
	snowflake *snowflake.Node
}

func NewUserRepository(dbtx postgres.DBTX, snowflake *snowflake.Node) UserRepository {
	return &userRepositoryImpl{
		dbtx,
		snowflake,
	}
}

func (r userRepositoryImpl) WithTx(tx postgres.DBTX) UserRepository {
	return NewUserRepository(tx, r.snowflake)
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (
  id, first_name, last_name, phone_number, email, hash, role
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, first_name, last_name, phone_number, email, role, user_status, created_at
`

type CreateParams struct {
	FirstName   string
	LastName    string
	PhoneNumber string
	Email       string
	Hash        string
	PictureName string
	Role        enum.UserRole
}

func (r *userRepositoryImpl) Create(ctx context.Context, arg CreateParams) (entity.User, error) {
	row := r.db.QueryRow(ctx, createUser,
		r.snowflake.Generate(),
		arg.FirstName,
		arg.LastName,
		arg.PhoneNumber,
		arg.Email,
		arg.Hash,
		arg.Role,
	)
	var i entity.User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.PhoneNumber,
		&i.Email,
		&i.Role,
		&i.UserStatus,
		&i.CreatedAt,
	)
	return i, err
}

const getById = `-- name: GetById :one
SELECT id, first_name, last_name, phone_number, email, picture_name, role, user_status, created_at FROM users WHERE id = $1
`

func (r *userRepositoryImpl) GetById(ctx context.Context, id int64) (entity.User, error) {
	row := r.db.QueryRow(ctx, getById, id)
	var i entity.User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.PhoneNumber,
		&i.Email,
		&i.PictureName,
		&i.Role,
		&i.UserStatus,
		&i.CreatedAt,
	)
	return i, err
}

const getUserCredentialsByEmail = `-- name: GetUserCredentialsByEmail :one
SELECT id, role, hash FROM users WHERE email = $1
`

type GetCredentialsByEmailRow struct {
	ID   int64
	Role enum.UserRole
	Hash string
}

func (r *userRepositoryImpl) GetCredentialsByEmail(ctx context.Context, email string) (GetCredentialsByEmailRow, error) {
	row := r.db.QueryRow(ctx, getUserCredentialsByEmail, email)
	var i GetCredentialsByEmailRow
	err := row.Scan(&i.ID, &i.Role, &i.Hash)
	return i, err
}

const getPasswordHashById = `-- name: GetPasswordHashById :one
SELECT hash FROM users WHERE id = $1
`

func (r *userRepositoryImpl) GetPasswordHashById(ctx context.Context, id int64) (string, error) {
	row := r.db.QueryRow(ctx, getPasswordHashById, id)
	var hash string
	err := row.Scan(&hash)
	return hash, err
}

const baseUpdateUserData = `
UPDATE users
SET 
`

type UpdateParams struct {
	ID          int64
	FirstName   *string
	LastName    *string
	PhoneNumber *string
	Email       *string
}

func (r *userRepositoryImpl) Update(ctx context.Context, arg UpdateParams) (entity.User, error) {
	query := baseUpdateUserData
	params := []interface{}{arg.ID}
	setClauses := []string{}

	if arg.FirstName != nil {
		setClauses = append(setClauses, "first_name = $"+strconv.Itoa(len(params)+1))
		params = append(params, *arg.FirstName)
	}
	if arg.LastName != nil {
		setClauses = append(setClauses, "last_name = $"+strconv.Itoa(len(params)+1))
		params = append(params, *arg.LastName)
	}
	if arg.PhoneNumber != nil {
		setClauses = append(setClauses, "phone_number = $"+strconv.Itoa(len(params)+1))
		params = append(params, *arg.PhoneNumber)
	}
	if arg.Email != nil {
		setClauses = append(setClauses, "email = $"+strconv.Itoa(len(params)+1))
		params = append(params, *arg.Email)
	}

	query += strings.Join(setClauses, ", ")
	query += " WHERE id = $1 RETURNING id, first_name, last_name, phone_number, email, picture_name, role, user_status, created_at"

	row := r.db.QueryRow(ctx, query, params...)
	var i entity.User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.PhoneNumber,
		&i.Email,
		&i.PictureName,
		&i.Role,
		&i.UserStatus,
		&i.CreatedAt,
	)
	return i, err
}

const updateUserPicture = `-- name: UpdateUserPicture :exec
UPDATE users SET picture_name = $2 WHERE id = $1
`

type UpdatePictureParams struct {
	ID          int64
	PictureName pgtype.Text
}

func (r *userRepositoryImpl) UpdatePicture(ctx context.Context, arg UpdatePictureParams) error {
	result, err := r.db.Exec(ctx, updateUserPicture, arg.ID, arg.PictureName)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

const updateUserStatus = `-- name: UpdateUserStatus :exec
UPDATE users SET user_status = $2 WHERE id = $1
`

type UpdateStatusParams struct {
	ID         int64
	UserStatus pgtype.Bool
}

func (r *userRepositoryImpl) UpdateStatus(ctx context.Context, arg UpdateStatusParams) error {
	result, err := r.db.Exec(ctx, updateUserStatus, arg.ID, arg.UserStatus)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

const updateUserPassword = `-- name: UpdateUserPassword :exec
UPDATE users SET hash = $2 where id = $1
`

type UpdatePasswordParams struct {
	ID   int64
	Hash string
}

func (r *userRepositoryImpl) UpdatePassword(ctx context.Context, arg UpdatePasswordParams) error {
	result, err := r.db.Exec(ctx, updateUserPassword, arg.ID, arg.Hash)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	
	return nil
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1
`

func (r *userRepositoryImpl) DeleteById(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, deleteUser, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
