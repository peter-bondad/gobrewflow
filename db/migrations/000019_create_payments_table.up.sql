CREATE TYPE payment_provider AS ENUM (
    'PAYMONGO'
);

CREATE TYPE payment_status AS ENUM (
    'PENDING',
    'PAID',
    'FAILED',
    'CANCELLED',
    'REFUNDED'
);

CREATE TYPE payment_method AS ENUM (
    'CARD',
    'GCASH',
    'GRABPAY',
    'MAYA',
    'QRPH'
);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    order_id UUID NOT NULL
        REFERENCES orders(id),

    provider payment_provider NOT NULL,

    provider_payment_id VARCHAR(255) NOT NULL,

    amount BIGINT NOT NULL
        CHECK (amount > 0),

    currency VARCHAR(3) NOT NULL
        DEFAULT 'PHP'
        CHECK (currency = 'PHP'),

    status payment_status NOT NULL
        DEFAULT 'PENDING',

    payment_method payment_method,

    paid_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL
        DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL
        DEFAULT NOW(),

    CONSTRAINT uq_payments_provider_payment_id
        UNIQUE (provider, provider_payment_id)
);

CREATE INDEX idx_payments_order_id
    ON payments(order_id);