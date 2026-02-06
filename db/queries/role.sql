-- name: CreateRole :one
INSERT INTO roles (
    code,
    name,
    description
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetRoleByID :one
SELECT * FROM roles
WHERE id = $1 LIMIT 1;

-- name: GetRoleByCode :one
SELECT * FROM roles
WHERE code = $1 LIMIT 1;

-- name: UpdateRole :one
UPDATE roles
SET 
    name = $2,
    description = $3,
    is_active = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1;

-- name: ListRoles :many
SELECT * FROM roles
ORDER BY created_at DESC;

-- name: UpdateRoleStatus :exec
UPDATE roles
SET is_active = $2, updated_at = NOW()
WHERE id = $1;