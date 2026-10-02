ALTER TABLE inventory_movements
DROP CONSTRAINT chk_inventory_movements_type;

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