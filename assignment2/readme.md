# Assignment 2: gRPC Microservices

## 1) Project Overview and Purpose

This repository implements two Go gRPC services that model a small healthcare scheduling domain:

- doctor-service: source of truth for doctors.
- appointment-service: source of truth for appointments.

The key architectural goal is service separation with explicit contracts:

- doctor data lives in one service.
- appointment creation depends on doctor existence, verified over gRPC.

## 2) Service Responsibilities and Data Ownership

### doctor-service owns

- Doctor entity storage and validation.
- Doctor creation and lookup APIs.
- Doctor-related business errors (for example, duplicate email, doctor not found).

### appointment-service owns

- Appointment entity storage and status lifecycle.
- Appointment validation rules (required fields, UUID format, status transitions).
- Dependency call to doctor-service for doctor existence checks before create.

### Why this split matters

- Each service can evolve its schema independently.
- Ownership boundaries are clear: no direct writes to another service database.
- Cross-service dependency is explicit and testable through proto contracts.

## 3) Requirements

- Go 1.26+
- PostgreSQL
- protoc (Protocol Buffers compiler)
- Go plugins for protoc:
	- protoc-gen-go
	- protoc-gen-go-grpc

## 4) Install protoc and Go gRPC Plugins

### Windows (recommended)

Install protoc via Chocolatey:

```powershell
choco install protoc -y
```

Install Go plugins:

```powershell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Ensure your Go bin directory is in PATH (usually %USERPROFILE%\go\bin).

Verify:

```powershell
protoc --version
protoc-gen-go --version
protoc-gen-go-grpc --version
```

## 5) Regenerate Stubs from .proto

You can regenerate service-local stubs directly from each service folder.

### doctor-service stubs

```powershell
cd doctor-service
protoc --go_out=. --go-grpc_out=. proto/doctor.proto
```

### appointment-service stubs

```powershell
cd appointment-service
protoc --go_out=. --go-grpc_out=. proto/appointment.proto
```

This updates generated files in each service's proto folder.

## 6) Configuration

Both services read database settings from environment variables:

- DB_HOST
- DB_USER
- DB_PASSWORD
- DB_NAME
- DB_PORT
- DB_TIMEZONE

appointment-service also reads:

- DOCTOR_SERVICE_URL (default: localhost:8080)

## 7) Local Startup (Order, Ports, Step-by-Step)

Startup order is important because appointment-service depends on doctor-service.

1. Start PostgreSQL and ensure both databases are reachable.
2. Start doctor-service first (port :8080):

```powershell
cd doctor-service
go run ./cmd/doctor-service
```

3. Start appointment-service second (port :8081):

```powershell
cd appointment-service
go run ./cmd/appointment-service
```

Both services auto-migrate their models at startup.

## 8) Proto Contract and Postman Examples

Open Postman and create a new gRPC request.

General setup:

- For doctor-service use server URL: localhost:8080
- For appointment-service use server URL: localhost:8081
- In Method picker choose full RPC name (examples below)
- In Message panel paste JSON payload and click Invoke

Because reflection is enabled on both servers, Postman should discover methods automatically.

### DoctorService RPCs

### CreateDoctor

- Proto request: CreateDoctorRequest { full_name, specialization, email }
- Proto response: DoctorResponse
- Business rule: full_name and email are required; duplicate doctor email is rejected.
- Postman method: doctor_service.proto.DoctorService/CreateDoctor

Postman message:

```json
{
	"full_name": "Dr. Alice Brown",
	"specialization": "Cardiology",
	"email": "alice@example.com"
}
```

Postman response example:

```json
{
	"id": "f5ad2b8e-8a8f-4c24-8f0d-1d7a34b7846b",
	"fullName": "Dr. Alice Brown",
	"specialization": "Cardiology",
	"email": "alice@example.com"
}
```

### GetDoctor

- Proto request: GetDoctorRequest { id }
- Proto response: DoctorResponse
- Business rule: id must be provided; missing doctor returns not found.
- Postman method: doctor_service.DoctorService/GetDoctor

Postman message:

```json
{
	"id": "f5ad2b8e-8a8f-4c24-8f0d-1d7a34b7846b"
}
```

### ListDoctors

- Proto request: ListDoctorsRequest {}
- Proto response: ListDoctorsResponse { repeated DoctorResponse doctors }
- Business rule: returns all known doctors.
- Postman method: doctor_service.DoctorService/ListDoctors

Postman message:

```json
{}
```

Postman response example:

```json
{
	"doctors": [
		{
			"id": "f5ad2b8e-8a8f-4c24-8f0d-1d7a34b7846b",
			"fullName": "Dr. Alice Brown",
			"specialization": "Cardiology",
			"email": "alice@example.com"
		}
	]
}
```

### AppointmentService RPCs

### CreateAppointment

- Proto request: CreateAppointmentRequest { title, description, doctor_id }
- Proto response: AppointmentResponse
- Business rules:
	- title and doctor_id are required.
	- doctor_id must refer to an existing doctor (checked via doctor-service call).
- Postman method: appointment_service.AppointmentService/CreateAppointment

Postman message:

```json
{
	"title": "Initial consultation",
	"description": "Chest pain follow-up",
	"doctor_id": "f5ad2b8e-8a8f-4c24-8f0d-1d7a34b7846b"
}
```

Postman response example:

```json
{
	"id": "93a63f3f-1824-49fa-b91a-b0f1d4a07819",
	"title": "Initial consultation",
	"description": "Chest pain follow-up",
	"doctorId": "f5ad2b8e-8a8f-4c24-8f0d-1d7a34b7846b",
	"status": "new",
	"createdAt": "2026-04-14T12:00:00Z",
	"updatedAt": "2026-04-14T12:00:00Z"
}
```

### GetAppointment

- Proto request: GetAppointmentRequest { id }
- Proto response: AppointmentResponse
- Business rules:
	- id is required.
	- id must be a valid UUID.
	- returns not found when appointment does not exist.
- Postman method: appointment_service.AppointmentService/GetAppointment

Postman message:

```json
{
	"id": "93a63f3f-1824-49fa-b91a-b0f1d4a07819"
}
```

### ListAppointments

- Proto request: ListAppointmentsRequest {}
- Proto response: ListAppointmentsResponse { repeated AppointmentResponse appointments }
- Business rule: returns all appointments.
- Postman method: appointment_service.AppointmentService/ListAppointments

Postman message:

```json
{}
```

### UpdateAppointmentStatus

- Proto request: UpdateStatusRequest { id, status }
- Proto response: AppointmentResponse
- Business rules:
	- id and status are required.
	- id must be a valid UUID.
	- status must be one of allowed status values.
	- invalid transition from done to disallowed states is rejected.
- Postman method: appointment_service.AppointmentService/UpdateAppointmentStatus

Postman message:

```json
{
	"id": "93a63f3f-1824-49fa-b91a-b0f1d4a07819",
	"status": "done"
}
```

## 9) Inter-Service Communication and Error Propagation

When creating an appointment, appointment-service calls doctor-service:

1. AppointmentService/CreateAppointment enters use case.
2. Use case calls DoctorClient.DoctorExists(doctor_id).
3. DoctorClient sends DoctorService/GetDoctor gRPC request.
4. Result mapping:
	 - doctor-service NotFound -> DoctorDoesNotExistError.
	 - transport or unknown dependency failure -> ServiceUnavailableError.
5. Appointment gRPC handler converts domain errors to API status codes:
	 - ServiceUnavailableError -> gRPC Unavailable.
	 - DoctorDoesNotExistError -> gRPC FailedPrecondition.

This keeps external errors clear while preserving internal dependency boundaries.

## 10) Architecture Diagram and Failure Scenario

### System Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                             Healthcare Scheduling System                │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                              Client (Postman)                           │
│                                                                         │
│  gRPC Requests:                                                         │
│  - DoctorService/CreateDoctor (localhost:8080)                         │
│  - DoctorService/GetDoctor (localhost:8080)                            │
│  - DoctorService/ListDoctors (localhost:8080)                          │
│  - AppointmentService/CreateAppointment (localhost:8081)               │
│  - AppointmentService/GetAppointment (localhost:8081)                  │
│  - AppointmentService/ListAppointments (localhost:8081)                │
│  - AppointmentService/UpdateAppointmentStatus (localhost:8081)         │
└─────────────────────────────────────────────────────────────────────────┘
         │                                    │
         │ gRPC                               │ gRPC
         ▼                                    ▼
┌──────────────────────────────┐  ┌──────────────────────────────┐
│     Doctor Service           │  │   Appointment Service        │
│     :8080                    │  │     :8081                    │
├──────────────────────────────┤  ├──────────────────────────────┤
│  - CreateDoctor              │  │  - CreateAppointment         │
│  - GetDoctor                 │  │  - GetAppointment            │
│  - ListDoctors               │  │  - ListAppointments          │
│                              │  │  - UpdateAppointmentStatus   │
└──────────────────────────────┘  └──────────────────────────────┘
         │                                    │
         │ gRPC Call                          │
         │ DoctorService/GetDoctor            │
         │◄─────────────────────────────────  │
         │ Returns Doctor or NotFound         │
         ├──────────────────────────────────► │
         │                                    │
         │ PostgreSQL                         │ PostgreSQL
         ▼                                    ▼
┌──────────────────────────────┐  ┌──────────────────────────────┐
│    Doctor Database           │  │  Appointment Database        │
│  (doctor_service.doctor)     │  │(appointment_service.appt)    │
│                              │  │                              │
│  - id (UUID)                 │  │  - id (UUID)                 │
│  - full_name                 │  │  - title                     │
│  - specialization            │  │  - description               │
│  - email (unique)            │  │  - doctor_id (FK)            │
│                              │  │  - status                    │
│                              │  │  - created_at                │
│                              │  │  - updated_at                │
└──────────────────────────────┘  └──────────────────────────────┘
```



### Failure Scenario Flow: Doctor Service Unavailable

1. Client sends CreateAppointment request to appointment-service (:8081).
2. appointment-service usecase attempts to call doctor-service (:8080) to verify doctor existence.
3. doctor-service (:8080) is unavailable → gRPC connection error.
4. appointment-service classifies it as dependency failure (ServiceUnavailableError).
5. internal log entry is written: `[ERROR] doctor service dependency failure: doctor_id=... err=...`
6. client receives gRPC status code **Unavailable** with message "doctor service is temporarily unavailable".

Why Unavailable is correct:

- The request itself may be valid.
- Failure is temporary and infrastructure/dependency related.
- Retrying may succeed once doctor-service recovers.

Contrast with FailedPrecondition:

- If doctor-service **returns** NotFound → FailedPrecondition, not Unavailable
- If doctor-service is **down** → Unavailable, not FailedPrecondition


## 11) REST vs gRPC Trade-Offs

Below are concrete differences and when each approach is preferred.

1. Contract and typing
	 - gRPC: strong schema-first contracts via proto, strict generated clients/servers.
	 - REST: often flexible JSON contracts, easier to change quickly but easier to drift.
	 - Choose gRPC when strict cross-service contracts matter.

2. Performance and payload size
	 - gRPC: protobuf is compact and fast; HTTP/2 supports multiplexing efficiently.
	 - REST: JSON is human-readable but larger and slower to parse.
	 - Choose gRPC for high-throughput, low-latency internal service-to-service traffic.

3. Browser/public API compatibility
	 - gRPC: not directly browser-friendly without grpc-web proxy.
	 - REST: native browser/tooling support (curl, Postman, fetch).
	 - Choose REST for public APIs and frontend-facing endpoints.

4. Streaming support
	 - gRPC: first-class client/server/bidirectional streaming.
	 - REST: streaming is possible but less standardized and often more complex.
	 - Choose gRPC for realtime or event-like streaming workflows.

Practical rule:

- Internal microservice mesh: gRPC.
- External/public consumption and ecosystem interoperability: REST.

## 12) Example .env Files

Use these as templates.

doctor-service/.env

```env
PORT=8080
DB_USER="postgres"
DB_PASSWORD="admin"
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="doctor"
DB_TIMEZONE="Asia/Almaty"
```

appointment-service/.env

```env
PORT=8081
DB_USER="postgres"
DB_PASSWORD="admin"
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="appointment"
DB_TIMEZONE="Asia/Almaty"
DOCTOR_SERVICE_URL="localhost:8080"
```

## 13) Example Error Responses

Create appointment with non-existing doctor:

Method:

- appointment_service.proto.AppointmentService/CreateAppointment

Message:

```json
{
	"title": "Checkup",
	"description": "Routine",
	"doctor_id": "00000000-0000-0000-0000-000000000000"
}
```

Expected error:

```text
ERROR:
	Code: FailedPrecondition
	Message: doctor does not exist
```

Create appointment when doctor-service is unavailable:

```text
ERROR:
	Code: Unavailable
	Message: doctor service is temporarily unavailable
```

Get appointment with invalid UUID format:

```text
ERROR:
	Code: InvalidArgument
	Message: id must be a valid UUID
```