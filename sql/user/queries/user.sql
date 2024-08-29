-- name: GetById :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserPasswordHash :one
SELECT hash FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (
  id, first_name, last_name, phone_number, email, hash, role
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING id, first_name, last_name, phone_number, email, role, user_status, created_at;

-- name: UpdateUserStatus :exec
UPDATE users
  set user_status = $2
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
