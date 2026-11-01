# Load-test report

## Scope

The k6 scenario in `load/k6_slots_list.js` exercises the read-heavy endpoint:

```text
GET /rooms/{roomId}/slots/list?date=...
```

The setup phase obtains demo admin and user tokens, creates rooms and schedules,
and optionally warms the cache. The measured phase selects rooms and requests
their slots.

## Recorded profile

- Executor: `ramping-vus`
- Ramp-up: 30 seconds to 20 virtual users
- Plateau: 1 minute at 20 virtual users
- Ramp-down: 20 seconds
- Total runtime: approximately 1 minute 50 seconds
- Optional cache warm-up controlled by `WARM_CACHE`

Run it against the local stack with:

```bash
make up
make load-slots
```

Useful variants:

```bash
WARM_CACHE=true k6 run load/k6_slots_list.js
WARM_CACHE=false k6 run load/k6_slots_list.js
ROOM_COUNT=50 k6 run load/k6_slots_list.js
BASE_URL=http://localhost:8080 k6 run load/k6_slots_list.js
```

## Recorded result

| Metric | Value |
| --- | ---: |
| Requests | 3,452 |
| Request rate | approximately 31.16 requests/s |
| Average latency | 1.36 ms |
| p90 | 1.76 ms |
| p95 | 2.23 ms |
| **p99** | **3.88 ms** |
| Maximum | 17.31 ms |
| Failed requests | 0.00% |
| Iterations | 3,390 |
| Checks | 6,864 |
| Failed checks | 0 |

Configured thresholds passed:

- `p95 < 200 ms`
- `p99 < 400 ms`
- request error rate `< 1%`

## Interpretation

These numbers describe one local run under the profile above. They demonstrate
that the pre-generated-slot and Redis-cache design is inexpensive for this
read-heavy scenario, but they are not a production capacity result and should
not be extrapolated to different hardware, datasets, concurrency, cache state,
or network conditions. The portfolio claim is therefore the measured
`p99 = 3.88 ms` for this specific run, not a rounded “sub-3 ms” claim.
