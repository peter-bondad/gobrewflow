CREATE SEQUENCE IF NOT EXISTS order_number_seq;

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    order_number TEXT NOT NULL UNIQUE DEFAULT (
        'ORD-' || LPAD(nextval('order_number_seq')::text, 6, '0')
    ),

    status TEXT NOT NULL DEFAULT 'pending',

    -- Store money in the smallest currency unit.
    subtotal BIGINT NOT NULL DEFAULT 0,
    tax BIGINT NOT NULL DEFAULT 0,
    discount BIGINT NOT NULL DEFAULT 0,
    total BIGINT NOT NULL DEFAULT 0,

    cashier_id UUID NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Set when the order is completed.
    completed_at TIMESTAMPTZ,

    -- Set when the order is cancelled.
    cancelled_at TIMESTAMPTZ,

    CONSTRAINT fk_orders_cashier
        FOREIGN KEY (cashier_id)
        REFERENCES users (id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_orders_status
        CHECK (status IN ('pending', 'completed', 'cancelled')),

    CONSTRAINT chk_orders_subtotal_non_negative
        CHECK (subtotal >= 0),

    CONSTRAINT chk_orders_tax_non_negative
        CHECK (tax >= 0),

    CONSTRAINT chk_orders_discount_non_negative
        CHECK (discount >= 0),

    CONSTRAINT chk_orders_total_non_negative
        CHECK (total >= 0),

    -- A completed order cannot have a cancelled timestamp,
    -- and a cancelled order cannot have a completed timestamp.
    CONSTRAINT chk_orders_status_timestamps
        CHECK (
            (status = 'pending' AND completed_at IS NULL AND cancelled_at IS NULL)
            OR
            (status = 'completed' AND completed_at IS NOT NULL AND cancelled_at IS NULL)
            OR
            (status = 'cancelled' AND completed_at IS NULL AND cancelled_at IS NOT NULL)
        )
);
