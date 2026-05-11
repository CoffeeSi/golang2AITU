# Clinic Management System - Assignment 4

## 1. Project Overview

This assignment evolved the system from a basic microservices architecture (Assignment 3) to a high-performance, resilient system. The main changes include:

* **Caching Layer:** Integrated Redis to reduce PostgreSQL load and decrease latency.
* **Background Jobs:** Implemented an asynchronous notification system to ensure reliable delivery without blocking the main business flow.
* **Rate Limiting:** Added protection against API abuse and DoS attacks.

## 2. Cache Strategy

* **Get Doctor Profile (Cache-Aside):** The service checks Redis first. On a miss, it fetches from Postgres and populates the cache. Chosen because doctor profiles are "read-heavy" and updated infrequently.
* **Search Doctors (Cache-Aside):** Search results are cached based on query parameters to handle repeated search traffic efficiently.
* **Create Appointment (Write-Around):** Appointment list cache is invalidated on creation, but the individual appointment is not cached. This prevents stale data and ensures consistency for bookings.
* **Appointments (Cache-Aside):** Appointment data is retrieved from cache if available, otherwise fetched from the database and cached. Updates trigger cache invalidation, and a TTL prevents indefinite staleness.

## 3. Rate-limiting Algorithm

* **Algorithm:** **Sliding Window** using Redis Sorted Sets.
* **Reason:** Provides per-window accuracy without boundary issues inherent to Fixed Window. Each request is tracked by timestamp.
* **Data Structure:** Redis **Sorted Sets** (ZSET) with scores representing request timestamps. Requests older than 1 minute are removed via `ZRemRangeByScore`, new requests are added with `ZAdd`, and the total count is checked with `ZCard`.
* **Error Contract:** When the limit is exceeded, gRPC returns `codes.ResourceExhausted` and includes a descriptive `retry after Xs` interval based on the oldest request in the current sliding window.

## 4. Cache Invalidation

* **Mechanism:** Manual invalidation on Update + TTL.
* **Process:** When a doctor's profile is updated, the service explicitly calls `DEL` on the corresponding Redis key.
* **Stale-read window:** If the manual invalidation fails, a stale read can persist only until the **TTL (Time-To-Live)** expires (e.g., 5 minutes).

## 5. Job Queue Design

* **Architecture:** A **Worker Pool** of goroutines listening to a shared Go channel.
* **Buffered Channel:** The channel buffer is sized at **10x the WORKER_POOL_SIZE** (minimum 100) to accommodate bursts of events from RabbitMQ without blocking the subscriber.
  - For example: if WORKER_POOL_SIZE=3, the buffer will be 30; if WORKER_POOL_SIZE=5, the buffer will be 50; if calculated size < 100, defaults to 100.
* **Backpressure:** When the buffer is full, the RabbitMQ subscriber blocks. This ensures that we don't drop messages; instead, they stay in the RabbitMQ queue until the system has the capacity to process them.
* **Example:** With WORKER_POOL_SIZE=10 and buffer=100, up to 100 jobs can be queued before the subscriber blocks, allowing workers to sustain processing without ACK timeouts.

## 6. Idempotency

* **Derivation:** SHA-256 hash of `event_type + appointment_id + occurred_at`.
* **Storage:** Keys are stored in Redis with a value of `"done"` and a **24-hour TTL**.
* **Prevention:** The subscriber checks Redis *before* enqueueing a job. If the key exists, the duplicate is logged at **info level** with status `duplicate_dropped` and silently discarded.
* **Best-Effort on Redis Failure:** If Redis is unavailable during the idempotency check, the subscriber logs a **warn level** message and continues processing (best-effort delivery). This prevents message loss at the expense of possible duplicates.
* **Worker Verification:** The worker also checks Redis *before* calling the Mock Gateway as an additional safety layer.

## 7. Dead-letter Strategy

* **After Max Retries:** If a job fails 3 times, it is logged as a `dead_letter` in `stderr` and discarded.
* **Inspection:** Errors are captured in structured JSON format for easy filtering in log management tools (e.g., ELK or Datadog).
* **Production Improvement:** In a real-world scenario, these would be moved to a **Dead Letter Queue (DLQ)** in RabbitMQ for manual replay or triggered alerts (e.g., via Sentry or Slack).

## 7.1. Event Routing & Filtering

* **Single Topic Consumption:** The notification service subscribes to a fanout exchange and receives all events (`doctors.created`, `appointments.created`, `appointments.status_updated`).
* **Event Type Filtering:** Only events with routing key `appointments.status_updated` trigger job enqueueing. Other events are logged to stdout as audit records but not enqueued for notification delivery.
* **Routing Key Check:** Filtering occurs in `handleDelivery()` by comparing `delivery.RoutingKey` before calling `HandleAppointmentStatusUpdated()`.
* **Status Filter:** Within `HandleAppointmentStatusUpdated()`, an additional filter ensures only appointments with `newStatus == "done"` generate notification jobs.

## 8. Infrastructure Setup

Ensure Docker is running, then execute:

```bash
docker-compose up -d --build
```

## 9. Service Startup Order

Run the binaries in the following order to ensure dependencies are ready:

1. **Doctor Service:** from `doctor-service/` run `go run .`
2. **Appointment Service:** from `appointment-service/` run `go run .`
3. **Mock Gateway:** from `mock-gateway/` run `go run .`
4. **Notification Service:** from `notification-service/` run `go run .`

## 10. Cache Consistency Trade-offs

* **Redis Downtime:** The system falls back to PostgreSQL. Reads become slower, but the system remains functional.
* **Distributed Consistency:** In a Redis Cluster, replication lag could lead to **eventual consistency**, where a user might see old data for a few milliseconds after an update.

## 11. Rate-limiting Trade-offs

* **Instance Limitation:** Per-instance limiting (in-memory) fails when scaling horizontally (e.g., 5 instances with a limit of 10 results in 50 total requests).
* **Accuracy:** Local limits don't account for the total load on the shared Database.
* **Solution:** A **centralized Redis counter** ensures the limit is global across all service instances, maintaining strict API quotas.

## 12. Environment Configuration

* **REDIS_URL:** Must be in URI format (e.g., `redis://localhost:6379` or `redis://:password@host:port`). This URL is parsed using `redis.ParseURL()` to properly configure authentication, database selection, and connection pooling.
* **DATABASE_URL:** PostgreSQL connection string in format `postgres://user:password@host:port/dbname`.
* **AMQP_URL:** RabbitMQ connection string (default: `amqp://guest:guest@localhost/`).
* **CACHE_TTL_SECONDS:** TTL for cached data in seconds (default: 3600).
* **RATE_LIMIT_RPM:** Rate limit in requests per minute (default: 100).
* **GATEWAY_URL:** Mock Gateway endpoint (default: `http://localhost:8080`).
* **WORKER_POOL_SIZE:** Number of concurrent notification workers (default: 3).
* **GRPC_PORT:** gRPC port for Doctor Service (default: 50051) and Appointment Service (default: 50052).

## 13. Redis Optionality

* **Startup Behavior:** All services check Redis availability on startup with a 2-second timeout via `PING`.
* **Warning Logging:** If Redis is unavailable, services log a `[WARNING]` message but continue operation without cache functionality.
* **Doctor Service:** Continues with database-only reads; rate limiting fails gracefully (all requests allowed).
* **Appointment Service:** Continues with database-only reads and updates; cache invalidation is skipped.
* **Notification Service:** Continues with best-effort delivery. If Redis is unavailable during idempotency checks, jobs are processed anyway (possible duplicates). Logs `warn` level message indicating Redis unavailability.
* **Self-healing:** When Redis becomes available, services automatically use the cache/idempotency features without restart.

## 14. Interface Boundaries

* **Cache Interface:** Use-cases depend on `CacheClientInterface` for cache CRUD operations only.
* **Rate Limiter Interface:** Rate limiting is isolated behind `RateLimiterInterface` (`Allow` and `RetryAfter`) and used by gRPC middleware.
* **Benefit:** Separates business caching concerns from traffic-control concerns while keeping infrastructure details in the cache implementation.
