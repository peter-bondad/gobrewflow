-- Seed categories, products, and inventory for local development and testing
-- Run after: task migrate-up

-- Categories
INSERT INTO categories (id, name, is_active)
VALUES
  (gen_random_uuid(), 'Coffee Beans', TRUE),
  (gen_random_uuid(), 'Brewed Coffee', TRUE),
  (gen_random_uuid(), 'Pastries', TRUE),
  (gen_random_uuid(), 'Merchandise', TRUE),
  (gen_random_uuid(), 'Equipment', TRUE),
  (gen_random_uuid(), 'Syrups & Sauces', TRUE)
ON CONFLICT (name) DO NOTHING;

-- Products
-- We reference categories by name via a CTE to avoid hardcoding UUIDs.
WITH category_lookup AS (
  SELECT id, name FROM categories
)
INSERT INTO products (id, name, sku, slug, description, price, is_active, category_id, image_url)
SELECT
  gen_random_uuid(),
  p.name,
  p.sku,
  p.slug,
  p.description,
  p.price,
  TRUE,
  c.id,
  p.image_url
FROM (VALUES
  ('Arabica Single Origin', 'PRD-000001', 'arabica-single-origin', 'Premium single-origin arabica beans.', 19500, 'https://images.brewflow.test/arabica.jpg'),
  ('Robusta House Blend', 'PRD-000002', 'robusta-house-blend', 'House blend with a strong, bold flavor.', 16500, 'https://images.brewflow.test/robusta.jpg'),
  ('Signature Cold Brew', 'PRD-000003', 'signature-cold-brew', 'Smooth cold brew steeped for 18 hours.', 12000, 'https://images.brewflow.test/cold-brew.jpg'),
  ('Caramel Macchiato', 'PRD-000004', 'caramel-macchiato', 'Classic macchiato with caramel drizzle.', 14000, 'https://images.brewflow.test/macchiato.jpg'),
  ('Butter Croissant', 'PRD-000005', 'butter-croissant', 'Flaky, buttery French-style croissant.', 8500, 'https://images.brewflow.test/croissant.jpg'),
  ('Blueberry Muffin', 'PRD-000006', 'blueberry-muffin', 'Moist muffin loaded with wild blueberries.', 7500, 'https://images.brewflow.test/muffin.jpg'),
  ('Brewflow Tumbler', 'PRD-000007', 'brewflow-tumbler', 'Insulated stainless steel tumbler.', 24000, 'https://images.brewflow.test/tumbler.jpg'),
  ('Pour-Over Kit', 'PRD-000008', 'pour-over-kit', 'Complete pour-over brewing kit.', 35000, 'https://images.brewflow.test/pour-over.jpg'),
  ('Vanilla Syrup', 'PRD-000009', 'vanilla-syrup', 'Rich vanilla-flavored syrup.', 9000, 'https://images.brewflow.test/vanilla.jpg'),
  ('Dark Chocolate Sauce', 'PRD-000010', 'dark-chocolate-sauce', 'Premium dark chocolate sauce.', 9500, 'https://images.brewflow.test/chocolate.jpg'),
  ('Hazelnut Syrup', 'PRD-000011', 'hazelnut-syrup', 'Roasted hazelnut flavored syrup.', 9000, 'https://images.brewflow.test/hazelnut.jpg'),
  ('Matcha Latte', 'PRD-000012', 'matcha-latte', 'Smooth matcha with steamed milk.', 15000, 'https://images.brewflow.test/matcha.jpg')
) AS p(name, sku, slug, description, price, image_url)
JOIN category_lookup c ON c.name = CASE
  WHEN p.name IN ('Arabica Single Origin', 'Robusta House Blend') THEN 'Coffee Beans'
  WHEN p.name IN ('Signature Cold Brew', 'Caramel Macchiato', 'Matcha Latte') THEN 'Brewed Coffee'
  WHEN p.name IN ('Butter Croissant', 'Blueberry Muffin') THEN 'Pastries'
  WHEN p.name IN ('Brewflow Tumbler', 'Pour-Over Kit') THEN 'Merchandise'
  WHEN p.name IN ('Vanilla Syrup', 'Dark Chocolate Sauce', 'Hazelnut Syrup') THEN 'Syrups & Sauces'
  ELSE 'Equipment'
END;

-- Inventory
-- Each product gets one unique inventory row linked by product_id.
INSERT INTO inventory (id, product_id, quantity)
SELECT gen_random_uuid(), p.id, i.quantity
FROM products p
JOIN (VALUES
  ('arabica-single-origin', 150),
  ('robusta-house-blend', 200),
  ('signature-cold-brew', 80),
  ('caramel-macchiato', 120),
  ('butter-croissant', 45),
  ('blueberry-muffin', 60),
  ('brewflow-tumbler', 30),
  ('pour-over-kit', 25),
  ('vanilla-syrup', 90),
  ('dark-chocolate-sauce', 85),
  ('hazelnut-syrup', 90),
  ('matcha-latte', 100)
) AS i(slug, quantity) ON i.slug = p.slug;
