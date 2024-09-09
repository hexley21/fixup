-- name: GetById :one
SELECT * FROM users WHERE id = $1;

-- name: GetCredentialsByEmail :one
SELECT id, role, hash FROM users WHERE email = $1;

-- name: GetPasswordHashById :one
SELECT hash FROM users WHERE id = $1;

-- name: Create :one
INSERT INTO users (
  id, first_name, last_name, phone_number, email, hash, role
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: Update :one
UPDATE users
SET 
    first_name = $2,
    last_name = $3,
    phone_number = $4,
    email = $5
WHERE 
    id = $1
Returning *;

-- name: UpdateStatus :exec
UPDATE users SET user_status = $2 WHERE id = $1;

-- name: UpdatePicture :exec
UPDATE users SET picture_name = $2 WHERE id = $1;

-- name: UpdatePassword :exec
UPDATE users SET hash = $2 where id = $1;

-- name: Delete :exec
DELETE FROM users
WHERE id = $1;
