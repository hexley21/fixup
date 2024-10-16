-- name: CreateSubcategory :one
INSERT INTO subcategories (category_id, name) VALUES ($1, $2) RETURNING id;

-- name: GetSubcategory :one
SELECT * FROM subcategories WHERE id = $1;

-- name: ListSubategories :many
SELECT * FROM subcategories ORDER BY id OFFSET $1 LIMIT $2;

-- name: ListSubategoriesByCategoryId :many
SELECT * FROM subcategories WHERE category_id = $1 ORDER BY id OFFSET $2 LIMIT $3;

-- name: ListSubategoriesByTypeId :many
SELECT s.* 
FROM subcategories s
JOIN categories c ON s.category_id = c.id
WHERE c.type_id = $1
ORDER BY s.id OFFSET $2 LIMIT $3;

-- name: UpdateSubcategory :one
UPDATE subcategories SET name = $1, category_id = $2 WHERE id = $3 RETURNING *;

-- name: DeleteSubcategory :exec
DELETE FROM subcategories WHERE id = $1;
