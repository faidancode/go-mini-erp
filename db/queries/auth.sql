-- name: GetUserByUsername :one
SELECT 
    id,
    username,
    email,
    password_hash,
    full_name,
    is_active,
    last_login_at,
    created_at,
    updated_at
FROM users
WHERE username = $1 
    AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByID :one
SELECT 
    id,
    username,
    email,
    full_name,
    is_active,
    last_login_at,
    created_at,
    updated_at
FROM users
WHERE id = $1 
    AND deleted_at IS NULL
LIMIT 1;

-- name: GetUserByEmail :one
SELECT 
    id,
    username,
    email,
    password_hash,
    full_name,
    is_active,
    last_login_at,
    created_at,
    updated_at
FROM users
WHERE email = $1 
    AND deleted_at IS NULL
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    username,
    email,
    password_hash,
    full_name,
    is_active
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, username, email, full_name, is_active, created_at;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: GetUserRoles :many
SELECT 
    r.id,
    r.code,
    r.name,
    r.description,
    ur.assigned_at
FROM roles r
INNER JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1 
    AND r.is_active = true
ORDER BY r.name;

-- name: GetUserMenus :many
SELECT 
    m.id,
    m.parent_id,
    m.code,
    m.name,
    m.path,
    m.icon,
    m.sort_order,
    -- Cast hasil MAX ke boolean agar SQLC men-generate tipe bool
    (MAX(rm.can_create::int) > 0)::BOOLEAN as can_create,
    (MAX(rm.can_read::int) > 0)::BOOLEAN as can_read,
    (MAX(rm.can_update::int) > 0)::BOOLEAN as can_update,
    (MAX(rm.can_delete::int) > 0)::BOOLEAN as can_delete
FROM menus m
INNER JOIN role_menus rm ON m.id = rm.menu_id
INNER JOIN user_roles ur ON rm.role_id = ur.role_id
WHERE ur.user_id = $1 
    AND m.is_active = true
GROUP BY m.id, m.parent_id, m.code, m.name, m.path, m.icon, m.sort_order
ORDER BY m.sort_order, m.name;

-- name: AssignRoleToUser :one
INSERT INTO user_roles (
    user_id,
    role_id,
    assigned_by
) VALUES (
    $1, $2, $3
)
RETURNING id, user_id, role_id, assigned_at;

-- name: CheckUsernameExists :one
SELECT EXISTS(
    SELECT 1 FROM users 
    WHERE username = $1 
    AND deleted_at IS NULL
) as exists;

-- name: CheckEmailExists :one
SELECT EXISTS(
    SELECT 1 FROM users 
    WHERE email = $1 
    AND deleted_at IS NULL
) as exists;