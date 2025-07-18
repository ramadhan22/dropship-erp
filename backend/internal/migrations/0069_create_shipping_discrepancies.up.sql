CREATE TABLE shipping_discrepancies (
    id BIGSERIAL PRIMARY KEY,
    invoice_number VARCHAR(255) NOT NULL,
    order_id VARCHAR(255),
    discrepancy_type VARCHAR(50) NOT NULL, -- 'selisih_ongkir' or 'reverse_shipping_fee'
    discrepancy_amount NUMERIC(10,2) NOT NULL,
    actual_shipping_fee NUMERIC(10,2),
    buyer_paid_shipping_fee NUMERIC(10,2),
    shopee_shipping_rebate NUMERIC(10,2),
    seller_shipping_discount NUMERIC(10,2),
    reverse_shipping_fee NUMERIC(10,2),
    order_date TIMESTAMP,
    store_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Add indexes for better query performance
CREATE INDEX idx_shipping_discrepancies_invoice ON shipping_discrepancies(invoice_number);
CREATE INDEX idx_shipping_discrepancies_order_id ON shipping_discrepancies(order_id);
CREATE INDEX idx_shipping_discrepancies_type ON shipping_discrepancies(discrepancy_type);
CREATE INDEX idx_shipping_discrepancies_store ON shipping_discrepancies(store_name);
CREATE INDEX idx_shipping_discrepancies_created_at ON shipping_discrepancies(created_at);