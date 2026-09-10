CREATE SEQUENCE IF NOT EXISTS product_sku_seq;

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name TEXT NOT NULL,

    sku TEXT NOT NULL UNIQUE DEFAULT (
        'PRD-' || LPAD(nextval('product_sku_seq')::text, 6, '0')
    ),

    slug TEXT NOT NULL UNIQUE,

    description TEXT,

    -- Store money in the smallest currency unit.
    -- Example: ₱199.50 = 19950 centavos.
    price BIGINT NOT NULL DEFAULT 0,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    category_id UUID NOT NULL,

    image_url TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_products_category
        FOREIGN KEY (category_id)
        REFERENCES categories (id)
        ON DELETE RESTRICT,

    CONSTRAINT chk_products_price_non_negative
        CHECK (price >= 0)
);

CREATE INDEX IF NOT EXISTS idx_products_category_id
    ON products(category_id);

CREATE INDEX IF NOT EXISTS idx_products_name
    ON products(name);

CREATE INDEX IF NOT EXISTS idx_products_is_active
    ON products(is_active);