# AP2 Assignment 3 – Medical Scheduling Platform

**Author:** Yevgeniy Averyanov

## 1. Project Overview

### What changed compared to Assignment 2

Assignment 2 delivered a two-service Medical Scheduling Platform (Doctor Service + Appointment Service) communicating exclusively over gRPC with in-memory storage.

Assignment 3 extends the system in two directions:

**Persistent storage.** Both services now connect to dedicated PostgreSQL databases. In-memory maps are fully replaced by `pgx/v5`-backed repository implementations. Schema is managed exclusively through versioned migration files using `golang-migrate` — no DDL exists anywhere in application code. Each service owns its own isolated database: `doctor` and `appointments` respectively.

**Asynchronous event-driven communication.** After every successful write operation, the responsible service publishes a domain event to a RabbitMQ fanout exchange (`ap2.events`). A new third service — the Notification Service — subscribes to all events and logs each one as a structured JSON record to stdout.

**What did NOT change:**
- Domain models (`Doctor`, `Appointment`, `Status`) are identical.
- Use-case business rules are identical.
- gRPC service contracts and generated stubs are identical.
- Clean Architecture layering and dependency direction are identical.
- The gRPC failure scenario: if Doctor Service is unreachable, Appointment Service returns a descriptive `codes.Unavailable` error.

---

## 2. Broker Choice

**Chosen broker: RabbitMQ**

RabbitMQ was chosen for the following reasons:

- **Durable queues.** RabbitMQ supports persistent queues at the protocol level. Even in this assignment where the Notification Service uses an exclusive auto-delete queue, the infrastructure is already capable of durable delivery without changing the publisher side.
- **Fanout exchange model.** The `ap2.events` fanout exchange allows any number of independent consumers (e.g., a future Audit Service or Analytics Service) to bind their own queue and receive every event without any change to the publishers.
- **Management UI.** The `rabbitmq:3-management` image provides a browser-based dashboard at `http://localhost:15672` for inspecting exchanges, queues, and message rates — useful for debugging during development and defense.
- **Production readiness.** RabbitMQ's publisher confirms and dead-letter queues make it straightforward to implement guaranteed delivery patterns (see Section 10).

NATS Core would have been simpler to set up but offers no persistence and no built-in queue durability without upgrading to JetStream.

---

## 3. Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client (grpcurl)                         │
└───────────────┬────────────────────────────┬────────────────────┘
                │ gRPC                       │ gRPC
                ▼                            ▼
   ┌────────────────────┐       ┌──────────────────────────┐
   │   Doctor Service   │◄──────│   Appointment Service    │
   │    :8080 (gRPC)    │ gRPC  │      :8081 (gRPC)        │
   │                    │       │                          │
   │  PostgreSQL: doctor│       │ PostgreSQL: appointments │
   └─────────┬──────────┘       └───────────┬──────────────┘
             │ doctors.created              │ appointments.created
             │                              │ appointments.status_updated
             └──────────────┬───────────────┘
                            │ publish
                            ▼
                  ┌─────────────────┐
                  │    RabbitMQ     │
                  │ exchange:       │
                  │  ap2.events     │
                  │  (fanout)       │
                  └────────┬────────┘
                           │ subscribe (exclusive queue)
                           ▼
              ┌────────────────────────┐
              │  Notification Service  │
              │  (no port, no DB)      │
              │  logs JSON to stdout   │
              └────────────────────────┘
```

---

## 4. Environment Variables

### Doctor Service

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL DSN for the doctor database |
| `AMQP_URL` | No | `amqp://guest:guest@localhost:5672/` | RabbitMQ connection URL |
| `PORT` | No | `8080` | gRPC server port |

Example (`.env`):
```env
DATABASE_URL="postgres://postgres:admin@localhost:5432/doctor?sslmode=disable"
AMQP_URL="amqp://guest:guest@localhost:5672/"
PORT="8080"
```

### Appointment Service

| Variable | Required | Default | Description |
|---|---|---|---|
| `DATABASE_URL` | Yes | — | PostgreSQL DSN for the appointments database |
| `AMQP_URL` | No | `amqp://guest:guest@localhost:5672/` | RabbitMQ connection URL |
| `DOCTOR_SERVICE_URL` | No | `localhost:8080` | Doctor Service gRPC address |
| `PORT` | No | `8081` | gRPC server port |

Example (`.env`):
```env
DATABASE_URL="postgres://postgres:admin@localhost:5433/appointments?sslmode=disable"
AMQP_URL="amqp://guest:guest@localhost:5672/"
DOCTOR_SERVICE_URL="localhost:8080"
PORT="8081"
```

### Notification Service

| Variable | Required | Default | Description |
|---|---|---|---|
| `AMQP_URL` | Yes | — | RabbitMQ connection URL |

Example (`.env`):
```env
AMQP_URL="amqp://guest:guest@localhost:5672/"
```

---

## 5. Infrastructure Setup

### Option A — Docker Compose (recommended)

Starts all infrastructure and all three services in a single command:

```bash
docker compose up --build
```

This brings up: two PostgreSQL instances, RabbitMQ with management UI, and all three Go services. Migrations run automatically on service startup.

To stop and remove containers:

```bash
docker compose down
```

To also remove persistent volumes (wipes all data):

```bash
docker compose down -v
```

### Option B — Manual Docker (run services locally with `go run .`)

**Step 1 — Start RabbitMQ:**
```bash
docker run -d \
  --name rabbitmq \
  -p 5672:5672 \
  -p 15672:15672 \
  rabbitmq:3-management
```

**Step 2 — Start Doctor Service PostgreSQL:**
```bash
docker run -d \
  --name postgres-doctor \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=admin \
  -e POSTGRES_DB=doctor \
  -p 5432:5432 \
  postgres:17-alpine
```

**Step 3 — Start Appointment Service PostgreSQL:**
```bash
docker run -d \
  --name postgres-appointment \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=admin \
  -e POSTGRES_DB=appointments \
  -p 5433:5432 \
  postgres:17-alpine
```

Wait ~5 seconds for PostgreSQL to be ready before starting services.

RabbitMQ management UI is available at: `http://localhost:15672` (user: `guest`, password: `guest`)

---

## 6. Migration Instructions

### Automatic (default)

Migrations run automatically on service startup, before the gRPC server begins accepting requests. If migrations are already up to date, the `no change` result is silently ignored. If a migration fails, the service exits with a non-zero code and a descriptive error message.

### Manual — using the golang-migrate CLI

Install the CLI:
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

**Apply migrations (up):**
```bash
# Doctor Service
migrate -path ./doctor-service/migrations \
        -database "postgres://postgres:admin@localhost:5432/doctor?sslmode=disable" \
        up

# Appointment Service
migrate -path ./appointment-service/migrations \
        -database "postgres://postgres:admin@localhost:5433/appointments?sslmode=disable" \
        up
```

**Roll back one step (down):**
```bash
# Doctor Service
migrate -path ./doctor-service/migrations \
        -database "postgres://postgres:admin@localhost:5432/doctor?sslmode=disable" \
        down 1

# Appointment Service
migrate -path ./appointment-service/migrations \
        -database "postgres://postgres:admin@localhost:5433/appointments?sslmode=disable" \
        down 1
```

**Check current migration version:**
```bash
migrate -path ./doctor-service/migrations \
        -database "postgres://postgres:admin@localhost:5432/doctor?sslmode=disable" \
        version
```

The down migrations correctly undo the corresponding up migrations:
- `000001_create_doctors.down.sql` — `DROP TABLE IF EXISTS doctors`
- `000001_create_appointments.down.sql` — `DROP TABLE IF EXISTS appointments`

---

## 7. Service Startup Order

The correct startup order when running services manually:

**1. Infrastructure first** — RabbitMQ and both PostgreSQL instances must be running and healthy before any Go service starts. See Section 5 for Docker commands.

**2. Doctor Service** — must start before Appointment Service, because the Appointment Service connects to the Doctor Service via gRPC at startup.

```bash
cd doctor-service
go run .
```

**3. Appointment Service:**

```bash
cd appointment-service
go run .
```

**4. Notification Service** — can be started at any point after RabbitMQ is available. It has no dependency on the Go services.

```bash
cd notification-service
go run .
```

Each service reads its configuration from the `.env` file in its own directory, or from environment variables set in the shell.

---

## 8. Event Contract

All events are published to the RabbitMQ fanout exchange `ap2.events`. Because the exchange type is fanout, the routing key is broadcast to all bound queues but is still transmitted in the delivery — this is how the Notification Service determines the event type in the `subject` field of its log output.

### `doctors.created`

Published by: **Doctor Service**  
Trigger: `CreateDoctor` RPC succeeds

```json
{
  "event_type": "doctors.created",
  "occurred_at": "2026-05-03T12:00:00Z",
  "id": "d3b5e1a2-...",
  "full_name": "Dr. Aisha Seitkali",
  "specialization": "Cardiology",
  "email": "a.seitkali@clinic.kz"
}
```

| Field | Type | Description |
|---|---|---|
| `event_type` | string | Always `"doctors.created"` |
| `occurred_at` | string (RFC3339) | Timestamp of event creation |
| `id` | string (UUID) | Newly created doctor ID |
| `full_name` | string | Doctor's full name |
| `specialization` | string | Medical specialization |
| `email` | string | Unique email address |

---

### `appointments.created`

Published by: **Appointment Service**  
Trigger: `CreateAppointment` RPC succeeds

```json
{
  "event_type": "appointments.created",
  "occurred_at": "2026-05-03T12:01:00Z",
  "id": "a7f2c3d1-...",
  "title": "Initial cardiac consultation",
  "doctor_id": "d3b5e1a2-...",
  "status": "new"
}
```

| Field | Type | Description |
|---|---|---|
| `event_type` | string | Always `"appointments.created"` |
| `occurred_at` | string (RFC3339) | Timestamp of event creation |
| `id` | string (UUID) | Newly created appointment ID |
| `title` | string | Appointment title |
| `doctor_id` | string (UUID) | Referenced doctor ID |
| `status` | string | Initial status — always `"new"` |

---

### `appointments.status_updated`

Published by: **Appointment Service**  
Trigger: `UpdateAppointmentStatus` RPC succeeds

```json
{
  "event_type": "appointments.status_updated",
  "occurred_at": "2026-05-03T12:05:00Z",
  "id": "a7f2c3d1-...",
  "old_status": "new",
  "new_status": "in_progress"
}
```

| Field | Type | Description |
|---|---|---|
| `event_type` | string | Always `"appointments.status_updated"` |
| `occurred_at` | string (RFC3339) | Timestamp of event creation |
| `id` | string (UUID) | Appointment ID |
| `old_status` | string | Previous status (`new` / `in_progress` / `done`) |
| `new_status` | string | New status (`new` / `in_progress` / `done`) |

---

## 9. Notification Service

### What it does

The Notification Service is a standalone Go binary with no gRPC server, no HTTP server, and no database. Its only responsibility is to:

1. Connect to RabbitMQ on startup (with exponential backoff: 1s, 2s, 4s, 8s — up to 5 attempts before exiting with a non-zero code).
2. Declare an exclusive, auto-delete queue and bind it to the `ap2.events` fanout exchange. This means it receives every event published by any service.
3. For each incoming message: deserialize the JSON payload, construct a structured log record, and print it to stdout as a single JSON line.
4. Acknowledge each successfully processed message (Nack on deserialization failure).
5. On SIGTERM or SIGINT: drain in-flight messages, close the broker connection, and exit with code 0.

### Log output format

Each event produces exactly one line on stdout:

```json
{
  "time": "2026-05-03T12:00:00Z",
  "subject": "doctors.created",
  "event": {
    "event_type": "doctors.created",
    "occurred_at": "2026-05-03T12:00:00Z",
    "id": "d3b5e1a2-...",
    "full_name": "Dr. Aisha Seitkali",
    "specialization": "Cardiology",
    "email": "a.seitkali@clinic.kz"
  }
}
```

| Field | Description |
|---|---|
| `time` | When the Notification Service received and processed the event (RFC3339 UTC) |
| `subject` | The RabbitMQ routing key set by the publisher, identifying the event type |
| `event` | Full deserialized payload as published by the source service |

### Verifying events during a live demo

**Step 1** — Start all infrastructure and services (Section 5 & 7). Keep the Notification Service terminal visible.

**Step 2** — Call `CreateDoctor`:
```bash
grpcurl -plaintext -d '{
  "full_name": "Dr. Aisha Seitkali",
  "specialization": "Cardiology",
  "email": "a.seitkali@clinic.kz"
}' localhost:8080 doctor.DoctorService/CreateDoctor
```
Expected Notification Service output:
```json
{"time":"2026-05-03T12:00:00Z","subject":"doctors.created","event":{"email":"a.seitkali@clinic.kz","event_type":"doctors.created","full_name":"Dr. Aisha Seitkali","id":"<uuid>","occurred_at":"2026-05-03T12:00:00Z","specialization":"Cardiology"}}
```

**Step 3** — Call `CreateAppointment` (use the doctor ID returned above):
```bash
grpcurl -plaintext -d '{
  "title": "Initial cardiac consultation",
  "description": "First visit",
  "doctor_id": "<doctor-uuid>"
}' localhost:8081 appointment.AppointmentService/CreateAppointment
```
Expected Notification Service output:
```json
{"time":"2026-05-03T12:01:00Z","subject":"appointments.created","event":{"doctor_id":"<doctor-uuid>","event_type":"appointments.created","id":"<appt-uuid>","occurred_at":"2026-05-03T12:01:00Z","status":"new","title":"Initial cardiac consultation"}}
```

**Step 4** — Call `UpdateAppointmentStatus` (use the appointment ID returned above):
```bash
grpcurl -plaintext -d '{
  "id": "<appt-uuid>",
  "status": "in_progress"
}' localhost:8081 appointment.AppointmentService/UpdateAppointmentStatus
```
Expected Notification Service output:
```json
{"time":"2026-05-03T12:05:00Z","subject":"appointments.status_updated","event":{"event_type":"appointments.status_updated","id":"<appt-uuid>","new_status":"in_progress","occurred_at":"2026-05-03T12:05:00Z","old_status":"new"}}
```

**Additional grpcurl commands:**

```bash
# Get a specific doctor
grpcurl -plaintext -d '{"id": "<doctor-uuid>"}' \
  localhost:8080 doctor.DoctorService/GetDoctor

# List all doctors
grpcurl -plaintext -d '{}' \
  localhost:8080 doctor.DoctorService/ListDoctors

# Get a specific appointment
grpcurl -plaintext -d '{"id": "<appt-uuid>"}' \
  localhost:8081 appointment.AppointmentService/GetAppointment

# List all appointments
grpcurl -plaintext -d '{}' \
  localhost:8081 appointment.AppointmentService/ListAppointments
```

---

## 10. Consistency Trade-offs

### Current behaviour (best-effort publishing)

Event publishing is fire-and-forget. The sequence for `CreateDoctor` is:

1. Write doctor to PostgreSQL — commits successfully.
2. Publish `doctors.created` to RabbitMQ — **may fail silently**.
3. Return success to the gRPC caller.

This means a broker publish failure (network blip, broker restart) causes the event to be lost. The gRPC response is not affected — the doctor is persisted correctly. The error is logged with context but no retry is attempted.

**Specific scenarios where events can be lost:**

- The process crashes between the DB commit (step 1) and the publish call (step 2).
- RabbitMQ is temporarily unavailable when the publish is attempted.
- The broker connection drops mid-publish.

If the broker is unavailable at service startup, the service falls back to a `NoopPublisher` that silently discards all events. All RPCs continue to function, but no events are delivered until the service is restarted with a healthy broker.

### How to achieve guaranteed delivery

**Outbox Pattern:** Instead of publishing directly to the broker, the service writes the event payload to an `outbox` table inside the same database transaction as the domain write. A separate background worker (or CDC tool like Debezium) reads unpublished outbox rows and publishes them to the broker, marking rows as published after a confirmed delivery. This eliminates the gap between DB commit and broker publish.

**RabbitMQ Publisher Confirms:** After calling `channel.Publish`, the publisher waits for a broker acknowledgement (`channel.Confirm` mode). If the broker does not confirm within a timeout, the publish is retried. This protects against network-level drops between the service and the broker.

**NATS JetStream:** If NATS were used, switching from Core NATS to JetStream would provide at-least-once delivery with stream persistence and consumer acknowledgements, without requiring the Outbox pattern in the application.

---

## 11. Broker Comparison: NATS vs RabbitMQ

### Difference 1 — Message persistence

**NATS Core** is entirely in-memory and stateless. A message published when no subscriber is connected is permanently lost. There is no queue buffering; the broker does not store messages between publish and consume.

**RabbitMQ** supports durable queues and persistent messages stored on disk. A message published to a durable queue survives broker restarts and is held until a consumer acknowledges it. This is critical in production systems where consumers may be temporarily unavailable.

### Difference 2 — Delivery guarantees and acknowledgements

**NATS Core** operates on a fire-and-forget model. There is no protocol-level acknowledgement from subscriber to broker. The publisher has no way to know whether the message was received and processed.

**RabbitMQ** has a first-class acknowledgement protocol. Consumers send explicit `ack` or `nack` per message. The broker re-queues unacknowledged messages when a consumer disconnects. On the publisher side, `Confirm` mode allows the application to wait for a broker-level acknowledgement before considering the publish complete.

### When to choose NATS Core

- Notifications or metrics where occasional loss is acceptable.
- Ultra-low-latency use cases where sub-millisecond overhead matters.
- Simple pub/sub between stateless services that are always online.
- Systems where operational simplicity (single binary, no dependencies) outweighs the need for durability.

### When to choose RabbitMQ

- Any use case where missing an event has business consequences (billing, audit, scheduling).
- Systems with consumers that may go offline temporarily (maintenance windows, deployments).
- Workflows requiring dead-letter queues, message TTL, or routing by topic/header.
- Teams already operating RabbitMQ for other workloads (shared infrastructure).

### What would need to change to switch brokers

Switching from RabbitMQ to NATS Core would require:

1. Replacing `github.com/rabbitmq/amqp091-go` with `github.com/nats-io/nats.go` in all three services.
2. Replacing `EventPublisher` (exchange declare + `channel.Publish`) with `nc.Publish(subject, payload)`.
3. Replacing `NotificationSubscriber` (queue declare + bind + consume) with `nc.Subscribe(subject, handler)` — one subscription per subject.
4. Removing explicit `delivery.Ack()` calls (NATS Core has no consumer acknowledgements).
5. Changing the environment variable from `AMQP_URL` to `NATS_URL`.
6. The `EventPublisher` interface and `NoopPublisher` remain unchanged — only the concrete implementation swaps.