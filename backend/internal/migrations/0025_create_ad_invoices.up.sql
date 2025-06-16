CREATE TABLE IF NOT EXISTS ad_invoices (
    invoice_no VARCHAR(64) PRIMARY KEY,
    username VARCHAR(64),
    store VARCHAR(128),
    invoice_date DATE,
    total NUMERIC,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
