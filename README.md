# Hospital Middleware API (`hospital-middleware-api`)

Middleware API service สำหรับค้นหาและแสดงข้อมูลคนไข้จากระบบสารสนเทศโรงพยาบาล (Hospital Information Systems: HIS) และฐานข้อมูลภายใน พร้อมกลไกความปลอดภัย Multi-tenancy Data Isolation ตามข้อกำหนดของ Agnos Health.

---

## 1. Tech Stack
- **Language:** Go 1.27 / 1.23+
- **HTTP Framework:** Gin Web Framework
- **ORM / Database Driver:** GORM + pgx (PostgreSQL driver)
- **Database:** PostgreSQL 15
- **Reverse Proxy:** Nginx (Alpine)
- **Authentication:** JWT (HMAC-SHA256) + bcrypt
- **Containerization:** Docker & Docker Compose

---

## 2. Project Architecture (Clean Architecture)

โปรเจกต์ถูกออกแบบตามหลัก Clean Architecture เพื่อแบ่งแยกความรับผิดชอบ (Separation of Concerns):

```text
hospital-middleware-api/
├── cmd/
│   └── server/
│       └── main.go                 # Entry point: โหลด config, ต่อ DB, ผูก Dependency Injection
├── config/
│   └── config.go                   # โหลด Configuration และ Environment Variables
├── internal/
│   ├── domain/                     # Entity models, DTOs, และ Interfaces
│   │   ├── hospital.go
│   │   ├── staff.go
│   │   └── patient.go
│   ├── repository/                 # Data Access Layer (ติดต่อ PostgreSQL ด้วย GORM)
│   │   ├── db.go
│   │   ├── hospital_repository.go
│   │   ├── staff_repository.go
│   │   └── patient_repository.go   # บังคับ WHERE hospital_id = ? ทุกการค้นหา
│   ├── usecase/                    # Business Logic Layer
│   │   ├── staff_usecase.go        # สมัคร/เข้าสู่ระบบ, Hash รหัสผ่าน, ออก JWT
│   │   └── patient_usecase.go      # ค้นหาคนไข้และดึงข้อมูลจาก External HIS
│   ├── delivery/
│   │   └── http/                   # HTTP Transport Layer (Gin Handlers & Routes)
│   │       ├── middleware/
│   │       │   └── auth_middleware.go # ตรวจสอบ JWT Bearer และฉีด hospital_id ลง context
│   │       ├── staff_handler.go
│   │       ├── patient_handler.go
│   │       └── router.go
│   └── client/
│       └── his_client.go           # External Integration ติดต่อ Hospital A API
├── pkg/
│   ├── utils/                      # Helper Functions (bcrypt password, JWT generator/validator)
│   └── response/                   # Standardized JSON Response (Success, Error)
├── migrations/                     # SQL DDL Scripts (schema & indexes)
├── nginx/                          # Reverse proxy configuration
│   ├── nginx.conf
│   └── conf.d/default.conf
├── Dockerfile                      # Multi-stage Docker build
├── docker-compose.yml              # PostgreSQL + Go API + Nginx setup
└── .env.example                    # ตัวอย่างค่า Configuration
```

---

## 3. Data Isolation Mechanism (Multi-tenancy by Hospital)

ความปลอดภัยในการแยกข้อมูลคนไข้ระหว่างโรงพยาบาลถูกบังคับใช้ 3 ระดับอย่างเด็ดขาด:

1. **ระดับ Token Claims:** เมื่อเจ้าหน้าที่เข้าสู่ระบบ (`/staff/login`) ค่า `hospital_id` และ `staff_id` จะถูกฝังลงใน Payload ของ JWT และเซ็นชื่อ (Sign) ด้วย Secret Key
2. **ระดับ Context Injection:** ทุก Request ที่เรียกมายัง `/patient/search` ต้องผ่าน `AuthMiddleware` ซึ่งทำการตรวจสอบ JWT และฝังค่า `hospital_id` ลงใน Context ของ Gin โดยตรง ป้องกันไม่ให้ Client ปลอมแปลงโรงพยาบาลได้
3. **ระดับ Database Query:** ใน `patient_repository.go` ทุก Query ถูกบังคับเงื่อนไข `WHERE hospital_id = ?` จาก Context เป็นเงื่อนไขหลักเสมอ ทำให้ผลลัพธ์การค้นหาถูกจำกัดเฉพาะโรงพยาบาลของตนเอง 100%

---

## 4. วิธีการติดตั้งและรันระบบ (Setup & Running)

### 4.1 รันด้วย Docker Compose (แนะนำ)

สั่งรันทั้งระบบ (PostgreSQL, Go API, Nginx) ด้วยคำสั่งเดียว:

```bash
docker-compose up -d --build
```

- **Nginx Entry Point:** `http://localhost/` (Port 80)
- **Go API Direct:** `http://localhost:8080/` (Port 8080)
- **PostgreSQL Database:** `localhost:5433` (พอร์ต Host 5433 เพื่อไม่ให้ชนกับฐานข้อมูลเดิม)

ตรวจสอบสถานะระบบ:
```bash
docker-compose ps
```

---

### 4.2 รันสำหรับ Local Development

1. คัดลอกไฟล์ `.env.example` ไปเป็น `.env`:
   ```bash
   cp .env.example .env
   ```
2. ตั้งค่าการเชื่อมต่อฐานข้อมูลใน `.env`
3. ติดตั้ง Dependencies:
   ```bash
   go mod download
   ```
4. สั่งรัน Server:
   ```bash
   go run ./cmd/server/main.go
   ```

---

## 5. การรัน Unit Tests

โปรเจกต์ประกอบด้วย Unit Tests ครอบคลุมทั้ง Positive และ Negative Test Scenarios:

```bash
go test -v ./...
```

### สรุปผลการทดสอบ:
- `pkg/utils`:
  - `TestPasswordHashing`: ทดสอบการ Hash และ Verify รหัสผ่านด้วย bcrypt
  - `TestJWTGenerationAndValidation`: ทดสอบการสร้างและตรวจสอบ Token พร้อมกรณี Token ผิดพลาด
- `internal/delivery/http/middleware`:
  - `TestAuthMiddleware`: ทดสอบ Header ถูกต้อง, ขาด Header, และ Token ปลอม
- `internal/usecase`:
  - `TestStaffUseCase_CreateAndLogin`: ทดสอบสร้างบัญชีสำเร็จ, สร้างซ้ำ (Duplicate), รหัสผ่านผิด, และต่างโรงพยาบาล
  - `TestPatientUseCase_DataIsolationAndHIS`: ทดสอบการค้นหาปกติ, **การล็อคสิทธิ์แยกโรงพยาบาล (Hospital Data Isolation)**, และการดึงข้อมูลจาก External HIS

---

## 6. API Specifications & Swagger UI

### 6.1 Interactive Swagger UI Documentation
ระบบมาพร้อมกับ Swagger UI (OpenAPI 2.0 / 3.0) ซึ่งสามารถเปิดทดสอบ API ผ่าน Web Browser ได้โดยตรง:
* **URL:** `http://localhost:8080/swagger/index.html`
* รองรับการกด **Authorize** ใส่ `Bearer <token>` เพื่อทดสอบ API `/patient/search` ได้ทันที

---

### 6.2 Health Check
- **Endpoint:** `GET /health`
- **Response:**
  ```json
  {
    "service": "hospital-middleware-api",
    "status": "ok",
    "uptime": "12m34s"
  }
  ```

---

### 6.2 สร้างบัญชีเจ้าหน้าที่ (`/staff/create`)
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
      "id": "c1f7a08b-...",
      "username": "nurse_somying",
      "hospital_id": "9b1deb4d-...",
      "full_name": "Somying Raksa",
      "created_at": "2026-09-15T00:00:00Z"
    }
  }
  ```

---

### 6.3 เข้าสู่ระบบของเจ้าหน้าที่ (`/staff/login`)
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
      "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6...",
      "staff_id": "c1f7a08b-...",
      "username": "nurse_somying",
      "hospital_id": "9b1deb4d-...",
      "hospital_name": "Hospital A"
    }
  }
  ```

---

### 6.4 ค้นหาข้อมูลคนไข้ (`/patient/search`)
- **Endpoint:** `GET /patient/search` หรือ `POST /patient/search`
- **Headers:**
  - `Authorization: Bearer <JWT_TOKEN>`
- **Query Parameters (ทุกฟิลด์เป็น Optional):**
  - `national_id` (เช่น `1100100111111`)
  - `passport_id` (เช่น `AA123456`)
  - `first_name` (ค้นหาทั้งชื่อไทยและอังกฤษ)
  - `last_name` (ค้นหาทั้งนามสกุลไทยและอังกฤษ)
  - `date_of_birth` (รูปแบบ `YYYY-MM-DD`)
  - `phone_number`
  - `email`
  - `patient_hn`
- **Response (200 OK):**
  ```json
  {
    "success": true,
    "message": "Patients retrieved successfully",
    "data": {
      "total": 1,
      "patients": [
        {
          "id": "e4eaaaf2-...",
          "hospital_id": "9b1deb4d-...",
          "patient_hn": "HN-001234",
          "national_id": "1100100111111",
          "passport_id": "",
          "first_name_th": "สมชาย",
          "middle_name_th": "",
          "last_name_th": "ใจดี",
          "first_name_en": "Somchai",
          "middle_name_en": "",
          "last_name_en": "Jaidee",
          "date_of_birth": "1990-05-15",
          "gender": "M",
          "phone_number": "0812345678",
          "email": "somchai@example.com"
        }
      ]
    }
  }
  ```
