-- +goose Up
CREATE TABLE bookings (
    id              UUID PRIMARY KEY,
    slot_id         UUID        NOT NULL REFERENCES slots(id) ON DELETE CASCADE,
    user_id         UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status          VARCHAR(10) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled')),
    conference_link TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_bookings_slot_active
    ON bookings (slot_id)
    WHERE status = 'active';

CREATE INDEX idx_bookings_user_id ON bookings (user_id);

-- +goose Down
DROP INDEX idx_bookings_user_id;
DROP INDEX idx_bookings_slot_active;
DROP TABLE bookings;