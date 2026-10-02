ALTER TABLE inventory_movements
    ADD CONSTRAINT chk_inventory_movements_type
        CHECK (type IN (
            'RECEIVED',
            'SOLD',
            'RETURNED',
            'DAMAGED',
            'ADJUSTED'
        ));