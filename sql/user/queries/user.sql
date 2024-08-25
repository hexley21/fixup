-- name: GetById :one
SELECT * FROM USERS WHERE id = $1;