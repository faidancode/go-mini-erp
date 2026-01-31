-- name: CreatePayment :one
INSERT INTO payments (
    payment_number,
    payment_type,
    partner_type,
    partner_id,
    invoice_id,
    payment_date,
    amount,
    payment_method,
    reference,
    notes,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING id, payment_number, payment_type, partner_id, payment_date, amount, created_at;

-- name: GetPaymentByID :one
SELECT 
    p.id,
    p.payment_number,
    p.payment_type,
    p.partner_type,
    p.partner_id,
    p.invoice_id,
    p.payment_date,
    p.amount,
    p.payment_method,
    p.reference,
    p.notes,
    p.created_at
FROM payments p
WHERE p.id = $1
LIMIT 1;

-- name: ListPayments :many
SELECT 
    p.id,
    p.payment_number,
    p.payment_type,
    p.partner_type,
    p.partner_id,
    CASE 
        WHEN p.partner_type = 'customer' THEN c.name
        WHEN p.partner_type = 'supplier' THEN s.name
    END as partner_name,
    p.payment_date,
    p.amount,
    p.payment_method,
    p.reference,
    p.created_at
FROM payments p
LEFT JOIN customers c ON p.partner_type = 'customer' AND p.partner_id = c.id
LEFT JOIN suppliers s ON p.partner_type = 'supplier' AND p.partner_id = s.id
WHERE ($1::varchar IS NULL OR p.payment_type = $1)
    AND ($2::varchar IS NULL OR p.partner_type = $2)
    AND ($3::uuid IS NULL OR p.partner_id = $3)
    AND ($4::date IS NULL OR p.payment_date >= $4)
    AND ($5::date IS NULL OR p.payment_date <= $5)
ORDER BY p.payment_date DESC, p.payment_number DESC
LIMIT $6 OFFSET $7;

-- name: GetAccountsReceivableSummary :many
SELECT 
    ci.customer_id,
    c.code as customer_code,
    c.name as customer_name,
    COUNT(ci.id) as invoice_count,
    SUM(ci.total_amount) as total_invoiced,
    SUM(ci.paid_amount) as total_paid,
    SUM(ci.total_amount - ci.paid_amount) as outstanding,
    SUM(CASE WHEN ci.status = 'unpaid' THEN 1 ELSE 0 END) as unpaid_count,
    SUM(CASE WHEN ci.status = 'partial' THEN 1 ELSE 0 END) as partial_count,
    MIN(CASE WHEN ci.status != 'paid' THEN ci.due_date END) as earliest_due_date,
    SUM(CASE 
        WHEN ci.status != 'paid' AND ci.due_date < CURRENT_DATE 
        THEN ci.total_amount - ci.paid_amount 
        ELSE 0 
    END) as overdue_amount
FROM customer_invoices ci
INNER JOIN customers c ON ci.customer_id = c.id
WHERE ci.status != 'cancelled'
    AND ($1::uuid IS NULL OR ci.customer_id = $1)
GROUP BY ci.customer_id, c.code, c.name
HAVING SUM(ci.total_amount - ci.paid_amount) > 0
ORDER BY outstanding DESC;

-- name: GetAccountsPayableSummary :many
SELECT 
    sb.supplier_id,
    s.code as supplier_code,
    s.name as supplier_name,
    COUNT(sb.id) as bill_count,
    SUM(sb.total_amount) as total_billed,
    SUM(sb.paid_amount) as total_paid,
    SUM(sb.total_amount - sb.paid_amount) as outstanding,
    SUM(CASE WHEN sb.status = 'unpaid' THEN 1 ELSE 0 END) as unpaid_count,
    SUM(CASE WHEN sb.status = 'partial' THEN 1 ELSE 0 END) as partial_count,
    MIN(CASE WHEN sb.status != 'paid' THEN sb.due_date END) as earliest_due_date,
    SUM(CASE 
        WHEN sb.status != 'paid' AND sb.due_date < CURRENT_DATE 
        THEN sb.total_amount - sb.paid_amount 
        ELSE 0 
    END) as overdue_amount
FROM supplier_bills sb
INNER JOIN suppliers s ON sb.supplier_id = s.id
WHERE sb.status != 'cancelled'
    AND ($1::uuid IS NULL OR sb.supplier_id = $1)
GROUP BY sb.supplier_id, s.code, s.name
HAVING SUM(sb.total_amount - sb.paid_amount) > 0
ORDER BY outstanding DESC;

-- name: GetAgingReceivables :many
SELECT 
    ci.customer_id,
    c.code as customer_code,
    c.name as customer_name,
    SUM(CASE 
        WHEN ci.due_date >= CURRENT_DATE 
        THEN ci.total_amount - ci.paid_amount 
        ELSE 0 
    END) as current_amount,
    SUM(CASE 
        WHEN ci.due_date < CURRENT_DATE 
        AND ci.due_date >= CURRENT_DATE - INTERVAL '30 days'
        THEN ci.total_amount - ci.paid_amount 
        ELSE 0 
    END) as days_1_30,
    SUM(CASE 
        WHEN ci.due_date < CURRENT_DATE - INTERVAL '30 days'
        AND ci.due_date >= CURRENT_DATE - INTERVAL '60 days'
        THEN ci.total_amount - ci.paid_amount 
        ELSE 0 
    END) as days_31_60,
    SUM(CASE 
        WHEN ci.due_date < CURRENT_DATE - INTERVAL '60 days'
        AND ci.due_date >= CURRENT_DATE - INTERVAL '90 days'
        THEN ci.total_amount - ci.paid_amount 
        ELSE 0 
    END) as days_61_90,
    SUM(CASE 
        WHEN ci.due_date < CURRENT_DATE - INTERVAL '90 days'
        THEN ci.total_amount - ci.paid_amount 
        ELSE 0 
    END) as over_90_days,
    SUM(ci.total_amount - ci.paid_amount) as total_outstanding
FROM customer_invoices ci
INNER JOIN customers c ON ci.customer_id = c.id
WHERE ci.status != 'cancelled'
    AND ci.status != 'paid'
GROUP BY ci.customer_id, c.code, c.name
HAVING SUM(ci.total_amount - ci.paid_amount) > 0
ORDER BY total_outstanding DESC;

-- name: GetGrossProfitSummary :one
SELECT 
    COUNT(DISTINCT so.id) as total_orders,
    SUM(sol.quantity * sol.unit_price) as total_revenue,
    SUM(sol.quantity * p.cost_price) as total_cost,
    SUM(sol.quantity * sol.unit_price) - SUM(sol.quantity * p.cost_price) as gross_profit,
    CASE 
        WHEN SUM(sol.quantity * sol.unit_price) > 0 
        THEN ((SUM(sol.quantity * sol.unit_price) - SUM(sol.quantity * p.cost_price)) / SUM(sol.quantity * sol.unit_price)) * 100
        ELSE 0 
    END as gross_margin_percent
FROM sales_orders so
INNER JOIN sales_order_lines sol ON so.id = sol.so_id
INNER JOIN products p ON sol.product_id = p.id
WHERE so.status != 'cancelled'
    AND ($1::date IS NULL OR so.order_date >= $1)
    AND ($2::date IS NULL OR so.order_date <= $2);

-- name: GetGrossProfitByProduct :many
SELECT 
    p.id as product_id,
    p.code as product_code,
    p.name as product_name,
    c.name as category_name,
    SUM(sol.quantity) as total_qty_sold,
    SUM(sol.quantity * sol.unit_price) as total_revenue,
    SUM(sol.quantity * p.cost_price) as total_cost,
    SUM(sol.quantity * sol.unit_price) - SUM(sol.quantity * p.cost_price) as gross_profit,
    CASE 
        WHEN SUM(sol.quantity * sol.unit_price) > 0 
        THEN ((SUM(sol.quantity * sol.unit_price) - SUM(sol.quantity * p.cost_price)) / SUM(sol.quantity * sol.unit_price)) * 100
        ELSE 0 
    END as gross_margin_percent
FROM sales_orders so
INNER JOIN sales_order_lines sol ON so.id = sol.so_id
INNER JOIN products p ON sol.product_id = p.id
LEFT JOIN categories c ON p.category_id = c.id
WHERE so.status != 'cancelled'
    AND ($1::date IS NULL OR so.order_date >= $1)
    AND ($2::date IS NULL OR so.order_date <= $2)
    AND ($3::uuid IS NULL OR p.category_id = $3)
GROUP BY p.id, p.code, p.name, c.name
ORDER BY gross_profit DESC
LIMIT $4 OFFSET $5;

-- name: GetCashFlowSummary :one
SELECT 
    COALESCE(SUM(CASE WHEN payment_type = 'receivable' THEN amount ELSE 0 END), 0)::DECIMAL as cash_in,
    COALESCE(SUM(CASE WHEN payment_type = 'payable' THEN amount ELSE 0 END), 0)::DECIMAL as cash_out,
    (
        COALESCE(SUM(CASE WHEN payment_type = 'receivable' THEN amount ELSE 0 END), 0) - 
        COALESCE(SUM(CASE WHEN payment_type = 'payable' THEN amount ELSE 0 END), 0)
    )::DECIMAL as net_cash_flow
FROM payments
WHERE payment_date >= $1 AND payment_date <= $2;
