-- name: CreateCategory :one
INSERT INTO categories (type_id, name) VALUES ($1, $2) RETURNING id;

-- name: GetCategory :one
SELECT * FROM categories WHERE id = $1;

-- name: ListCategoriesByTypeId :many
SELECT * FROM categories WHERE type_id = $1 ORDER BY id DESC OFFSET $2 LIMIT $3;

-- name: ListCategories :many
SELECT * FROM categories ORDER BY id DESC OFFSET $1 LIMIT $2;

-- name: UpdateCategory :one
UPDATE categories SET name = $2, type_id = $3 WHERE id = $1 Returning *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;
