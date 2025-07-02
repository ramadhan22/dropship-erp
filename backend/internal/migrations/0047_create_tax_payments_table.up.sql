CREATE TABLE tax_payments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  store TEXT NOT NULL,
  period_type TEXT NOT NULL,
  period_value TEXT NOT NULL,
  revenue NUMERIC NOT NULL,
  tax_rate NUMERIC NOT NULL DEFAULT 0.005,
  tax_amount NUMERIC NOT NULL,
  is_paid BOOLEAN NOT NULL DEFAULT FALSE,
  paid_at TIMESTAMP
);
