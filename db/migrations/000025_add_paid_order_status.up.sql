ALTER TABLE orders
    DROP CONSTRAINT chk_orders_status,
    DROP CONSTRAINT chk_orders_status_timestamps;

ALTER TABLE orders
    ADD CONSTRAINT chk_orders_status
        CHECK (status IN ('pending', 'paid', 'completed', 'cancelled'));

ALTER TABLE orders
    ADD CONSTRAINT chk_orders_status_timestamps
        CHECK (
            (status IN ('pending', 'paid') AND completed_at IS NULL AND cancelled_at IS NULL)
            OR
            (status = 'completed' AND completed_at IS NOT NULL AND cancelled_at IS NULL)
            OR
            (status = 'cancelled' AND completed_at IS NULL AND cancelled_at IS NOT NULL)
        );