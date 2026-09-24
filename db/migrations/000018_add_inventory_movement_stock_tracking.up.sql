ALTER TABLE inventory_movements
    ADD COLUMN before_stock INTEGER,
    ADD COLUMN after_stock INTEGER;

TRUNCATE TABLE inventory_movements;

ALTER TABLE inventory_movements
    ALTER COLUMN before_stock SET NOT NULL,
    ALTER COLUMN after_stock SET NOT NULL;

ALTER TABLE inventory_movements
    ADD CONSTRAINT chk_inventory_movements_before_stock_non_negative
        CHECK (before_stock >= 0),

    ADD CONSTRAINT chk_inventory_movements_after_stock_non_negative
        CHECK (after_stock >= 0),

    ADD CONSTRAINT chk_inventory_movements_stock_change
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
        );