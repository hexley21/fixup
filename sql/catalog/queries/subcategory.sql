-- name: CreateSubcategory :one
INSERT INTO subcategories (category_id, name) VALUES ($1, $2) RETURNING *;

-- name: GetSubcategoryById :one
SELECT * FROM subcategories WHERE id = $1;

-- name: GetSubategories :many
SELECT * FROM subcategories ORDER BY id LIMIT $1 OFFSET $2;

-- name: GetSubategoriesByCategoryId :many
SELECT * FROM subcategories WHERE category_id = $1 ORDER BY id LIMIT $2 OFFSET $3;

-- name: GetSubategoriesByTypeId :many
SELECT s.* 
FROM subcategories s
JOIN categories c ON s.category_id = c.id
WHERE c.type_id = $1
ORDER BY s.id LIMIT $2 OFFSET $3;

-- name: UpdateSubcategoryById :one
UPDATE subcategories SET name = $1, category_id = $2 WHERE id = $3 RETURNING *;

-- name: DeleteSubcategoryById :exec
DELETE FROM subcategories WHERE id = $1;
