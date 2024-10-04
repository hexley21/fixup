-- name: CreateProvider :exec
INSERT INTO providers (
  personal_id_number, personal_id_preview, user_id
) VALUES (
  $1, $2, $3
);

-- name: GetByUserId :one
SELECT 
  personal_id_number, 
  personal_id_preview, 
  user_id 
FROM 
  providers
WHERE 
  user_id = $1;