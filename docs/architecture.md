# Architecture and design decisions

## Layer boundaries

The service separates domain objects, use cases, repository interfaces,
PostgreSQL implementations, HTTP transport, and application wiring. Business
rules can therefore be tested without HTTP or SQL, while repository tests focus
on the actual database contract.

## Seven-day slot window

Schedules generate slots for the next seven days when they are created. A
background task extends the window daily. This makes the dominant slot-list
request a simple read, keeps `slotId` stable for clients, and works well with
caching. Generating slots on every read would reduce stored data but add
computation and complicate stable identifiers and cache behavior.

## Database-enforced booking invariant

Only one active booking may reference a slot. Migration
`003_create_bookings.sql` enforces this with a partial unique index:

```sql
CREATE UNIQUE INDEX idx_bookings_slot_active
    ON bookings (slot_id)
    WHERE status = 'active';
```

The repository maps PostgreSQL unique violations to the domain error
`ErrSlotAlreadyBooked`. This keeps HTTP error mapping independent of pgx while
leaving the concurrency guarantee in the only component that observes all
writes: PostgreSQL. The integration suite sends two inserts concurrently and
checks that one succeeds, one is rejected, and one active row remains.

## Redis cache

The service caches the read-heavy endpoints `GET /rooms/list` and
`GET /rooms/{roomId}/slots/list`. Booking creation and cancellation invalidate
the affected slot data. Caching is deliberately limited to reads with a clear
invalidator rather than applied as a transparent global layer.

## Time handling

API and database timestamps use UTC. Repository results are normalized to UTC
so process-local timezone settings do not change behavior.

## Authentication and authorization

RS256 JWTs carry the user identifier, role, and expiration. Middleware validates
the signature and applies admin/user role checks. Registration and login are
real flows; `dummyLogin` is intentionally a local demo helper and is not suitable
for production exposure.

Development RSA keys are generated rather than committed. Docker Compose stores
them in a named volume; local-process keys are ignored by Git. Tests generate
ephemeral keys in memory or temporary directories.

## Observability and lifecycle

Prometheus collectors use an explicit registry to avoid global test state.
During shutdown the service stops HTTP traffic, cancels the slot generator, and
closes Redis, migration SQL, and pgx resources.

## Test boundaries

- Domain and use-case tests exercise business rules without infrastructure.
- Handler and middleware tests exercise HTTP contracts and error mapping.
- Repository integration tests exercise migrations and SQL on PostgreSQL.
- E2E tests exercise full scenarios with PostgreSQL and Redis containers.
- The race detector runs over the infrastructure-free suite in CI.
