-- name: CreateCategory :one
INSERT INTO categories (
    code,
    name,
    description,
    parent_id,
    is_active
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, code, name, created_at;

-- name: GetCategoryByID :one
SELECT 
    id,
    code,
    name,
    description,
    parent_id,
    is_active,
    created_at,
    updated_at
FROM categories
WHERE id = $1
LIMIT 1;

-- name: ListActiveCategories :many
SELECT 
    id,
    code,
    name,
    description,
    parent_id,
    is_active
FROM categories
WHERE is_active = true
ORDER BY name;

-- name: CreateUnitOfMeasure :one
INSERT INTO units_of_measure (
    code,
    name,
    unit_type,
    is_base_unit,
    conversion_factor,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id, code, name, unit_type;

-- name: ListActiveUoM :many
SELECT 
    id,
    code,
    name,
    unit_type,
    is_base_unit,
    conversion_factor
FROM units_of_measure
WHERE is_active = true
ORDER BY unit_type, name;

-- name: CreateProduct :one
INSERT INTO products (
    code,
    name,
    description,
    category_id,
    uom_id,
    product_type,
    cost_price,
    sale_price,
    min_stock,
    max_stock,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING id, code, name, created_at;

-- name: GetProductByID :one
SELECT 
    p.id,
    p.code,
    p.name,
    p.description,
    p.category_id,
    c.name as category_name,
    p.uom_id,
    u.code as uom_code,
    u.name as uom_name,
    p.product_type,
    p.cost_price,
    p.sale_price,
    p.min_stock,
    p.max_stock,
    p.is_active,
    p.created_at,
    p.updated_at
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
INNER JOIN units_of_measure u ON p.uom_id = u.id
WHERE p.id = $1 
    AND p.deleted_at IS NULL
LIMIT 1;

-- name: GetProductByCode :one
SELECT 
    p.id,
    p.code,
    p.name,
    p.description,
    p.category_id,
    p.uom_id,
    p.product_type,
    p.cost_price,
    p.sale_price,
    p.is_active
FROM products p
WHERE p.code = $1 
    AND p.deleted_at IS NULL
LIMIT 1;

-- name: ListActiveProducts :many
SELECT 
    p.id,
    p.code,
    p.name,
    p.category_id,
    c.name as category_name,
    p.product_type,
    p.cost_price,
    p.sale_price,
    p.is_active
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
WHERE p.deleted_at IS NULL
    AND p.is_active = true
    AND ($1::uuid IS NULL OR p.category_id = $1)
ORDER BY p.name;

-- name: UpdateProduct :exec
UPDATE products
SET name = $2,
    description = $3,
    category_id = $4,
    cost_price = $5,
    sale_price = $6,
    min_stock = $7,
    max_stock = $8,
    is_active = $9,
    updated_at = NOW()
WHERE id = $1 
    AND deleted_at IS NULL;

-- name: CreateStockLocation :one
INSERT INTO stock_locations (
    code,
    name,
    location_type,
    is_active
) VALUES (
    $1, $2, $3, $4
)
RETURNING id, code, name, location_type;

-- name: ListActiveStockLocations :many
SELECT 
    id,
    code,
    name,
    location_type,
    is_active
FROM stock_locations
WHERE is_active = true
ORDER BY name;

-- name: GetStockBalance :one
SELECT 
    sb.id,
    sb.product_id,
    p.code as product_code,
    p.name as product_name,
    sb.location_id,
    sl.name as location_name,
    sb.quantity,
    sb.reserved_qty,
    sb.available_qty,
    sb.last_updated
FROM stock_balances sb
INNER JOIN products p ON sb.product_id = p.id
INNER JOIN stock_locations sl ON sb.location_id = sl.id
WHERE sb.product_id = $1 
    AND sb.location_id = $2
LIMIT 1;

-- name: UpsertStockBalance :one
INSERT INTO stock_balances (
    product_id,
    location_id,
    quantity,
    reserved_qty,
    last_updated
) VALUES (
    $1, $2, $3, $4, NOW()
)
ON CONFLICT (product_id, location_id)
DO UPDATE SET
    quantity = stock_balances.quantity + EXCLUDED.quantity,
    last_updated = NOW()
RETURNING id, product_id, location_id, quantity, reserved_qty, available_qty;

-- name: UpdateStockQuantity :exec
UPDATE stock_balances
SET quantity = quantity + $3,
    last_updated = NOW()
WHERE product_id = $1 
    AND location_id = $2;

-- name: ListStockBalances :many
SELECT 
    sb.id,
    sb.product_id,
    p.code as product_code,
    p.name as product_name,
    sb.location_id,
    sl.name as location_name,
    sb.quantity,
    sb.reserved_qty,
    sb.available_qty,
    p.min_stock,
    p.max_stock,
    sb.last_updated
FROM stock_balances sb
INNER JOIN products p ON sb.product_id = p.id
INNER JOIN stock_locations sl ON sb.location_id = sl.id
WHERE ($1::uuid IS NULL OR sb.product_id = $1)
    AND ($2::uuid IS NULL OR sb.location_id = $2)
    AND ($3::boolean IS NULL OR (sb.quantity > 0) = $3)
ORDER BY p.name, sl.name;

-- name: CreateStockMovement :one
INSERT INTO stock_movements (
    movement_number,
    product_id,
    location_id,
    movement_type,
    quantity,
    reference_type,
    reference_id,
    movement_date,
    notes,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING id, movement_number, product_id, location_id, movement_type, quantity, movement_date;

-- name: ListStockMovements :many
SELECT 
    sm.id,
    sm.movement_number,
    sm.product_id,
    p.code as product_code,
    p.name as product_name,
    sm.location_id,
    sl.name as location_name,
    sm.movement_type,
    sm.quantity,
    sm.reference_type,
    sm.reference_id,
    sm.movement_date,
    sm.notes,
    sm.created_at
FROM stock_movements sm
INNER JOIN products p ON sm.product_id = p.id
INNER JOIN stock_locations sl ON sm.location_id = sl.id
WHERE ($1::uuid IS NULL OR sm.product_id = $1)
    AND ($2::uuid IS NULL OR sm.location_id = $2)
    AND ($3::varchar IS NULL OR sm.movement_type = $3)
    AND ($4::timestamptz IS NULL OR sm.movement_date >= $4)
    AND ($5::timestamptz IS NULL OR sm.movement_date <= $5)
ORDER BY sm.movement_date DESC, sm.created_at DESC
LIMIT $6 OFFSET $7;

-- name: CreateStockAdjustment :one
INSERT INTO stock_adjustments (
    adjustment_number,
    location_id,
    adjustment_date,
    reason,
    status,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id, adjustment_number, location_id, adjustment_date, status, created_at;

-- name: CreateStockAdjustmentLine :one
INSERT INTO stock_adjustment_lines (
    adjustment_id,
    product_id,
    quantity_before,
    quantity_after,
    line_number,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id, adjustment_id, product_id, quantity_before, quantity_after, difference, line_number;

-- name: GetStockAdjustmentByID :one
SELECT 
    sa.id,
    sa.adjustment_number,
    sa.location_id,
    sl.name as location_name,
    sa.adjustment_date,
    sa.reason,
    sa.status,
    sa.created_by,
    u.full_name as created_by_name,
    sa.created_at,
    sa.updated_at
FROM stock_adjustments sa
INNER JOIN stock_locations sl ON sa.location_id = sl.id
LEFT JOIN users u ON sa.created_by = u.id
WHERE sa.id = $1
LIMIT 1;

-- name: GetStockAdjustmentLines :many
SELECT 
    sal.id,
    sal.adjustment_id,
    sal.product_id,
    p.code as product_code,
    p.name as product_name,
    sal.quantity_before,
    sal.quantity_after,
    sal.difference,
    sal.line_number,
    sal.notes
FROM stock_adjustment_lines sal
INNER JOIN products p ON sal.product_id = p.id
WHERE sal.adjustment_id = $1
ORDER BY sal.line_number;

-- name: UpdateStockAdjustmentStatus :exec
UPDATE stock_adjustments
SET status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: ListStockAdjustments :many
SELECT 
    sa.id,
    sa.adjustment_number,
    sa.location_id,
    sl.name as location_name,
    sa.adjustment_date,
    sa.reason,
    sa.status,
    sa.created_at
FROM stock_adjustments sa
INNER JOIN stock_locations sl ON sa.location_id = sl.id
WHERE ($1::uuid IS NULL OR sa.location_id = $1)
    AND ($2::varchar IS NULL OR sa.status = $2)
ORDER BY sa.adjustment_date DESC, sa.adjustment_number DESC
LIMIT $3 OFFSET $4;