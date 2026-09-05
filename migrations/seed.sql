-- =====================================================================
-- Dummy seed data for local development/testing.
-- Run AFTER migrations/0001_init.sql.
--
-- Includes bookings in every status so you can test the full lifecycle
-- (pending queue, confirm, cancel, pickup, return, expiry, lookup)
-- without manually creating each one through the API first.
-- =====================================================================

-- ---------------------------------------------------------------------
-- Employees
-- Login with: email = admin@rentalshop.test / password = password123
-- (bcrypt hash below was generated for that exact password)
-- ---------------------------------------------------------------------
INSERT INTO employees (name, phone, email, password_hash) VALUES
('Budi Santoso', '+6281234567890', 'admin@rentalshop.test',
 '$2b$12$r64.lJNjaMzg8tHtzwaebOW0q8U04fViKk1ulaBO9YLbUmSQ9b7Ti');

-- ---------------------------------------------------------------------
-- Categories
-- ---------------------------------------------------------------------
INSERT INTO categories (name, slug) VALUES
('Camping', 'camping'),
('Hiking', 'hiking'),
('Water Sports', 'water-sports'),
('Climbing', 'climbing');

-- ---------------------------------------------------------------------
-- Equipment items
-- ---------------------------------------------------------------------
INSERT INTO equipment_items (category_id, name, slug, description, daily_rate, total_quantity, condition_notes, photo_url, is_active) VALUES
((SELECT id FROM categories WHERE slug = 'camping'), '4-Person Dome Tent', '4-person-dome-tent', 'Spacious dome tent, easy setup, waterproof rainfly included.', 75000, 5, 'All units in good condition, checked monthly.', 'https://placehold.co/400x300?text=Tent', true),
((SELECT id FROM categories WHERE slug = 'camping'), '2-Person Backpacking Tent', '2-person-backpacking-tent', 'Lightweight tent, ideal for solo or duo trips.', 50000, 8, 'One unit has a minor zipper stiffness, still functional.', 'https://placehold.co/400x300?text=Tent+2P', true),
((SELECT id FROM categories WHERE slug = 'camping'), 'Sleeping Bag (0°C)', 'sleeping-bag-0c', 'Cold-weather sleeping bag rated to 0°C.', 25000, 15, 'Cleaned after every rental.', 'https://placehold.co/400x300?text=Sleeping+Bag', true),
((SELECT id FROM categories WHERE slug = 'camping'), 'Portable Camping Stove', 'portable-camping-stove', 'Single-burner gas stove, gas canister sold separately.', 20000, 10, NULL, 'https://placehold.co/400x300?text=Stove', true),
((SELECT id FROM categories WHERE slug = 'hiking'), '60L Hiking Backpack', '60l-hiking-backpack', 'Multi-day trekking backpack with rain cover.', 40000, 12, NULL, 'https://placehold.co/400x300?text=Backpack', true),
((SELECT id FROM categories WHERE slug = 'hiking'), 'Trekking Poles (Pair)', 'trekking-poles-pair', 'Adjustable aluminum trekking poles, sold as a pair.', 15000, 20, NULL, 'https://placehold.co/400x300?text=Poles', true),
((SELECT id FROM categories WHERE slug = 'hiking'), 'Headlamp', 'headlamp', 'LED headlamp, batteries included.', 10000, 25, 'Batteries replaced regularly.', 'https://placehold.co/400x300?text=Headlamp', true),
((SELECT id FROM categories WHERE slug = 'water-sports'), 'Inflatable Kayak (2-Person)', 'inflatable-kayak-2-person', 'Includes paddles and hand pump.', 100000, 4, 'One unit has a small repaired patch, holds pressure fine.', 'https://placehold.co/400x300?text=Kayak', true),
((SELECT id FROM categories WHERE slug = 'water-sports'), 'Life Jacket', 'life-jacket', 'Adult-size, adjustable straps.', 15000, 20, NULL, 'https://placehold.co/400x300?text=Life+Jacket', true),
((SELECT id FROM categories WHERE slug = 'climbing'), 'Climbing Harness', 'climbing-harness', 'Adjustable full-body harness, adult sizes.', 30000, 8, 'Inspected before every rental per safety policy.', 'https://placehold.co/400x300?text=Harness', true),
((SELECT id FROM categories WHERE slug = 'climbing'), 'Climbing Helmet', 'climbing-helmet', 'Adjustable, meets safety standards.', 20000, 8, NULL, 'https://placehold.co/400x300?text=Helmet', true),
((SELECT id FROM categories WHERE slug = 'climbing'), 'Dynamic Climbing Rope (60m)', 'dynamic-climbing-rope-60m', '60m dynamic rope, retired after 200 uses per safety policy.', 60000, 3, 'Usage count tracked manually, 40 uses so far.', 'https://placehold.co/400x300?text=Rope', true);

-- ---------------------------------------------------------------------
-- Bookings — one in each status, using dates around "today"
-- (adjust relative dates if you're seeding well after 2026-09-03)
-- ---------------------------------------------------------------------

-- 1) PENDING — created 10 min ago, still within the 30-min hold window
INSERT INTO bookings (reference, customer_name, customer_phone, start_date, end_date, status, created_at, expires_at) VALUES
('BK-A1B2C3', 'Siti Rahayu', '+6281111111111', '2026-09-05', '2026-09-07', 'pending', now() - interval '10 minutes', now() + interval '20 minutes');
INSERT INTO booking_items (booking_id, equipment_item_id, quantity)
SELECT id, (SELECT id FROM equipment_items WHERE slug = '4-person-dome-tent'), 1 FROM bookings WHERE reference = 'BK-A1B2C3';

-- 2) CONFIRMED — upcoming rental, employee already confirmed
INSERT INTO bookings (reference, customer_name, customer_phone, start_date, end_date, status, created_at, expires_at, confirmed_at) VALUES
('BK-D4E5F6', 'Andi Wijaya', '+6282222222222', '2026-09-10', '2026-09-12', 'confirmed', now() - interval '1 day', now() - interval '1 day' + interval '30 minutes', now() - interval '23 hours');
INSERT INTO booking_items (booking_id, equipment_item_id, quantity) VALUES
((SELECT id FROM bookings WHERE reference = 'BK-D4E5F6'), (SELECT id FROM equipment_items WHERE slug = 'inflatable-kayak-2-person'), 1),
((SELECT id FROM bookings WHERE reference = 'BK-D4E5F6'), (SELECT id FROM equipment_items WHERE slug = 'life-jacket'), 2);

-- 3) ONGOING — customer already picked up, currently renting
INSERT INTO bookings (reference, customer_name, customer_phone, start_date, end_date, status, created_at, expires_at, confirmed_at, picked_up_at) VALUES
('BK-G7H8I9', 'Rina Kusuma', '+6283333333333', '2026-09-01', '2026-09-04', 'ongoing', now() - interval '3 days', now() - interval '3 days' + interval '30 minutes', now() - interval '3 days' + interval '10 minutes', now() - interval '2 days');
INSERT INTO booking_items (booking_id, equipment_item_id, quantity) VALUES
((SELECT id FROM bookings WHERE reference = 'BK-G7H8I9'), (SELECT id FROM equipment_items WHERE slug = '2-person-backpacking-tent'), 2),
((SELECT id FROM bookings WHERE reference = 'BK-G7H8I9'), (SELECT id FROM equipment_items WHERE slug = 'sleeping-bag-0c'), 2);

-- 4) RETURNED — completed rental in the past
INSERT INTO bookings (reference, customer_name, customer_phone, start_date, end_date, status, created_at, expires_at, confirmed_at, picked_up_at, returned_at) VALUES
('BK-J1K2L3', 'Dewi Lestari', '+6284444444444', '2026-08-20', '2026-08-22', 'returned', now() - interval '15 days', now() - interval '15 days' + interval '30 minutes', now() - interval '15 days' + interval '5 minutes', now() - interval '14 days', now() - interval '12 days');
INSERT INTO booking_items (booking_id, equipment_item_id, quantity)
SELECT id, (SELECT id FROM equipment_items WHERE slug = '60l-hiking-backpack'), 1 FROM bookings WHERE reference = 'BK-J1K2L3';

-- 5) CANCELLED (customer_request) — customer cancelled via WhatsApp
INSERT INTO bookings (reference, customer_name, customer_phone, start_date, end_date, status, cancel_reason, created_at, expires_at, cancelled_at) VALUES
('BK-M4N5O6', 'Joko Prasetyo', '+6285555555555', '2026-09-15', '2026-09-16', 'cancelled', 'customer_request', now() - interval '2 days', now() - interval '2 days' + interval '30 minutes', now() - interval '2 days' + interval '15 minutes');
INSERT INTO booking_items (booking_id, equipment_item_id, quantity)
SELECT id, (SELECT id FROM equipment_items WHERE slug = 'climbing-harness'), 1 FROM bookings WHERE reference = 'BK-M4N5O6';

-- 6) CANCELLED (no_show) — customer never showed up by closing time
INSERT INTO bookings (reference, customer_name, customer_phone, start_date, end_date, status, cancel_reason, created_at, expires_at, confirmed_at, cancelled_at) VALUES
('BK-P7Q8R9', 'Maya Sari', '+6286666666666', '2026-08-25', '2026-08-25', 'cancelled', 'no_show', now() - interval '9 days', now() - interval '9 days' + interval '30 minutes', now() - interval '9 days' + interval '10 minutes', now() - interval '8 days');
INSERT INTO booking_items (booking_id, equipment_item_id, quantity)
SELECT id, (SELECT id FROM equipment_items WHERE slug = 'trekking-poles-pair'), 1 FROM bookings WHERE reference = 'BK-P7Q8R9';

-- 7) EXPIRED — pending booking that was never confirmed in time
INSERT INTO bookings (reference, customer_name, customer_phone, start_date, end_date, status, created_at, expires_at) VALUES
('BK-S1T2U3', 'Agus Setiawan', '+6287777777777', '2026-09-08', '2026-09-09', 'expired', now() - interval '5 days', now() - interval '5 days' + interval '30 minutes');
INSERT INTO booking_items (booking_id, equipment_item_id, quantity)
SELECT id, (SELECT id FROM equipment_items WHERE slug = 'headlamp'), 3 FROM bookings WHERE reference = 'BK-S1T2U3';
