# Clinic Management System - Assignment 4

## 1. Project Overview

This assignment evolved the system from a basic microservices architecture (Assignment 3) to a high-performance, resilient system. The main changes include:

* **Caching Layer:** Integrated Redis to reduce PostgreSQL load and decrease latency.
* **Background Jobs:** Implemented an asynchronous notification system to ensure reliable delivery without blocking the main business flow.
* **Rate Limiting:** Added protection against API abuse and DoS attacks.

## 2. Cache Strategy

* **Get Doctor Profile (Cache-Aside):** The service checks Redis first. On a miss, it fetches from Postgres and populates the cache. Chosen because doctor profiles are "read-heavy" and updated infrequently.
* **Search Doctors (Cache-Aside):** Search results are cached based on query parameters to handle repeated search traffic efficiently.
* **Appointments (Write-Around):** Appointment data is written directly to the database. Since consistency is critical for bookings, we avoid caching these to prevent double-booking or stale status views.

## 3. Rate-limiting Algorithm

* **Algorithm:** **Fixed Window Counter**.
* **Reason:** It is computationally inexpensive and easy to implement using atomic Redis operations.
* **Data Structure:** Redis **Strings** are used with the `INCR` command and an `EXPIRE` time set to the window duration (e.g., 60s).

## 4. Cache Invalidation

* **Mechanism:** Manual invalidation on Update + TTL.
* **Process:** When a doctor's profile is updated, the service explicitly calls `DEL` on the corresponding Redis key.
* **Stale-read window:** If the manual invalidation fails, a stale read can persist only until the **TTL (Time-To-Live)** expires (e.g., 5 minutes).

## 5. Job Queue Design

* **Architecture:** A **Worker Pool** of goroutines listening to a shared Go channel.
* **Buffered Channel:** The channel is buffered with a size equal to `WORKER_POOL_SIZE` (default: 3).
* **Backpressure:** When the buffer is full, the RabbitMQ subscriber blocks. This ensures that we don't drop messages; instead, they stay in the RabbitMQ queue until the system has the capacity to process them.

## 6. Idempotency

* **Derivation:** SHA-256 hash of `event_type + appointment_id + occurred_at`.
* **Storage:** Keys are stored in Redis with a value of `"done"` and a **24-hour TTL**.
* **Prevention:** The worker checks Redis *before* calling the Mock Gateway. If the key exists, the job is silently dropped, preventing duplicate notifications.

## 7. Dead-letter Strategy

* **After Max Retries:** If a job fails 3 times, it is logged as a `dead_letter` in `stderr` and discarded.
* **Inspection:** Errors are captured in structured JSON format for easy filtering in log management tools (e.g., ELK or Datadog).
* **Production Improvement:** In a real-world scenario, these would be moved to a **Dead Letter Queue (DLQ)** in RabbitMQ for manual replay or triggered alerts (e.g., via Sentry or Slack).

## 8. Infrastructure Setup

Ensure Docker is running, then execute:

```bash
docker-compose up -d --build
```

## 9. Service Startup Order

Run the binaries in the following order to ensure dependencies are ready:

1. **Doctor Service:** `go run ./doctor-service/cmd/main.go`
2. **Appointment Service:** `go run ./appointment-service/cmd/main.go`
3. **Mock Gateway:** `go run ./mock-gateway/main.go`
4. **Notification Service:** `go run ./notification-service/cmd/main.go`

## 10. Cache Consistency Trade-offs

* **Redis Downtime:** The system falls back to PostgreSQL. Reads become slower, but the system remains functional.
* **Distributed Consistency:** In a Redis Cluster, replication lag could lead to **eventual consistency**, where a user might see old data for a few milliseconds after an update.

## 11. Rate-limiting Trade-offs

* **Instance Limitation:** Per-instance limiting (in-memory) fails when scaling horizontally (e.g., 5 instances with a limit of 10 results in 50 total requests).
* **Accuracy:** Local limits don't account for the total load on the shared Database.
* **Solution:** A **centralized Redis counter** ensures the limit is global across all service instances, maintaining strict API quotas.