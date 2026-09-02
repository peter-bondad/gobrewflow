-- Seed users for local development and testing
-- Run after: task migrate-up

INSERT INTO users (id, email, full_name, first_name, middle_name, last_name, phone_number, role)
VALUES
  ('00000000-0000-0000-0000-000000000001', 'owner@brewflow.test', 'Alice Owner', 'Alice', NULL, 'Owner', '+15550000001', 'owner'),
  ('00000000-0000-0000-0000-000000000002', 'manager@brewflow.test', 'Bob Manager', 'Bob', NULL, 'Manager', '+15550000002', 'manager'),
  ('00000000-0000-0000-0000-000000000003', 'staff@brewflow.test', 'Carol Staff', 'Carol', NULL, 'Staff', '+15550000003', 'staff')
ON CONFLICT (id) DO NOTHING;

-- Seed accounts with password hashes for local development
-- Default password for all test accounts: "password"
INSERT INTO accounts (id, user_id, provider, provider_account_id, password_hash)
VALUES
  ('00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'credentials', 'owner@brewflow.test', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'),
  ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000002', 'credentials', 'manager@brewflow.test', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'),
  ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000003', 'credentials', 'staff@brewflow.test', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi')
ON CONFLICT (id) DO NOTHING;