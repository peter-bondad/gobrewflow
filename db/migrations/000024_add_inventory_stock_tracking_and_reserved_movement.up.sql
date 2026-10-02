ALTER TABLE inventory_movements
    DROP CONSTRAINT chk_inventory_movements_type,
    DROP CONSTRAINT chk_inventory_movements_stock_change;

ALTER TABLE inventory_movements
    ADD CONSTRAINT chk_inventory_movements_type
        CHECK (type IN (
            'RECEIVED',
            'RESERVED',
            'SOLD',
            'RETURNED',
            'DAMAGED',
            'ADJUSTED'
        ));

ALTER TABLE inventory_movements
    ADD CONSTRAINT chk_inventory_movements_stock_change
        CHECK (
            (
                type IN ('RECEIVED', 'RETURNED')
                AND after_stock = before_stock + quantity
            )
            OR
            (
                type IN ('RESERVED', 'SOLD', 'DAMAGED')
                AND after_stock = before_stock - quantity
            )
            OR
            (
                type = 'ADJUSTED'
                AND ABS(after_stock - before_stock) = quantity
            )
        );