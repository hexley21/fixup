--- name: CreateCategoryType :one
INSERT INTO category_types (name) VALUES ($1) RETURNING *;

-- name: GetCategoryType :one
SELECT * FROM category_types WHERE id = $1;

-- name: ListCategoryTypes :many
SELECT * FROM category_types ORDER BY id DESC OFFSET $1 LIMIT $2;

-- name: UpdateCategoryType :exec
UPDATE category_types SET name = $2 WHERE id = $1 Returning *;

-- name: DeleteCategoryType :exec
DELETE FROM category_types WHERE id = $1;
