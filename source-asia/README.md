# Source Asia – Backend Assignment

Built with Go (standard library only, no external dependencies).

---

## How to Run

```bash
cd source-asia
go run ./cmd/server
```

Server starts on port 8080 by default. To use a different port:

```bash
PORT=9090 go run ./cmd/server
```

---

## Part 1 – Rate-Limited API

### Design Decisions

| Decision | Choice |
|---|---|
| HTTP status on accept | 201 Created |
| Rate-limit window | Sliding (rolling) 60 seconds |
| Rejected counter | Cumulative lifetime count |
| Concurrency | Per-user Mutex + RWMutex on the map |

**Why 201 and not 200:** Each accepted request creates a new entry in the rate-limit log, so 201 Created is more accurate.

**Why sliding window:** A fixed window resets at set intervals. A user could fire 5 requests at second :59 and 5 more at second :00 — that is 10 requests in under a second. The sliding window prevents this by checking any rolling 60-second period.

**Why cumulative rejected count:** Per-window rejected counts reset and lose history. A cumulative count gives a better audit trail.

---

### POST /request

**Request body:**
```json
{
  "user_id": "alice",
  "payload": { "any": "json value" }
}
```

**201 Created – request accepted:**
```json
{
  "message": "request accepted",
  "user_id": "alice",
  "accepted": true
}
```

**429 Too Many Requests – rate limit hit:**
```json
{
  "error": "rate_limit_exceeded",
  "message": "you have exceeded the limit of 5 requests per 60-second window"
}
```

**400 Bad Request – missing user_id:**
```json
{
  "error": "missing_user_id",
  "message": "user_id is required and must be non-empty"
}
```

**400 Bad Request – missing payload:**
```json
{
  "error": "missing_payload",
  "message": "payload is required"
}
```

**400 Bad Request – bad JSON:**
```json
{
  "error": "invalid_json",
  "message": "request body must be valid JSON: ..."
}
```

---

### GET /stats

**200 OK:**
```json
{
  "users": [
    {
      "user_id": "alice",
      "accepted_in_window": 3,
      "rejected_total": 2
    }
  ]
}
```

- `accepted_in_window` — accepted requests in the current rolling 60-second window
- `rejected_total` — cumulative lifetime rejected requests for this user

---

### curl Examples

```bash
# Send a valid request
curl -X POST http://localhost:8080/request \
  -H "Content-Type: application/json" \
  -d '{"user_id":"alice","payload":{"action":"buy"}}'

# Trigger the rate limit — run this 6 times quickly
for i in $(seq 1 6); do
  curl -s -X POST http://localhost:8080/request \
    -H "Content-Type: application/json" \
    -d '{"user_id":"alice","payload":"test"}'
  echo ""
done

# Missing user_id — expect 400
curl -X POST http://localhost:8080/request \
  -H "Content-Type: application/json" \
  -d '{"payload":"no user"}'

# Empty user_id — expect 400
curl -X POST http://localhost:8080/request \
  -H "Content-Type: application/json" \
  -d '{"user_id":"","payload":"empty"}'

# Bad JSON — expect 400
curl -X POST http://localhost:8080/request \
  -H "Content-Type: application/json" \
  -d 'not-valid-json'

# View stats
curl http://localhost:8080/stats
```

---

### Production Limitations

| Limitation | Detail |
|---|---|
| Single instance only | The in-memory limiter is local to one process. Multiple instances would each allow 5 requests, breaking the global limit. Fix: use Redis with atomic operations. |
| Restart loses state | All counters reset on restart. Fix: persist to Redis or a database with TTL support. |
| No authentication | Any caller can use any user_id. In production, derive the user identity from a verified JWT or API key. |
| No request size limit | A very large payload could exhaust memory. Fix: wrap the handler with http.MaxBytesReader. |
| Clock dependency | The sliding window uses the local machine clock. In a multi-instance setup, small clock differences cause minor inconsistencies. Fix: use a central time source or Redis TTL. |

---

## Project Structure

```
source-asia/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handler/
│   │   └── part1.go
│   ├── models/
│   │   └── models.go
│   └── ratelimit/
│       └── limiter.go
├── go.mod
└── README.md
```

---

## AI Usage

Used Claude (Anthropic) to assist with code structure, sliding window design, and README writing.
