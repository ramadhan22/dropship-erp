-- Reverse the changes
DROP INDEX IF EXISTS idx_shipping_discrepancies_return_id;
ALTER TABLE shipping_discrepancies RENAME COLUMN return_id TO order_id;
CREATE INDEX idx_shipping_discrepancies_order_id ON shipping_discrepancies(order_id);
ALTER TABLE shipping_discrepancies DROP CONSTRAINT IF EXISTS unique_shipping_discrepancy;