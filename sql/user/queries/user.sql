-- name: GetById :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserCredentialsByEmail :one
SELECT id, role, hash FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (
  id, first_name, last_name, phone_number, email, hash, role
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateUserStatus :exec
UPDATE users SET user_status = $2 WHERE id = $1;

-- name: UpdateUserPicture :exec
UPDATE users SET picture_name = $2 WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
