package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"
	"github.com/hexley21/fixup/internal/user/entity"
	"github.com/hexley21/fixup/internal/user/enum"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	postgres.Repository[UserRepository]
	CreateUser(ctx context.Context, arg CreateUserParams) (entity.User, error)
	DeleteUser(ctx context.Context, id int64) error
	GetById(ctx context.Context, id int64) (entity.User, error)
	GetUserCredentialsByEmail(ctx context.Context, email string) (GetUserCredentialsByEmailRow, error)
	UpdateUserPicture(ctx context.Context, arg UpdateUserPictureParams) error
	UpdateUserStatus(ctx context.Context, arg UpdateUserStatusParams) error
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

type CreateUserParams struct {
	FirstName   string
	LastName    string
	PhoneNumber string
	Email       string
	Hash        string
	PictureName string
	Role        enum.UserRole
}

func (r *userRepositoryImpl) CreateUser(ctx context.Context, arg CreateUserParams) (entity.User, error) {
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

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1
`

func (r *userRepositoryImpl) DeleteUser(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, deleteUser, id)
	return err
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

type GetUserCredentialsByEmailRow struct {
	ID   int64
	Role enum.UserRole
	Hash string
}

func (r *userRepositoryImpl) GetUserCredentialsByEmail(ctx context.Context, email string) (GetUserCredentialsByEmailRow, error) {
	row := r.db.QueryRow(ctx, getUserCredentialsByEmail, email)
	var i GetUserCredentialsByEmailRow
	err := row.Scan(&i.ID, &i.Role, &i.Hash)
	return i, err
}

const updateUserPicture = `-- name: UpdateUserPicture :exec
UPDATE users SET picture_name = $2 WHERE id = $1
`

type UpdateUserPictureParams struct {
	ID          int64
	PictureName pgtype.Text
}

func (r *userRepositoryImpl) UpdateUserPicture(ctx context.Context, arg UpdateUserPictureParams) error {
	_, err := r.db.Exec(ctx, updateUserPicture, arg.ID, arg.PictureName)
	return err
}

const updateUserStatus = `-- name: UpdateUserStatus :exec
UPDATE users SET user_status = $2 WHERE id = $1
`

type UpdateUserStatusParams struct {
	ID         int64       `json:"id"`
	UserStatus pgtype.Bool `json:"user_status"`
}

func (r *userRepositoryImpl) UpdateUserStatus(ctx context.Context, arg UpdateUserStatusParams) error {
	_, err := r.db.Exec(ctx, updateUserStatus, arg.ID, arg.UserStatus)
	return err
}
