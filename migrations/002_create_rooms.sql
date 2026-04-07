-- +goose Up
CREATE TABLE rooms (
    id          UUID PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    capacity    INT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE schedules (
    id         UUID PRIMARY KEY,
    room_id    UUID        NOT NULL UNIQUE REFERENCES rooms(id) ON DELETE CASCADE,
    days       INT         NOT NULL, -- bitmask: 1=Mon, 2=Tue, 4=Wed, 8=Thu, 16=Fri, 32=Sat, 64=Sun
    start_time TIME        NOT NULL,
    end_time   TIME        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE slots (
    id         UUID PRIMARY KEY,
    room_id    UUID        NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    start_at   TIMESTAMPTZ NOT NULL,
    end_at     TIMESTAMPTZ NOT NULL,
    UNIQUE (room_id, start_at)
);

CREATE INDEX idx_slots_room_start ON slots (room_id, start_at);

-- +goose Down
DROP INDEX idx_slots_room_start;
DROP TABLE slots;
DROP TABLE schedules;
DROP TABLE rooms;
