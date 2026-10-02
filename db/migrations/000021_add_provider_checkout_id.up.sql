ALTER TABLE payments
ADD COLUMN provider_checkout_id VARCHAR(255);

ALTER TABLE payments
ADD CONSTRAINT uq_payments_provider_checkout_id
UNIQUE (provider, provider_checkout_id);