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

-- name: GetMenusByUserID :many
SELECT DISTINCT
    m.id,
    m.parent_id,
    m.code,
    m.name,
    m.path,
    m.icon,
    m.sort_order,
    rm.can_create,
    rm.can_read,
    rm.can_update,
    rm.can_delete
FROM menus m
INNER JOIN role_menus rm ON m.id = rm.menu_id
INNER JOIN user_roles ur ON rm.role_id = ur.role_id
WHERE ur.user_id = $1 
    AND m.is_active = true
ORDER BY m.sort_order, m.name;

-- name: CheckUserMenuAccess :one
SELECT 
    COALESCE(MAX(rm.can_read::int), 0) > 0 as can_read,
    COALESCE(MAX(rm.can_create::int), 0) > 0 as can_create,
    COALESCE(MAX(rm.can_update::int), 0) > 0 as can_update,
    COALESCE(MAX(rm.can_delete::int), 0) > 0 as can_delete
FROM role_menus rm
INNER JOIN user_roles ur ON rm.role_id = ur.role_id
INNER JOIN menus m ON rm.menu_id = m.id
WHERE ur.user_id = $1 
    AND m.code = $2
    AND m.is_active = true;

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

-- name: AssignRoleToUser :one
INSERT INTO user_roles (
    user_id,
    role_id,
    assigned_by
) VALUES (
    $1, $2, $3
)
RETURNING id, user_id, role_id, assigned_at;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: ListActiveUsers :many
SELECT 
    id,
    username,
    email,
    full_name,
    is_active,
    last_login_at,
    created_at
FROM users
WHERE deleted_at IS NULL
    AND ($1::boolean IS NULL OR is_active = $1)
ORDER BY full_name;

-- name: GetRoleByCode :one
SELECT 
    id,
    code,
    name,
    description,
    is_active,
    created_at
FROM roles
WHERE code = $1
    AND is_active = true
LIMIT 1;