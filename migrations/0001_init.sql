-- =====================================================================
-- Outdoor Equipment Rental Booking System — Database Schema
-- Target: PostgreSQL
--
-- Design assumption (see PRD): equipment is tracked as TYPES with a
-- total quantity (e.g. "4-person Tent, qty 5"), not individually
-- serialized units. Availability for a date range = total_quantity minus
-- quantity already booked (pending/confirmed/ongoing) for overlapping
-- dates. If per-unit tracking is needed later, add an `equipment_units`
-- table and change availability logic to count units instead of a sum.
-- =====================================================================

-- ---------------------------------------------------------------------
-- Employees (the only role that logs in)
-- ---------------------------------------------------------------------
CREATE TABLE employees (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(100) NOT NULL,
    phone           VARCHAR(20) NOT NULL,
    email           VARCHAR(255) UNIQUE NOT NULL,
    password_hash   TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------
-- Equipment catalog
-- ---------------------------------------------------------------------
CREATE TABLE categories (
    id      SERIAL PRIMARY KEY,
    name    VARCHAR(100) NOT NULL,
    slug    VARCHAR(120) UNIQUE NOT NULL
);

CREATE TABLE equipment_items (
    id              SERIAL PRIMARY KEY,
    category_id     INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    name            VARCHAR(150) NOT NULL,
    slug            VARCHAR(180) UNIQUE NOT NULL,
    description     TEXT,
    daily_rate      NUMERIC(12,2) NOT NULL,
    total_quantity  INTEGER NOT NULL CHECK (total_quantity >= 0),
    condition_notes TEXT,
    photo_url       TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT true,  -- soft-hide from catalog without deleting
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_equipment_items_category ON equipment_items(category_id);
CREATE INDEX idx_equipment_items_active ON equipment_items(is_active);

-- ---------------------------------------------------------------------
-- Bookings
-- ---------------------------------------------------------------------
CREATE TYPE booking_status AS ENUM (
    'pending',      -- created at checkout, soft hold active
    'confirmed',    -- employee confirmed via manual WhatsApp reply
    'ongoing',      -- items picked up
    'returned',     -- items returned, rental complete
    'cancelled',    -- see cancel_reason
    'expired'       -- pending booking not confirmed within hold window
);

CREATE TYPE cancel_reason AS ENUM (
    'customer_request',
    'no_show'
);

CREATE TABLE bookings (
    id              SERIAL PRIMARY KEY,
    reference       VARCHAR(20) UNIQUE NOT NULL,   -- e.g. 'BK-0842', app-generated
    customer_name   VARCHAR(150) NOT NULL,
    customer_phone  VARCHAR(20) NOT NULL,
    start_date      DATE NOT NULL,                 -- pickup date
    end_date        DATE NOT NULL,                 -- return date
    status          booking_status NOT NULL DEFAULT 'pending',
    cancel_reason   cancel_reason,                 -- only set when status = 'cancelled'

    -- lifecycle timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ NOT NULL,           -- created_at + 30 minutes; pending hold deadline
    confirmed_at    TIMESTAMPTZ,
    picked_up_at    TIMESTAMPTZ,
    returned_at     TIMESTAMPTZ,
    cancelled_at    TIMESTAMPTZ,

    CHECK (end_date >= start_date),
    CHECK (
        (status = 'cancelled' AND cancel_reason IS NOT NULL)
        OR (status != 'cancelled' AND cancel_reason IS NULL)
    )
);

-- Fast lookup for: customer status-check page, employee search by reference
CREATE INDEX idx_bookings_reference ON bookings(reference);
-- Fast lookup for: employee search/filter by phone number
CREATE INDEX idx_bookings_customer_phone ON bookings(customer_phone);
-- Fast lookup for: pending queue, availability overlap queries
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_date_range ON bookings(start_date, end_date);

-- ---------------------------------------------------------------------
-- Booking line items (which equipment + quantity, per booking)
-- ---------------------------------------------------------------------
CREATE TABLE booking_items (
    id                  SERIAL PRIMARY KEY,
    booking_id          INTEGER NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    equipment_item_id   INTEGER NOT NULL REFERENCES equipment_items(id),
    quantity            INTEGER NOT NULL CHECK (quantity > 0),

    UNIQUE (booking_id, equipment_item_id)  -- one row per item per booking; bump quantity instead of duplicating
);

CREATE INDEX idx_booking_items_equipment ON booking_items(equipment_item_id);
CREATE INDEX idx_booking_items_booking ON booking_items(booking_id);

-- =====================================================================
-- Availability query (the core logic used by both the customer site
-- and the employee app)
--
-- "Booked" quantity for an item in a date range = sum of quantities
-- from booking_items, joined to bookings that are in an
-- active/blocking status (pending, confirmed, ongoing) and whose
-- date range overlaps the requested range.
--
-- Overlap condition for [req_start, req_end] vs [b.start_date, b.end_date]:
--   b.start_date <= req_end AND b.end_date >= req_start
-- =====================================================================

-- Example: available quantity of a single equipment item for a date range
-- (parameterize :equipment_item_id, :req_start, :req_end in application code)
--
-- SELECT
--     ei.id,
--     ei.name,
--     ei.total_quantity,
--     ei.total_quantity - COALESCE(SUM(bi.quantity), 0) AS available_quantity
-- FROM equipment_items ei
-- LEFT JOIN booking_items bi ON bi.equipment_item_id = ei.id
-- LEFT JOIN bookings b ON b.id = bi.booking_id
--     AND b.status IN ('pending', 'confirmed', 'ongoing')
--     AND b.start_date <= :req_end
--     AND b.end_date >= :req_start
-- WHERE ei.id = :equipment_item_id
-- GROUP BY ei.id;
--
-- Run this per item (or as one grouped query across all items) to render
-- the catalog with available/unavailable flags for the customer's chosen
-- date range.

-- =====================================================================
-- Expiring pending bookings
--
-- Run on a schedule (cron / background job) — or check lazily on read:
--
-- UPDATE bookings
-- SET status = 'expired'
-- WHERE status = 'pending' AND expires_at < now();
-- =====================================================================
