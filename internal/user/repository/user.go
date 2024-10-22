package repository

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/bwmarrin/snowflake"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/jackc/pgx/v5/pgtype"
)

var(
	ErrInvalidUpdateParams = errors.New("invalid update params")
)

type UserRepository interface {
	postgres.Repository[UserRepository]
	Create(ctx context.Context, arg CreateUserParams) (User, error)
	Delete(ctx context.Context, id int64) (bool, error)
	GetHashById(ctx context.Context, id int64) (string, error)
	GetAccountInfo(ctx context.Context, id int64) (GetUserAccountInfoRow, error)
	GetVerificationInfo(ctx context.Context, email string) (GetUserVerificationInfoRow, error)
	GetPicture(ctx context.Context, id int64) (pgtype.Text, error)
	Get(ctx context.Context, id int64) (User, error)
	GetAuthInfoByEmail(ctx context.Context, email string) (GetUserAuthInfoByEmailRow, error)
	Update(ctx context.Context, id int64, arg UpdateUserRow) (UpdateUserRow, error)
	UpdateVerification(ctx context.Context, id int64, verified bool) (bool, error)
	UpdateHash(ctx context.Context, id int64, hash string) (bool, error)
	UpdatePicture(ctx context.Context, id int64, picture string) (bool, error)
}

type pgsqlUserRepository struct {
	db        postgres.PGXQuerier
	snowflake *snowflake.Node
}

func NewUserRepository(q postgres.PGXQuerier, snowflake *snowflake.Node) UserRepository {
	return &pgsqlUserRepository{
		q,
		snowflake,
	}
}

func (r *pgsqlUserRepository) WithTx(tx postgres.PGXQuerier) UserRepository {
	return NewUserRepository(tx, r.snowflake)
}

const createUser = `-- name: CreateUser :one
INSERT INTO users (
  id, first_name, last_name, phone_number, email, hash, role
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, first_name, last_name, phone_number, email, picture, role, verified, created_at
`

type CreateUserParams struct {
	ID          int64
	FirstName   string
	LastName    string
	PhoneNumber string
	Email       string
	Hash        string
	Role        string
}

func (r *pgsqlUserRepository) Create(ctx context.Context, arg CreateUserParams) (User, error) {
	row := r.db.QueryRow(ctx, createUser,
		r.snowflake.Generate(),
		arg.FirstName,
		arg.LastName,
		arg.PhoneNumber,
		arg.Email,
		arg.Hash,
		arg.Role,
	)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.PhoneNumber,
		&i.Email,
		&i.Picture,
		&i.Role,
		&i.Verified,
		&i.CreatedAt,
	)
	return i, err
}

const deleteUser = `-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1
`

func (r *pgsqlUserRepository) Delete(ctx context.Context, id int64) (bool, error) {
	result, err := r.db.Exec(ctx, deleteUser, id)
	return result.RowsAffected() > 0, err
}

const getHashById = `-- name: GetHashById :one
SELECT hash FROM users WHERE id = $1
`

func (r *pgsqlUserRepository) GetHashById(ctx context.Context, id int64) (string, error) {
	row := r.db.QueryRow(ctx, getHashById, id)
	var hash string
	err := row.Scan(&hash)
	return hash, err
}

const getUserAccountInfo = `-- name: GetUserAccountInfo :one
SELECT role, verifid FROM users WHERE id = $1
`

type GetUserAccountInfoRow struct {
	Role     string
	Verified pgtype.Bool
}

func (r *pgsqlUserRepository) GetAccountInfo(ctx context.Context, id int64) (GetUserAccountInfoRow, error) {
	row := r.db.QueryRow(ctx, getUserAccountInfo, id)
	var i GetUserAccountInfoRow
	err := row.Scan(&i.Role, &i.Verified)
	return i, err
}

const getUserActivationInfo = `-- name: GetUserActivationInfo :one
SELECT id, verified, first_name FROM users WHERE email = $1
`

type GetUserVerificationInfoRow struct {
	ID        int64
	Verified  pgtype.Bool
	FirstName string
}

func (r *pgsqlUserRepository) GetVerificationInfo(ctx context.Context, email string) (GetUserVerificationInfoRow, error) {
	row := r.db.QueryRow(ctx, getUserActivationInfo, email)
	var i GetUserVerificationInfoRow
	err := row.Scan(&i.ID, &i.Verified, &i.FirstName)
	return i, err
}

const getUserPicture = `-- name: GetUserPicture :one
SELECT picture FROM users WHERE id = $1
`

func (r *pgsqlUserRepository) GetPicture(ctx context.Context, id int64) (pgtype.Text, error) {
	row := r.db.QueryRow(ctx, getUserPicture, id)
	var picture pgtype.Text
	err := row.Scan(&picture)
	return picture, err
}

const getUser = `-- name: GetUser :one
SELECT id, first_name, last_name, phone_number, email, picture, role, verified, created_at FROM users WHERE id = $1
`

func (r *pgsqlUserRepository) Get(ctx context.Context, id int64) (User, error) {
	row := r.db.QueryRow(ctx, getUser, id)
	var i User
	err := row.Scan(
		&i.ID,
		&i.FirstName,
		&i.LastName,
		&i.PhoneNumber,
		&i.Email,
		&i.Picture,
		&i.Role,
		&i.Verified,
		&i.CreatedAt,
	)
	return i, err
}

const getUserAuthInfoByEmail = `-- name: GetUserAuthInfoByEmail :one
SELECT id, role, verified, hash FROM users WHERE email = $1
`

type GetUserAuthInfoByEmailRow struct {
	ID       int64
	Role     string
	Verified pgtype.Bool
	Hash     string
}

func (r *pgsqlUserRepository) GetAuthInfoByEmail(ctx context.Context, email string) (GetUserAuthInfoByEmailRow, error) {
	row := r.db.QueryRow(ctx, getUserAuthInfoByEmail, email)
	var i GetUserAuthInfoByEmailRow
	err := row.Scan(
		&i.ID,
		&i.Role,
		&i.Verified,
		&i.Hash,
	)
	return i, err
}

const baseUpdateUserData = `
UPDATE users
SET 
`

type UpdateUserRow struct {
	FirstName   string
	LastName    string
	PhoneNumber string
	Email       string
}

// Update updates a user's information by their ID, supporting partial updates.
// It constructs an SQL query based on the provided fields in UpdateUserRow and executes it.
// It returns the updated user information or an error if the update fails or if no fields are provided.
func (r *pgsqlUserRepository) Update(ctx context.Context, id int64, arg UpdateUserRow) (UpdateUserRow, error) {
	var i UpdateUserRow

	query := baseUpdateUserData
	params := []any{id}
	var setClauses []string

	if arg.FirstName != "" {
		setClauses = append(setClauses, "first_name = $"+strconv.Itoa(len(params)+1))
		params = append(params, arg.FirstName)
	}
	if arg.LastName != "" {
		setClauses = append(setClauses, "last_name = $"+strconv.Itoa(len(params)+1))
		params = append(params, arg.LastName)
	}
	if arg.PhoneNumber != "" {
		setClauses = append(setClauses, "phone_number = $"+strconv.Itoa(len(params)+1))
		params = append(params, arg.PhoneNumber)
	}
	if arg.Email != "" {
		setClauses = append(setClauses, "email = $"+strconv.Itoa(len(params)+1))
		params = append(params, arg.Email)
	}

	if len(params) == 1 {
		return i, ErrInvalidUpdateParams
	}

	query += strings.Join(setClauses, ", ")
	query += " WHERE id = $1 RETURNING first_name, last_name, phone_number, email"

	row := r.db.QueryRow(ctx, query, params...)

	err := row.Scan(
		&i.FirstName,
		&i.LastName,
		&i.PhoneNumber,
		&i.Email,
	)
	return i, err
}

const updateUserVerification = `-- name: UpdateUserVerification :exec
UPDATE users SET verified = $2 WHERE id = $1
`

func (r *pgsqlUserRepository) UpdateVerification(ctx context.Context, id int64, verifid bool) (bool, error) {
	result, err := r.db.Exec(ctx, updateUserVerification, id, pgtype.Bool{Bool: verifid, Valid: true})
	return result.RowsAffected() > 0, err
}

const updateUserHash = `-- name: UpdateUserHash :exec
UPDATE users SET hash = $2 where id = $1
`

func (r *pgsqlUserRepository) UpdateHash(ctx context.Context, id int64, hash string) (bool, error) {
	result, err := r.db.Exec(ctx, updateUserHash, id, hash)
	return result.RowsAffected() > 0, err
}

const updateUserPicture = `-- name: UpdateUserPicture :exec
UPDATE users SET picture = $2 WHERE id = $1
`

func (r *pgsqlUserRepository) UpdatePicture(ctx context.Context, id int64, picture string) (bool, error) {
	result, err := r.db.Exec(ctx, updateUserPicture, id, pgtype.Text{String: picture, Valid: true})
	return result.RowsAffected() > 0, err
}
