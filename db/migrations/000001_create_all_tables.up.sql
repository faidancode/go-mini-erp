-- =====================================================
-- SCHEMA: Mini ERP System
-- Database: PostgreSQL 14+
-- =====================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- RBAC & Authentication
-- =====================================================

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_roles_code ON roles(code) WHERE is_active = true;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;

CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    assigned_by UUID REFERENCES users(id),
    UNIQUE(user_id, role_id)
);

CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);

CREATE TABLE menus (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_id UUID REFERENCES menus(id) ON DELETE CASCADE,
    code VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    path VARCHAR(500),
    icon VARCHAR(100),
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_menus_parent ON menus(parent_id) WHERE is_active = true;
CREATE INDEX idx_menus_code ON menus(code) WHERE is_active = true;

CREATE TABLE role_menus (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    menu_id UUID NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
    can_create BOOLEAN DEFAULT false,
    can_read BOOLEAN DEFAULT true,
    can_update BOOLEAN DEFAULT false,
    can_delete BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(role_id, menu_id)
);

CREATE INDEX idx_role_menus_role ON role_menus(role_id);
CREATE INDEX idx_role_menus_menu ON role_menus(menu_id);

-- =====================================================
-- Master Data
-- =====================================================

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES categories(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_categories_parent ON categories(parent_id) WHERE is_active = true;

CREATE TABLE units_of_measure (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    unit_type VARCHAR(50) NOT NULL, -- 'unit', 'weight', 'volume', 'length'
    is_base_unit BOOLEAN DEFAULT false,
    conversion_factor DECIMAL(15,6) DEFAULT 1.0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_uom_type ON units_of_measure(unit_type) WHERE is_active = true;

CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id UUID REFERENCES categories(id),
    uom_id UUID NOT NULL REFERENCES units_of_measure(id),
    product_type VARCHAR(20) DEFAULT 'storable', -- 'storable', 'consumable', 'service'
    cost_price DECIMAL(15,2) DEFAULT 0,
    sale_price DECIMAL(15,2) DEFAULT 0,
    min_stock DECIMAL(15,3) DEFAULT 0,
    max_stock DECIMAL(15,3) DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_products_code ON products(code) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_category ON products(category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_type ON products(product_type) WHERE is_active = true;

CREATE TABLE suppliers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    contact_person VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    tax_id VARCHAR(100),
    payment_terms INTEGER DEFAULT 30, -- days
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_suppliers_code ON suppliers(code) WHERE deleted_at IS NULL;

CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    contact_person VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    tax_id VARCHAR(100),
    credit_limit DECIMAL(15,2) DEFAULT 0,
    payment_terms INTEGER DEFAULT 30, -- days
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_customers_code ON customers(code) WHERE deleted_at IS NULL;

-- =====================================================
-- Procurement Module
-- =====================================================

CREATE TABLE purchase_orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    po_number VARCHAR(50) UNIQUE NOT NULL,
    supplier_id UUID NOT NULL REFERENCES suppliers(id),
    order_date DATE NOT NULL,
    expected_date DATE,
    status VARCHAR(20) DEFAULT 'draft', -- 'draft', 'confirmed', 'received', 'cancelled'
    total_amount DECIMAL(15,2) DEFAULT 0,
    notes TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_po_number ON purchase_orders(po_number);
CREATE INDEX idx_po_supplier ON purchase_orders(supplier_id);
CREATE INDEX idx_po_status ON purchase_orders(status);
CREATE INDEX idx_po_date ON purchase_orders(order_date);

CREATE TABLE purchase_order_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    po_id UUID NOT NULL REFERENCES purchase_orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity DECIMAL(15,3) NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(15,2) NOT NULL,
    subtotal DECIMAL(15,2) GENERATED ALWAYS AS (quantity * unit_price) STORED,
    received_qty DECIMAL(15,3) DEFAULT 0,
    line_number INTEGER NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(po_id, line_number)
);

CREATE INDEX idx_po_lines_po ON purchase_order_lines(po_id);
CREATE INDEX idx_po_lines_product ON purchase_order_lines(product_id);

CREATE TABLE goods_receipts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    receipt_number VARCHAR(50) UNIQUE NOT NULL,
    po_id UUID NOT NULL REFERENCES purchase_orders(id),
    receipt_date DATE NOT NULL,
    notes TEXT,
    received_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_gr_number ON goods_receipts(receipt_number);
CREATE INDEX idx_gr_po ON goods_receipts(po_id);

CREATE TABLE goods_receipt_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    receipt_id UUID NOT NULL REFERENCES goods_receipts(id) ON DELETE CASCADE,
    po_line_id UUID NOT NULL REFERENCES purchase_order_lines(id),
    product_id UUID NOT NULL REFERENCES products(id),
    quantity DECIMAL(15,3) NOT NULL CHECK (quantity > 0),
    line_number INTEGER NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(receipt_id, line_number)
);

CREATE INDEX idx_gr_lines_receipt ON goods_receipt_lines(receipt_id);
CREATE INDEX idx_gr_lines_po_line ON goods_receipt_lines(po_line_id);

CREATE TABLE supplier_bills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    bill_number VARCHAR(50) UNIQUE NOT NULL,
    supplier_id UUID NOT NULL REFERENCES suppliers(id),
    po_id UUID REFERENCES purchase_orders(id),
    bill_date DATE NOT NULL,
    due_date DATE NOT NULL,
    total_amount DECIMAL(15,2) NOT NULL,
    paid_amount DECIMAL(15,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'unpaid', -- 'unpaid', 'partial', 'paid', 'cancelled'
    notes TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_bill_number ON supplier_bills(bill_number);
CREATE INDEX idx_bill_supplier ON supplier_bills(supplier_id);
CREATE INDEX idx_bill_status ON supplier_bills(status);
CREATE INDEX idx_bill_due_date ON supplier_bills(due_date);

-- =====================================================
-- Sales Module
-- =====================================================

CREATE TABLE quotations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quote_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers(id),
    quote_date DATE NOT NULL,
    valid_until DATE,
    status VARCHAR(20) DEFAULT 'draft', -- 'draft', 'sent', 'confirmed', 'cancelled'
    total_amount DECIMAL(15,2) DEFAULT 0,
    notes TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_quote_number ON quotations(quote_number);
CREATE INDEX idx_quote_customer ON quotations(customer_id);
CREATE INDEX idx_quote_status ON quotations(status);

CREATE TABLE quotation_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quote_id UUID NOT NULL REFERENCES quotations(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity DECIMAL(15,3) NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(15,2) NOT NULL,
    subtotal DECIMAL(15,2) GENERATED ALWAYS AS (quantity * unit_price) STORED,
    line_number INTEGER NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(quote_id, line_number)
);

CREATE INDEX idx_quote_lines_quote ON quotation_lines(quote_id);
CREATE INDEX idx_quote_lines_product ON quotation_lines(product_id);

CREATE TABLE sales_orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    so_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers(id),
    quote_id UUID REFERENCES quotations(id),
    order_date DATE NOT NULL,
    delivery_date DATE,
    status VARCHAR(20) DEFAULT 'draft', -- 'draft', 'confirmed', 'delivered', 'cancelled'
    total_amount DECIMAL(15,2) DEFAULT 0,
    notes TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_so_number ON sales_orders(so_number);
CREATE INDEX idx_so_customer ON sales_orders(customer_id);
CREATE INDEX idx_so_status ON sales_orders(status);
CREATE INDEX idx_so_date ON sales_orders(order_date);

CREATE TABLE sales_order_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    so_id UUID NOT NULL REFERENCES sales_orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity DECIMAL(15,3) NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(15,2) NOT NULL,
    subtotal DECIMAL(15,2) GENERATED ALWAYS AS (quantity * unit_price) STORED,
    delivered_qty DECIMAL(15,3) DEFAULT 0,
    line_number INTEGER NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(so_id, line_number)
);

CREATE INDEX idx_so_lines_so ON sales_order_lines(so_id);
CREATE INDEX idx_so_lines_product ON sales_order_lines(product_id);

CREATE TABLE customer_invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    customer_id UUID NOT NULL REFERENCES customers(id),
    so_id UUID REFERENCES sales_orders(id),
    invoice_date DATE NOT NULL,
    due_date DATE NOT NULL,
    total_amount DECIMAL(15,2) NOT NULL,
    paid_amount DECIMAL(15,2) DEFAULT 0,
    status VARCHAR(20) DEFAULT 'unpaid', -- 'unpaid', 'partial', 'paid', 'cancelled'
    notes TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_invoice_number ON customer_invoices(invoice_number);
CREATE INDEX idx_invoice_customer ON customer_invoices(customer_id);
CREATE INDEX idx_invoice_status ON customer_invoices(status);
CREATE INDEX idx_invoice_due_date ON customer_invoices(due_date);

-- =====================================================
-- Inventory/Warehouse Module
-- =====================================================

CREATE TABLE stock_locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    location_type VARCHAR(20) DEFAULT 'warehouse', -- 'warehouse', 'retail', 'transit'
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE stock_balances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    product_id UUID NOT NULL REFERENCES products(id),
    location_id UUID NOT NULL REFERENCES stock_locations(id),
    quantity DECIMAL(15,3) DEFAULT 0,
    reserved_qty DECIMAL(15,3) DEFAULT 0,
    available_qty DECIMAL(15,3) GENERATED ALWAYS AS (quantity - reserved_qty) STORED,
    last_updated TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(product_id, location_id)
);

CREATE INDEX idx_stock_product ON stock_balances(product_id);
CREATE INDEX idx_stock_location ON stock_balances(location_id);

CREATE TABLE stock_movements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    movement_number VARCHAR(50) UNIQUE NOT NULL,
    product_id UUID NOT NULL REFERENCES products(id),
    location_id UUID NOT NULL REFERENCES stock_locations(id),
    movement_type VARCHAR(20) NOT NULL, -- 'in', 'out', 'adjustment', 'transfer'
    quantity DECIMAL(15,3) NOT NULL,
    reference_type VARCHAR(50), -- 'purchase_order', 'sales_order', 'adjustment', 'transfer'
    reference_id UUID,
    movement_date TIMESTAMPTZ NOT NULL,
    notes TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_movement_product ON stock_movements(product_id);
CREATE INDEX idx_movement_location ON stock_movements(location_id);
CREATE INDEX idx_movement_type ON stock_movements(movement_type);
CREATE INDEX idx_movement_date ON stock_movements(movement_date);
CREATE INDEX idx_movement_reference ON stock_movements(reference_type, reference_id);

CREATE TABLE stock_adjustments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    adjustment_number VARCHAR(50) UNIQUE NOT NULL,
    location_id UUID NOT NULL REFERENCES stock_locations(id),
    adjustment_date DATE NOT NULL,
    reason VARCHAR(255),
    status VARCHAR(20) DEFAULT 'draft', -- 'draft', 'confirmed'
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_adjustment_number ON stock_adjustments(adjustment_number);
CREATE INDEX idx_adjustment_location ON stock_adjustments(location_id);

CREATE TABLE stock_adjustment_lines (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    adjustment_id UUID NOT NULL REFERENCES stock_adjustments(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    quantity_before DECIMAL(15,3) NOT NULL,
    quantity_after DECIMAL(15,3) NOT NULL,
    difference DECIMAL(15,3) GENERATED ALWAYS AS (quantity_after - quantity_before) STORED,
    line_number INTEGER NOT NULL,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(adjustment_id, line_number)
);

CREATE INDEX idx_adj_lines_adjustment ON stock_adjustment_lines(adjustment_id);
CREATE INDEX idx_adj_lines_product ON stock_adjustment_lines(product_id);

-- =====================================================
-- Finance Module
-- =====================================================

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    payment_number VARCHAR(50) UNIQUE NOT NULL,
    payment_type VARCHAR(20) NOT NULL, -- 'receivable', 'payable'
    partner_type VARCHAR(20) NOT NULL, -- 'customer', 'supplier'
    partner_id UUID NOT NULL,
    invoice_id UUID, -- can be customer_invoice or supplier_bill
    payment_date DATE NOT NULL,
    amount DECIMAL(15,2) NOT NULL CHECK (amount > 0),
    payment_method VARCHAR(50), -- 'cash', 'bank_transfer', 'check', 'card'
    reference VARCHAR(255),
    notes TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_payment_number ON payments(payment_number);
CREATE INDEX idx_payment_type ON payments(payment_type);
CREATE INDEX idx_payment_partner ON payments(partner_type, partner_id);
CREATE INDEX idx_payment_invoice ON payments(invoice_id);
CREATE INDEX idx_payment_date ON payments(payment_date);

-- =====================================================
-- Audit & System
-- =====================================================

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id),
    table_name VARCHAR(100) NOT NULL,
    record_id UUID NOT NULL,
    action VARCHAR(20) NOT NULL, -- 'insert', 'update', 'delete'
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_table ON audit_logs(table_name);
CREATE INDEX idx_audit_record ON audit_logs(record_id);
CREATE INDEX idx_audit_user ON audit_logs(user_id);
CREATE INDEX idx_audit_created ON audit_logs(created_at);