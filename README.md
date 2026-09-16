# Hospital Middleware API

A secure, enterprise-grade middleware API service developed in Go for querying and synchronizing patient records across Hospital Information Systems (HIS) and internal databases, featuring strict multi-tenancy data isolation in accordance with Agnos Health requirements.

---

## 1. Tech Stack

- **Language:** Go 1.27 / 1.23+
- **HTTP Framework:** Gin Web Framework
- **ORM / Database Driver:** GORM + pgx (PostgreSQL driver)
- **Database:** PostgreSQL 15
- **Reverse Proxy:** Nginx (Alpine)
- **Authentication:** JWT (HMAC-SHA256) + bcrypt
- **Containerization:** Docker & Docker Compose
- **API Documentation:** Swagger / OpenAPI 2.0 (`swag`, `gin-swagger`)

---

## 2. Project Architecture (Clean Architecture)

The codebase strictly follows Clean Architecture (Layered Architecture) principles to ensure separation of concerns, high maintainability, and thorough testability:

```text
hospital-middleware-api/
├── api/
│   └── docs/                       # Swagger 2.0 / OpenAPI generated spec (docs.go, swagger.json, swagger.yaml)
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point: config loading, DB connection, DI wiring
├── config/
│   └── config.go                   # Environment variables and configuration loader
├── docs/                           # Architecture, database schema, and QA test reports (Markdown)
│   ├── architecture_and_schema.md  # Clean Architecture flows, ER model, SQL DDL, and Sequence Diagrams
│   └── test_report.md              # Complete QA test report and 100% statement coverage matrix
├── internal/
│   ├── domain/                     # Core business entities, DTOs, and interface contracts
│   │   ├── hospital.go
│   │   ├── staff.go
│   │   └── patient.go
│   ├── repository/                 # Data access layer (PostgreSQL interaction via GORM)
│   │   ├── db.go
│   │   ├── hospital_repository.go
│   │   ├── staff_repository.go
│   │   └── patient_repository.go   # Enforces WHERE hospital_id = ? on all queries
│   ├── usecase/                    # Business logic layer
│   │   ├── staff_usecase.go        # Registration, login, password hashing, JWT issuance
│   │   └── patient_usecase.go      # Patient search and external HIS synchronization
│   ├── delivery/
│   │   └── http/                   # Transport layer (Gin handlers and routing)
│   │       ├── middleware/
│   │       │   └── auth_middleware.go # Validates JWT Bearer and injects hospital_id into context
│   │       ├── staff_handler.go
│   │       ├── patient_handler.go
│   │       └── router.go
│   └── client/
│       └── his_client.go           # External integration client for Hospital A HIS API
├── pkg/
│   ├── utils/                      # Utilities (bcrypt password hashing, JWT generator and validator)
│   └── response/                   # Standardized JSON response envelope (Success, Error)
├── migrations/                     # SQL DDL migration scripts (schema and composite indexes)
├── nginx/                          # Nginx reverse proxy configuration
│   ├── nginx.conf
│   └── conf.d/default.conf
├── Dockerfile                      # Multi-stage Docker build file
├── docker-compose.yml              # Multi-container orchestration (PostgreSQL + Go API + Nginx)
└── .env.example                    # Sample environment variables configuration
```

---

## 3. Data Isolation Mechanism (Hospital-Level Multi-Tenancy)

Patient data security and hospital isolation are strictly enforced through a **3-tier defense-in-depth architecture**:

1. **Cryptographically Signed Token Claims:** When a hospital staff member authenticates (`/staff/login`), their assigned `hospital_id` and `staff_id` are embedded directly inside the JWT claims payload and signed with HMAC-SHA256.
2. **Server-Side Context Injection:** Every request to protected routes (`/patient/search`) passes through `AuthMiddleware`. The middleware verifies the token signature, extracts the validated `hospital_id`, and injects it into the Gin request context (`c.Set("hospital_id", ...)`). Clients cannot manipulate or spoof this value via query strings or request headers.
3. **Mandatory Database Query Filter:** In `patient_repository.go`, all queries unconditionally enforce `WHERE hospital_id = :authenticated_hospital_id`. Staff from Hospital A will receive an empty result set (`[]`) if querying patient data belonging to Hospital B, guaranteeing 100% isolation across tenants.

---

## 4. Installation and Setup Guide

### 4.1 Running with Docker Compose (Recommended)

Launch the entire stack (PostgreSQL, Go API, and Nginx) with a single command:

```bash
docker compose up --build -d
```

- **Nginx Reverse Proxy:** `http://localhost/` (Port 80)
- **Go Web API Direct:** `http://localhost:5000/` (Port 5000)
- **PostgreSQL Database:** `localhost:5435` (Host port mapped to 5435 to avoid port collisions with pre-existing local databases)

Check container health and status:
```bash
docker compose ps
```

---

### 4.2 Local Development Setup

1. Copy the sample environment file:
   ```bash
   cp .env.example .env
   ```
2. Adjust database credentials in `.env` if necessary.
3. Download dependencies:
   ```bash
   go mod download
   ```
4. Start the server:
   ```bash
   go run ./cmd/server/main.go
   ```

---

## 5. Running Unit Tests

The test suite covers both positive and negative test scenarios using mock repositories, mock HTTP clients, and mock HTTP servers with high statement coverage:

```bash
go test -v -cover ./...
```

### Test Suite Overview:
- `config`: Tests environment variable parsing, default fallback values, and PostgreSQL DSN string formatting.
- `internal/client`: Tests external Hospital A HIS client with positive lookup, 404 not found, and 500 server error handling using `httptest.Server`.
- `internal/delivery/http`: Tests HTTP handlers for `/staff/create`, `/staff/login`, `/patient/search`, and root/health routes.
- `internal/delivery/http/middleware`: Validates authorization header presence, bearer token parsing, expired token rejection, and context injection (100% coverage).
- `internal/usecase`:
  - `TestStaffUseCase_CreateAndLogin`: Validates staff account creation, duplicate username rejection (409), credential matching, and cross-hospital login validation.
  - `TestPatientUseCase_DataIsolationAndHIS`: Validates intra-hospital patient searches, **strictly verifies hospital-level data isolation boundaries** (National ID and Passport ID), and tests external HIS synchronization fallback.
- `pkg/response`: Validates standardized API response formatting for success and error scenarios (100% coverage).
- `pkg/utils`:
  - `TestPasswordHashing`: Verifies password hashing and invalid password rejection using bcrypt.
  - `TestJWTGenerationAndValidation`: Validates JWT token generation, claims extraction, expired token rejection, and tampering detection.

> For complete testing matrix, coverage analysis, and Data Isolation Proof of Concept, refer to [QA Test Report](docs/test_report.md).

---


## 6. API Specifications & Interactive Documentation

### 6.1 Interactive Swagger UI Documentation

The API includes embedded interactive Swagger documentation:
- **Direct URL:** `http://localhost:5000/swagger/index.html`
- **Via Nginx:** `http://localhost/swagger/index.html`
- Supports interactive testing with Bearer Token Authorization (`Authorize` button).

---

### 6.2 Health Check Endpoint

- **Endpoint:** `GET /health`
- **Response (200 OK):**
  ```json
  {
    "service": "hospital-middleware-api",
    "status": "ok",
    "uptime": "12m34s"
  }
  ```

---

### 6.3 Staff Registration (`POST /staff/create`)

- **Endpoint:** `POST /staff/create`
- **Headers:** `Content-Type: application/json`
- **Request Body:**
  ```json
  {
    "username": "nurse_somying",
    "password": "Password123!",
    "hospital": "Hospital A",
    "full_name": "Somying Raksa"
  }
  ```
- **Response (201 Created):**
  ```json
  {
    "success": true,
    "message": "Staff account created successfully",
    "data": {
      "id": "c1f7a08b-1234-4567-89ab-cdef01234567",
      "username": "nurse_somying",
      "hospital_id": "9b1deb4d-5678-4321-8765-abcdefabcdef",
      "full_name": "Somying Raksa",
      "created_at": "2026-09-15T00:00:00Z"
    }
  }
  ```

---

### 6.4 Staff Authentication (`POST /staff/login`)

- **Endpoint:** `POST /staff/login`
- **Headers:** `Content-Type: application/json`
- **Request Body:**
  ```json
  {
    "username": "nurse_somying",
    "password": "Password123!",
    "hospital": "Hospital A"
  }
  ```
- **Response (200 OK):**
  ```json
  {
    "success": true,
    "message": "Authentication successful",
    "data": {
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
      "token_type": "Bearer",
      "expires_in": 86400,
      "staff": {
        "id": "c1f7a08b-1234-4567-89ab-cdef01234567",
        "username": "nurse_somying",
        "hospital_id": "9b1deb4d-5678-4321-8765-abcdefabcdef",
        "full_name": "Somying Raksa"
      }
    }
  }
  ```

---

### 6.5 Patient Search (`GET /patient/search` or `POST /patient/search`)

- **Endpoint:** `GET /patient/search` or `POST /patient/search`
- **Headers:**
  - `Authorization: Bearer <JWT_TOKEN>`
- **Query Parameters (All fields are optional):**
  - `national_id` (e.g. `1100100111111`)
  - `passport_id` (e.g. `AA123456`)
  - `first_name` (searches Thai and English first names)
  - `middle_name` (searches Thai and English middle names)
  - `last_name` (searches Thai and English last names)
  - `date_of_birth` (format: `YYYY-MM-DD`)
  - `phone_number`
  - `email`
  - `patient_hn`
- **Response (200 OK):**
  ```json
  {
    "success": true,
    "message": "Patients retrieved successfully",
    "data": [
      {
        "id": "e4eaaaf2-0000-0000-0000-111122223333",
        "hospital_id": "9b1deb4d-5678-4321-8765-abcdefabcdef",
        "patient_hn": "HN-001234",
        "national_id": "1100100111111",
        "passport_id": "AA123456",
        "first_name_th": "สมชาย",
        "middle_name_th": "",
        "last_name_th": "ใจดี",
        "first_name_en": "Somchai",
        "middle_name_en": "",
        "last_name_en": "Jaidee",
        "date_of_birth": "1990-05-15T00:00:00Z",
        "gender": "M",
        "phone_number": "0812345678",
        "email": "somchai@example.com",
        "created_at": "2026-09-15T00:00:00Z",
        "updated_at": "2026-09-15T00:00:00Z"
      }
    ]
  }
  ```
