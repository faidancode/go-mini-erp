-- name: CreateCustomer :one
INSERT INTO customers (
    code,
    name,
    contact_person,
    email,
    phone,
    address,
    tax_id,
    credit_limit,
    payment_terms,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING id, code, name, created_at;

-- name: GetCustomerByID :one
SELECT 
    id,
    code,
    name,
    contact_person,
    email,
    phone,
    address,
    tax_id,
    credit_limit,
    payment_terms,
    is_active,
    created_at,
    updated_at
FROM customers
WHERE id = $1 
    AND deleted_at IS NULL
LIMIT 1;

-- name: ListActiveCustomers :many
SELECT 
    id,
    code,
    name,
    contact_person,
    email,
    phone,
    credit_limit,
    payment_terms,
    is_active
FROM customers
WHERE deleted_at IS NULL
    AND is_active = true
ORDER BY name;

-- name: UpdateCustomer :exec
UPDATE customers
SET name = $2,
    contact_person = $3,
    email = $4,
    phone = $5,
    address = $6,
    tax_id = $7,
    credit_limit = $8,
    payment_terms = $9,
    is_active = $10,
    updated_at = NOW()
WHERE id = $1 
    AND deleted_at IS NULL;

-- name: CreateQuotation :one
INSERT INTO quotations (
    quote_number,
    customer_id,
    quote_date,
    valid_until,
    status,
    total_amount,
    notes,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id, quote_number, customer_id, quote_date, status, created_at;

-- name: CreateQuotationLine :one
INSERT INTO quotation_lines (
    quote_id,
    product_id,
    quantity,
    unit_price,
    line_number,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id, quote_id, product_id, quantity, unit_price, subtotal, line_number;

-- name: GetQuotationByID :one
SELECT 
    q.id,
    q.quote_number,
    q.customer_id,
    c.name as customer_name,
    q.quote_date,
    q.valid_until,
    q.status,
    q.total_amount,
    q.notes,
    q.created_at,
    q.updated_at
FROM quotations q
INNER JOIN customers c ON q.customer_id = c.id
WHERE q.id = $1
LIMIT 1;

-- name: GetQuotationLines :many
SELECT 
    ql.id,
    ql.quote_id,
    ql.product_id,
    p.code as product_code,
    p.name as product_name,
    ql.quantity,
    ql.unit_price,
    ql.subtotal,
    ql.line_number,
    ql.notes
FROM quotation_lines ql
INNER JOIN products p ON ql.product_id = p.id
WHERE ql.quote_id = $1
ORDER BY ql.line_number;

-- name: UpdateQuotationStatus :exec
UPDATE quotations
SET status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: ListQuotations :many
SELECT 
    q.id,
    q.quote_number,
    q.customer_id,
    c.name as customer_name,
    q.quote_date,
    q.valid_until,
    q.status,
    q.total_amount,
    q.created_at
FROM quotations q
INNER JOIN customers c ON q.customer_id = c.id
WHERE ($1::uuid IS NULL OR q.customer_id = $1)
    AND ($2::varchar IS NULL OR q.status = $2)
    AND ($3::date IS NULL OR q.quote_date >= $3)
    AND ($4::date IS NULL OR q.quote_date <= $4)
ORDER BY q.quote_date DESC, q.quote_number DESC
LIMIT $5 OFFSET $6;

-- name: CreateSalesOrder :one
INSERT INTO sales_orders (
    so_number,
    customer_id,
    quote_id,
    order_date,
    delivery_date,
    status,
    total_amount,
    notes,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id, so_number, customer_id, order_date, status, created_at;

-- name: CreateSalesOrderLine :one
INSERT INTO sales_order_lines (
    so_id,
    product_id,
    quantity,
    unit_price,
    line_number,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING id, so_id, product_id, quantity, unit_price, subtotal, line_number;

-- name: GetSalesOrderByID :one
SELECT 
    so.id,
    so.so_number,
    so.customer_id,
    c.name as customer_name,
    so.quote_id,
    so.order_date,
    so.delivery_date,
    so.status,
    so.total_amount,
    so.notes,
    so.created_at,
    so.updated_at
FROM sales_orders so
INNER JOIN customers c ON so.customer_id = c.id
WHERE so.id = $1
LIMIT 1;

-- name: GetSalesOrderLines :many
SELECT 
    sol.id,
    sol.so_id,
    sol.product_id,
    p.code as product_code,
    p.name as product_name,
    sol.quantity,
    sol.unit_price,
    sol.subtotal,
    sol.delivered_qty,
    sol.line_number,
    sol.notes
FROM sales_order_lines sol
INNER JOIN products p ON sol.product_id = p.id
WHERE sol.so_id = $1
ORDER BY sol.line_number;

-- name: UpdateSalesOrderStatus :exec
UPDATE sales_orders
SET status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateSOLineDeliveredQty :exec
UPDATE sales_order_lines
SET delivered_qty = delivered_qty + $2
WHERE id = $1;

-- name: ListSalesOrders :many
SELECT 
    so.id,
    so.so_number,
    so.customer_id,
    c.name as customer_name,
    so.order_date,
    so.delivery_date,
    so.status,
    so.total_amount,
    so.created_at
FROM sales_orders so
INNER JOIN customers c ON so.customer_id = c.id
WHERE ($1::uuid IS NULL OR so.customer_id = $1)
    AND ($2::varchar IS NULL OR so.status = $2)
    AND ($3::date IS NULL OR so.order_date >= $3)
    AND ($4::date IS NULL OR so.order_date <= $4)
ORDER BY so.order_date DESC, so.so_number DESC
LIMIT $5 OFFSET $6;

-- name: CreateCustomerInvoice :one
INSERT INTO customer_invoices (
    invoice_number,
    customer_id,
    so_id,
    invoice_date,
    due_date,
    total_amount,
    status,
    notes,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING id, invoice_number, customer_id, invoice_date, due_date, total_amount, status, created_at;

-- name: GetCustomerInvoiceByID :one
SELECT 
    ci.id,
    ci.invoice_number,
    ci.customer_id,
    c.name as customer_name,
    ci.so_id,
    ci.invoice_date,
    ci.due_date,
    ci.total_amount,
    ci.paid_amount,
    ci.status,
    ci.notes,
    ci.created_at
FROM customer_invoices ci
INNER JOIN customers c ON ci.customer_id = c.id
WHERE ci.id = $1
LIMIT 1;

-- name: ListCustomerInvoices :many
SELECT 
    ci.id,
    ci.invoice_number,
    ci.customer_id,
    c.name as customer_name,
    ci.invoice_date,
    ci.due_date,
    ci.total_amount,
    ci.paid_amount,
    (ci.total_amount - ci.paid_amount) as outstanding,
    ci.status,
    ci.created_at
FROM customer_invoices ci
INNER JOIN customers c ON ci.customer_id = c.id
WHERE ($1::uuid IS NULL OR ci.customer_id = $1)
    AND ($2::varchar IS NULL OR ci.status = $2)
ORDER BY ci.due_date DESC, ci.invoice_number DESC
LIMIT $3 OFFSET $4;

-- name: UpdateCustomerInvoicePaidAmount :exec
UPDATE customer_invoices
SET paid_amount = $2,
    status = CASE
        WHEN $2 >= total_amount THEN 'paid'
        WHEN $2 > 0 THEN 'partial'
        ELSE 'unpaid'
    END,
    updated_at = NOW()
WHERE id = $1;