-- Add unique constraint to prevent duplicate shipping discrepancies
ALTER TABLE shipping_discrepancies ADD CONSTRAINT unique_shipping_discrepancy 
    UNIQUE (invoice_number, discrepancy_type);

-- Rename order_id to return_id
ALTER TABLE shipping_discrepancies RENAME COLUMN order_id TO return_id;

-- Update the index
DROP INDEX IF EXISTS idx_shipping_discrepancies_order_id;
CREATE INDEX idx_shipping_discrepancies_return_id ON shipping_discrepancies(return_id);