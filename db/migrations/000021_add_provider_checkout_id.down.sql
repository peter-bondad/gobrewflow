ALTER TABLE payments
DROP CONSTRAINT IF EXISTS uq_payments_provider_checkout_id;

ALTER TABLE payments
DROP COLUMN IF EXISTS provider_checkout_id;