ALTER TABLE inventory_movements
    DROP CONSTRAINT chk_inventory_movements_stock_change,
    DROP CONSTRAINT chk_inventory_movements_after_stock_non_negative,
    DROP CONSTRAINT chk_inventory_movements_before_stock_non_negative;

ALTER TABLE inventory_movements
    DROP COLUMN after_stock,
    DROP COLUMN before_stock;