-- name: CreateSupplier :one
INSERT INTO suppliers (
    code,
    name,
    contact_person,
    email,
    phone,
    address,
    tax_id,
    payment_terms,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id, code, name, created_at;

-- name: GetSupplierByID :one
SELECT 
    id,
    code,
    name,
    contact_person,
    email,
    phone,
    address,
    tax_id,
    payment_terms,
    is_active,
    created_at,
    updated_at
FROM suppliers
WHERE id = $1 
    AND deleted_at IS NULL
LIMIT 1;

-- name: ListActiveSuppliers :many
SELECT 
    id,
    code,
    name,
    contact_person,
    email,
    phone,
    payment_terms,
    is_active
FROM suppliers
WHERE deleted_at IS NULL
    AND is_active = true
ORDER BY name;

-- name: UpdateSupplier :exec
UPDATE suppliers
SET name = $2,
    contact_person = $3,
    email = $4,
    phone = $5,
    address = $6,
    tax_id = $7,
    payment_terms = $8,
    is_active = $9,
    updated_at = NOW()
WHERE id = $1 
    AND deleted_at IS NULL;

-- name: CreatePurchaseOrder :one
INSERT INTO purchase_orders (
    po_number,
    supplier_id,
    order_date,
    expected_date,
    status,
    total_amount,
    notes,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id, po_number, supplier_id, order_date, status, created_at;

-- name: CreatePurchaseOrderLine :one
INSERT INTO purchase_order_lines (
    po_id,
    product_id,
    quantity,
    unit_price,
    line_number,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id, po_id, product_id, quantity, unit_price, subtotal, line_number;

-- name: GetPurchaseOrderByID :one
SELECT 
    po.id,
    po.po_number,
    po.supplier_id,
    s.name as supplier_name,
    po.order_date,
    po.expected_date,
    po.status,
    po.total_amount,
    po.notes,
    po.created_at,
    po.updated_at
FROM purchase_orders po
INNER JOIN suppliers s ON po.supplier_id = s.id
WHERE po.id = $1
LIMIT 1;

-- name: GetPurchaseOrderLines :many
SELECT 
    pol.id,
    pol.po_id,
    pol.product_id,
    p.code as product_code,
    p.name as product_name,
    pol.quantity,
    pol.unit_price,
    pol.subtotal,
    pol.received_qty,
    pol.line_number,
    pol.notes
FROM purchase_order_lines pol
INNER JOIN products p ON pol.product_id = p.id
WHERE pol.po_id = $1
ORDER BY pol.line_number;

-- name: UpdatePurchaseOrderStatus :exec
UPDATE purchase_orders
SET status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: ListPurchaseOrders :many
SELECT 
    po.id,
    po.po_number,
    po.supplier_id,
    s.name as supplier_name,
    po.order_date,
    po.expected_date,
    po.status,
    po.total_amount,
    po.created_at
FROM purchase_orders po
INNER JOIN suppliers s ON po.supplier_id = s.id
WHERE ($1::uuid IS NULL OR po.supplier_id = $1)
    AND ($2::varchar IS NULL OR po.status = $2)
    AND ($3::date IS NULL OR po.order_date >= $3)
    AND ($4::date IS NULL OR po.order_date <= $4)
ORDER BY po.order_date DESC, po.po_number DESC
LIMIT $5 OFFSET $6;

-- name: CreateGoodsReceipt :one
INSERT INTO goods_receipts (
    receipt_number,
    po_id,
    receipt_date,
    notes,
    received_by
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, receipt_number, po_id, receipt_date, created_at;

-- name: CreateGoodsReceiptLine :one
INSERT INTO goods_receipt_lines (
    receipt_id,
    po_line_id,
    product_id,
    quantity,
    line_number
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING id, receipt_id, po_line_id, product_id, quantity, line_number;

-- name: UpdatePOLineReceivedQty :exec
UPDATE purchase_order_lines
SET received_qty = received_qty + $2
WHERE id = $1;

-- name: GetGoodsReceiptsByPO :many
SELECT 
    gr.id,
    gr.receipt_number,
    gr.po_id,
    gr.receipt_date,
    gr.notes,
    gr.received_by,
    u.full_name as received_by_name,
    gr.created_at
FROM goods_receipts gr
LEFT JOIN users u ON gr.received_by = u.id
WHERE gr.po_id = $1
ORDER BY gr.receipt_date DESC;

-- name: CreateSupplierBill :one
INSERT INTO supplier_bills (
    bill_number,
    supplier_id,
    po_id,
    bill_date,
    due_date,
    total_amount,
    status,
    notes,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id, bill_number, supplier_id, bill_date, due_date, total_amount, status, created_at;

-- name: GetSupplierBillByID :one
SELECT 
    sb.id,
    sb.bill_number,
    sb.supplier_id,
    s.name as supplier_name,
    sb.po_id,
    sb.bill_date,
    sb.due_date,
    sb.total_amount,
    sb.paid_amount,
    sb.status,
    sb.notes,
    sb.created_at
FROM supplier_bills sb
INNER JOIN suppliers s ON sb.supplier_id = s.id
WHERE sb.id = $1
LIMIT 1;

-- name: ListSupplierBills :many
SELECT 
    sb.id,
    sb.bill_number,
    sb.supplier_id,
    s.name as supplier_name,
    sb.bill_date,
    sb.due_date,
    sb.total_amount,
    sb.paid_amount,
    (sb.total_amount - sb.paid_amount) as outstanding,
    sb.status,
    sb.created_at
FROM supplier_bills sb
INNER JOIN suppliers s ON sb.supplier_id = s.id
WHERE ($1::uuid IS NULL OR sb.supplier_id = $1)
    AND ($2::varchar IS NULL OR sb.status = $2)
ORDER BY sb.due_date DESC, sb.bill_number DESC
LIMIT $3 OFFSET $4;

-- name: UpdateSupplierBillPaidAmount :exec
UPDATE supplier_bills
SET paid_amount = $2,
    status = CASE
        WHEN $2 >= total_amount THEN 'paid'
        WHEN $2 > 0 THEN 'partial'
        ELSE 'unpaid'
    END,
    updated_at = NOW()
WHERE id = $1;