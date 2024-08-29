package repository

import (
	"context"

	"github.com/bwmarrin/snowflake"
	"github.com/hexley21/handy/internal/user/entity"
	"github.com/hexley21/handy/internal/user/enum"
	"github.com/hexley21/handy/pkg/infra/postgres"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	postgres.Repository[UserRepository]
	CreateUser(ctx context.Context, arg CreateUserParams) (entity.User, error)
	DeleteUser(ctx context.Context, id int64) error
	GetById(ctx context.Context, id int64) (entity.User, error)
	GetUserPasswordHash(ctx context.Context, id int64) (string, error)
	UpdateUserStatus(ctx context.Context, arg UpdateUserStatusParams) error
}

type userRepositoryImpl struct {
	db postgres.DBTX
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
	FirstName   string        `json:"first_name"`
	LastName    string        `json:"last_name"`
	PhoneNumber string        `json:"phone_number"`
	Email       string        `json:"email"`
	Hash        string        `json:"hash"`
	Role        enum.UserRole `json:"role"`
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
SELECT id, first_name, last_name, phone_number, email, role, user_status, created_at FROM users WHERE id = $1
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
		&i.Role,
		&i.UserStatus,
		&i.CreatedAt,
	)
	return i, err
}

const getUserPasswordHash = `-- name: GetUserPasswordHash :one
SELECT hash FROM users WHERE id = $1
`

func (r *userRepositoryImpl) GetUserPasswordHash(ctx context.Context, id int64) (string, error) {
	row := r.db.QueryRow(ctx, getUserPasswordHash, id)
	var hash string
	err := row.Scan(&hash)
	return hash, err
}

const updateUserStatus = `-- name: UpdateUserStatus :exec
UPDATE users
  set user_status = $2
WHERE id = $1
`

type UpdateUserStatusParams struct {
	ID         int64       `json:"id"`
	UserStatus pgtype.Bool `json:"user_status"`
}

func (r *userRepositoryImpl) UpdateUserStatus(ctx context.Context, arg UpdateUserStatusParams) error {
	_, err := r.db.Exec(ctx, updateUserStatus, arg.ID, arg.UserStatus)
	return err
}