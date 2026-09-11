CREATE TABLE IF NOT EXISTS inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    product_id UUID NOT NULL UNIQUE,

    quantity INTEGER NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_inventory_product
        FOREIGN KEY (product_id)
        REFERENCES products (id)
        ON DELETE CASCADE,

    CONSTRAINT chk_inventory_quantity_non_negative
        CHECK (quantity >= 0)
);
