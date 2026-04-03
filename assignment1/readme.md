# Appointment and Doctor Microservices

## Project Overview and Purpose

This project implements a small microservice-based healthcare scheduling system with two independent services:

- Doctor Service: owns doctor data and doctor availability status.
- Appointment Service: owns appointments and booking lifecycle.

The system is intentionally split this way to demonstrate clear domain ownership and service boundaries. Appointments depend on doctor information, but they do not own that data. Instead, Appointment Service queries Doctor Service when it needs to validate doctor existence/availability.

This structure keeps each service focused, independently deployable, and easier to evolve.

## Service Responsibilities

### Doctor Service

Owns and manages:

- Doctor entity and doctor-related data.
- API endpoints for reading doctor information.
- Availability/source of truth used by other services.

It does not manage appointment booking logic.

### Appointment Service

Owns and manages:

- Appointment entity and appointment status lifecycle.
- API endpoints for creating and managing appointments.
- Validation flow that may require doctor data from Doctor Service.

It does not own doctor records and does not persist doctor state as authoritative data.

## Folder Structure and Dependency Flow

Top-level layout:

assignment1/
	appointment-service/
	doctor-service/

Each service follows a layered internal structure:

service/
	cmd/service/main.go
	internal/
		app/
		model/
		repository/
		usecase/
		transport/http/
            dto/

Layer intent:

- transport/http: HTTP handlers and request/response DTO mapping.
- usecase: business logic and orchestration.
- repository: persistence interfaces and implementations.
- model: domain entities and enums/statuses.
- app: service wiring and startup composition.

Dependency direction (important):

- handler -> usecase -> repository -> database
- model is shared by inner layers as domain types.
- outer layers depend on inner abstractions, not the other way around.

This avoids framework/network concerns leaking into domain logic.

## Inter-Service Communication

Appointment Service calls Doctor Service during appointment validation (for example, before appointment creation) to verify doctor availability/validity.

Communication style:

- Protocol: HTTP/JSON
- Caller: Appointment Service
- Callee: Doctor Service
- Typical operation: GET doctor details by doctor ID

Expected HTTP contract (conceptual):

- Request:
	- GET /doctors/{doctorId}
- Success response:
	- Status: 200 OK
	- Body:
    ```json
    {
        "id": "d-123",
        "name": "Dr. Alice",
        "specialty": "Cardiology",
        "available": true
    }
    ```

- Not found response:
	- Status: 404 Not Found
- Service failure/unavailable:
	- Status: 5xx or network timeout/connection error

The Appointment Service uses this response to decide whether booking can proceed.

## How to Run the Project Locally

Prerequisites:

Step 0: Configure .env files in each service folder

For Doctor Service (.env):

```
    DB_USER="postgres"
    DB_PASSWORD="admin"
    DB_HOST="localhost"
    DB_PORT="5432"
    DB_NAME="appointment"
    DB_TIMEZONE="Asia/Almaty"
```

For Appointment Service (.env):

```
    DB_USER="postgres"
    DB_PASSWORD="admin"
    DB_HOST="localhost"
    DB_PORT="5432"
    DB_NAME="appointment"
    DB_TIMEZONE="Asia/Almaty"
    DOCTOR_SERVICE_URL="http://localhost:8080"
```

Step 1: start Doctor Service

```powershell
cd assignment1/doctor-service
go mod tidy
go run ./cmd/doctor-service/main.go
```

Step 2: start Appointment Service

```powershell
cd assignment1/appointment-service
go mod tidy
go run ./cmd/appointment-service/main.go
```

Step 3: call endpoints

- Use curl, Postman, or any HTTP client to:
	- Query doctor endpoints in Doctor Service.
	- Create/query appointments in Appointment Service.

Notes:

- Ensure the Appointment Service is configured to call the correct Doctor Service base URL/port.
- Start Doctor Service first so dependency checks in Appointment Service can succeed.

## Why a Shared Database Was Not Used

A shared database was intentionally avoided to preserve service-level data ownership.

Why this matters:

- Each service controls its own schema and lifecycle.
- Services evolve independently without cross-service DB coupling.
- Boundaries remain explicit through APIs instead of hidden SQL-level dependencies.

Using one shared database would blur ownership and tightly couple services, making autonomous deployment and change management harder.

## Failure Scenario and Resilience Discussion

Scenario: Doctor Service is unavailable when Appointment Service tries to validate doctor data.

Current expected behavior:

- Appointment Service cannot safely verify doctor state.
- It should reject or fail the request with an error response (typically 503 Service Unavailable or equivalent error handling used by the implementation).
- The appointment is not created to avoid inconsistent data.

Where production-grade resilience patterns fit:

- Timeouts: bound how long Appointment Service waits for Doctor Service.
- Retries: retry transient failures with backoff for short-lived outages.
- Circuit breaker: stop repeated failing calls temporarily and fail fast.
- Fallback strategy: optional degraded behavior (if business rules allow), with clear consistency trade-offs.

These patterns would be implemented in the outbound client/integration layer used by Appointment Service when calling Doctor Service.
