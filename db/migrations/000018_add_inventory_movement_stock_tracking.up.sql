-- V2: Add stock state tracking to inventory movements

CREATE TABLE inventory_movements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    product_id UUID NOT NULL,

    type VARCHAR(20) NOT NULL,

    quantity INTEGER NOT NULL,

    before_stock INTEGER NOT NULL,

    after_stock INTEGER NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_inventory_movements_product
        FOREIGN KEY (product_id)
        REFERENCES products (id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_inventory_movements_quantity_positive
        CHECK (quantity > 0),

    CONSTRAINT chk_inventory_movements_before_stock_non_negative
        CHECK (before_stock >= 0),

    CONSTRAINT chk_inventory_movements_after_stock_non_negative
        CHECK (after_stock >= 0),

    CONSTRAINT chk_inventory_movements_type
        CHECK (type IN (
            'RECEIVED',
            'SOLD',
            'RETURNED',
            'DAMAGED',
            'ADJUSTED'
        )),

    CONSTRAINT chk_inventory_movements_stock_change
        CHECK (
            (
                type IN ('RECEIVED', 'RETURNED')
                AND after_stock = before_stock + quantity
            )
            OR
            (
                type IN ('SOLD', 'DAMAGED')
                AND after_stock = before_stock - quantity
            )
            OR
            (
                type = 'ADJUSTED'
                AND ABS(after_stock - before_stock) = quantity
            )
        )
);