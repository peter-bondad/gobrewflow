CREATE TABLE IF NOT EXISTS inventory_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    product_id UUID NOT NULL,

    type VARCHAR(20) NOT NULL,

    quantity INTEGER NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_inventory_movements_product
        FOREIGN KEY (product_id)
        REFERENCES products (id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_inventory_movements_quantity_positive
        CHECK (quantity > 0),

    CONSTRAINT chk_inventory_movements_type
        CHECK (type IN (
            'RECEIVED',
            'SOLD',
            'RETURNED',
            'DAMAGED',
            'ADJUSTED'
        ))
);