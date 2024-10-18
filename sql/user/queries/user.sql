-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserAuthInfoByEmail :one
SELECT id, role, verified, hash FROM users WHERE email = $1;

-- name: GetHashById :one
SELECT hash FROM users WHERE id = $1;

-- name: GetUserVerificationInfo :one
SELECT id, verified, first_name FROM users WHERE email = $1;

-- name: GetUserAccountInfo :one
SELECT role, verified FROM users WHERE id = $1;

-- name: GetUserPicture :one
SELECT picture FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (
  id, first_name, last_name, phone_number, email, hash, role
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateUser :one
UPDATE users SET first_name = $2, last_name = $3, phone_number = $4, email = $5 WHERE id = $1 Returning first_name, last_name, phone_number, email;

-- name: UpdateUserVerification :exec
UPDATE users SET verified = $2 WHERE id = $1;

-- name: UpdateUserPicture :exec
UPDATE users SET picture = $2 WHERE id = $1;

-- name: UpdateUserHash :exec
UPDATE users SET hash = $2 where id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
