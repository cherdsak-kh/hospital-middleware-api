# Hospital Middleware API: เอกสารประกอบการพัฒนาและการวางแผนระบบ

> **ผู้สมัคร:** เชิดศักดิ์ (Cherdsak Kh.)  
> **ตำแหน่ง:** Back-end Developer  
> **บริษัท:** Agnos Health Co., Ltd.  
> **GitHub Repository:** [https://github.com/cherdsak-kh/hospital-middleware-api](https://github.com/cherdsak-kh/hospital-middleware-api)  
> **Swagger UI:** `http://localhost:5000/swagger/index.html`

---

## 1. ภาพรวมระบบและสถาปัตยกรรม (System Architecture)

ระบบ **Hospital Middleware API** (`hospital-middleware-api`) พัฒนาขึ้นด้วยภาษา **Go (Golang 1.27)** ร่วมกับ **Gin Web Framework** เพื่อทำหน้าที่เป็นตัวกลาง (Middleware Gateway) ในการเชื่อมต่อกับระบบสารสนเทศโรงพยาบาล (Hospital Information Systems: HIS) ช่วยให้เจ้าหน้าที่สามารถค้นหาและแสดงข้อมูลคนไข้ได้อย่างปลอดภัย พร้อมทั้งบังคับใช้การแยกสิทธิ์ข้อมูลระดับโรงพยาบาล (Hospital Data Isolation) อย่างเข้มงวด

### 1.1 รูปแบบสถาปัตยกรรม: Clean Architecture (Layered Architecture)

ระบบถูกออกแบบตามหลักการ Clean Architecture เพื่อแยกความรับผิดชอบของแต่ละส่วนอย่างชัดเจน (Separation of Concerns) ทำให้โค้ดมีความเป็นระเบียบ ง่ายต่อการดูแลรักษา และสามารถเขียน Unit Test ได้อย่างมีประสิทธิภาพ โดยแบ่งออกเป็น 4 เลเยอร์หลัก:

1. **Delivery Layer (`internal/delivery/http`):**
   * รับและจัดการ HTTP Request ผ่าน Gin Framework
   * ตรวจสอบความถูกต้องของ Input (Validation), Binding ข้อมูล, และจัดรูปแบบ Response กลับเป็น JSON มาตรฐาน
   * จัดการความปลอดภัยผ่าน Middleware (CORS, JWT Authentication, Hospital Context Injection)
2. **UseCase Layer (`internal/usecase`):**
   * บรรจุ Business Logic หลักของระบบ โดยไม่ยึดติดกับเฟรมเวิร์กภายนอก
   * ควบคุมการสมัครเจ้าหน้าที่, การตรวจสอบสิทธิ์เข้าสู่ระบบ, และการค้นหาข้อมูลคนไข้
   * ประสานงานกับ External HIS Client เมื่อต้องการค้นหาหรือซิงค์ข้อมูลจากโรงพยาบาลภายนอก
3. **Domain Layer (`internal/domain`):**
   * กำหนด Entity หลัก, Data Transfer Objects (DTOs), และ Interface กลางของระบบ
   * เป็นข้อตกลง (Contract) ร่วมกันระหว่างทุกเลเยอร์
4. **Data Access Layer (`internal/repository` และ `internal/client`):**
   * `internal/repository`: จัดการข้อมูลใน PostgreSQL ผ่าน GORM โดยบังคับเงื่อนไข SQL WHERE สำหรับ Data Isolation ทุกครั้ง
   * `internal/client`: ทำหน้าที่เป็น HTTP Client เพื่อเชื่อมต่อกับ Hospital A API (`GET https://hospital-a.api.co.th/patient/search/{id}`)

### 1.2 แผนผังการไหลของข้อมูล (Request Flow Diagram)

```mermaid
flowchart TD
    Client([Client / Web / Mobile / ระบบภายนอก])
    
    subgraph Infrastructure [Infrastructure Layer]
        Nginx[Nginx Reverse Proxy :80]
    end

    subgraph Delivery [Delivery Layer - Gin Framework :5000]
        Router[Gin Router]
        AuthMiddleware{Auth Middleware<br/>ตรวจสอบ JWT และฉีด Hospital ID}
        StaffHandler[Staff Handler]
        PatientHandler[Patient Handler]
    end

    subgraph BusinessLogic [UseCase Layer]
        StaffUseCase[Staff UseCase]
        PatientUseCase[Patient UseCase]
    end

    subgraph DataAccess [Data Access Layer]
        StaffRepo[(Staff Repository)]
        PatientRepo[(Patient Repository)]
        HISClient[HIS External Client]
    end

    subgraph Storage [Datastores & External APIs]
        Postgres[(PostgreSQL 15)]
        HospitalAAPI[Hospital A External API]
    end

    Client -->|HTTP Request| Nginx
    Nginx -->|Proxy Pass :5000| Router
    
    Router -->|Public Routes /staff/*| StaffHandler
    Router -->|Protected Routes /patient/*| AuthMiddleware
    AuthMiddleware -->|Valid Token & Hospital ID| PatientHandler

    StaffHandler --> StaffUseCase
    PatientHandler --> PatientUseCase

    StaffUseCase --> StaffRepo
    PatientUseCase -->|WHERE hospital_id = ?| PatientRepo
    PatientUseCase -.->|ซิงค์เมื่อไม่พบใน DB| HISClient

    StaffRepo --> Postgres
    PatientRepo --> Postgres
    HISClient -.->|GET /patient/search/:id| HospitalAAPI
```

---

## 2. โครงสร้างโฟลเดอร์โปรเจกต์ (Deliverable 1a: Project Structure)

```text
hospital-middleware-api/
├── api/
│   └── docs/                        # Swagger 2.0 / OpenAPI generated spec (docs.go, swagger.json, swagger.yaml)
├── cmd/
│   └── server/
│       └── main.go                  # จุดเริ่มต้นของแอปพลิเคชันและการทำ Dependency Injection
├── config/
│   └── config.go                    # โหลดค่า Environment Variables (.env และ System Environment)
├── docs/                            # เอกสารประกอบการพัฒนา (Markdown)
│   ├── deliverables_en.md           # เอกสารฉบับภาษาอังกฤษ (สำหรับส่งตรวจ)
│   ├── deliverables_th.md           # เอกสารฉบับภาษาไทย
│   ├── architecture_and_schema.md   # รายละเอียดสถาปัตยกรรมและ Schema
│   └── test_report.md               # รายงานผลการทดสอบ QA & Test Coverage 100%
├── internal/
│   ├── domain/                      # Domain Entities, DTOs และ Interfaces
│   │   ├── hospital.go              # โมเดลและสัญญา Interface ของ Hospital
│   │   ├── staff.go                 # โมเดล Staff, DTOs สำหรับ Create และ Login
│   │   └── patient.go               # โมเดล Patient, DTOs สำหรับ Search Query และ External HIS
│   ├── repository/                  # เลเยอร์ติดต่อฐานข้อมูล (PostgreSQL / GORM)
│   │   ├── db.go                    # เริ่มต้นการเชื่อมต่อฐานข้อมูลและการทำ AutoMigrate
│   │   ├── hospital_repository.go   # ฟังก์ชันจัดการข้อมูลโรงพยาบาล
│   │   ├── staff_repository.go      # ฟังก์ชันจัดการข้อมูลเจ้าหน้าที่
│   │   └── patient_repository.go    # ค้นหาคนไข้แบบ Dynamic Query ที่บังคับ hospital_id เสมอ
│   ├── usecase/                     # เลเยอร์ Business Logic
│   │   ├── staff_usecase.go         # แฮชรหัสผ่านด้วย bcrypt และการออก JWT Token
│   │   ├── staff_usecase_test.go    # Unit Tests สำหรับระบบ Staff
│   │   ├── patient_usecase.go       # ค้นหาคนไข้และการเชื่อมต่อ External HIS
│   │   └── patient_usecase_test.go  # Unit Tests สำหรับ Data Isolation และการซิงค์ HIS
│   ├── delivery/
│   │   └── http/                    # เลเยอร์ HTTP Transport (Gin Framework)
│   │       ├── middleware/
│   │       │   ├── auth_middleware.go      # ตรวจสอบ JWT Bearer และฉีด hospital_id เข้าสู่ Context
│   │       │   └── auth_middleware_test.go # Unit Tests สำหรับ Auth Middleware
│   │       ├── staff_handler.go            # Handler สำหรับ /staff/create และ /staff/login
│   │       ├── patient_handler.go          # Handler สำหรับ /patient/search (รองรับทั้ง GET และ POST)
│   │       └── router.go                   # ตั้งค่า Routes, CORS, Health Check, และ Swagger UI
│   └── client/
│       └── his_client.go            # HTTP Client สำหรับเชื่อมต่อ Hospital A External API
├── pkg/
│   ├── utils/                       # ฟังก์ชัน Helper ส่วนกลาง
│   │   ├── password.go              # ฟังก์ชันแฮชและตรวจสอบรหัสผ่านด้วย bcrypt
│   │   ├── password_test.go         # Unit Tests สำหรับฟังก์ชัน Password
│   │   ├── jwt.go                   # ฟังก์ชันสร้างและตรวจสอบ JWT Token
│   │   └── jwt_test.go              # Unit Tests สำหรับฟังก์ชัน JWT
│   └── response/
│       └── response.go              # โครงสร้าง JSON Response มาตรฐาน
├── migrations/
│   └── 000001_init_schema.up.sql    # สคริปต์ SQL DDL สร้างตารางและ Composite Indexes
├── nginx/                           # การตั้งค่า Nginx Reverse Proxy
│   ├── nginx.conf                   # การตั้งค่า Nginx หลัก
│   └── conf.d/default.conf          # กำหนด Reverse Proxy ส่งต่อ Request ไปยัง Go App
├── Dockerfile                       # Multi-stage Docker Build (Alpine Runtime)
├── docker-compose.yml               # กำหนด Services: db (PostgreSQL 15), app, nginx
├── .env.example                     # ตัวอย่างการตั้งค่า Environment Variables
├── .gitignore                       # รายการไฟล์ที่ไม่รวมเข้าใน Git
└── README.md                        # เอกสารแนะนำการติดตั้งและคู่มือการรันระบบ
```

---

## 3. การออกแบบฐานข้อมูลและ ER Diagram (Deliverable 1c: ER Diagram)

### 3.1 แผนภาพความสัมพันธ์ของเอนทิตี (ER Diagram)

```mermaid
erDiagram
    HOSPITALS ||--o{ STAFFS : "สังกัด / จ้างงาน"
    HOSPITALS ||--o{ PATIENTS : "ลงทะเบียน / ดูแลคนไข้"

    HOSPITALS {
        uuid id PK "Primary Key"
        varchar code UK "รหัสย่อโรงพยาบาล (Unique)"
        varchar name "ชื่อโรงพยาบาล"
        timestamp created_at "เวลาที่สร้าง"
        timestamp updated_at "เวลาที่แก้ไข"
    }

    STAFFS {
        uuid id PK "Primary Key"
        uuid hospital_id FK "อ้างอิง hospitals(id)"
        varchar username UK "ชื่อผู้ใช้งาน (Unique)"
        varchar password_hash "รหัสผ่านที่แฮชด้วย bcrypt"
        varchar full_name "ชื่อ-นามสกุลเจ้าหน้าที่"
        timestamp created_at "เวลาที่สร้าง"
        timestamp updated_at "เวลาที่แก้ไข"
    }

    PATIENTS {
        uuid id PK "Primary Key"
        uuid hospital_id FK "อ้างอิง hospitals(id) - แยกสิทธิ์ระดับโรงพยาบาล"
        varchar patient_hn "เลขประจำตัวผู้ป่วย (HN)"
        varchar national_id "เลขบัตรประชาชน 13 หลัก (Indexed)"
        varchar passport_id "เลขหนังสือเดินทาง (Indexed)"
        varchar first_name_th "ชื่อ (ภาษาไทย)"
        varchar middle_name_th "ชื่อกลาง (ภาษาไทย)"
        varchar last_name_th "นามสกุล (ภาษาไทย)"
        varchar first_name_en "ชื่อ (ภาษาอังกฤษ)"
        varchar middle_name_en "ชื่อกลาง (ภาษาอังกฤษ)"
        varchar last_name_en "นามสกุล (ภาษาอังกฤษ)"
        date date_of_birth "วันเกิด (YYYY-MM-DD)"
        varchar gender "เพศ ('M', 'F')"
        varchar phone_number "เบอร์โทรศัพท์ (Indexed)"
        varchar email "อีเมล (Indexed)"
        timestamp created_at "เวลาที่สร้าง"
        timestamp updated_at "เวลาที่แก้ไข"
    }
```

### 3.2 กฎทางธุรกิจและความสัมพันธ์ระหว่างข้อมูล (Business Rules & Relationship Narrative)

โครงสร้างฐานข้อมูลบังคับใช้กฎทางธุรกิจ (Business Rules) ที่สำคัญดังนี้:

#### 1. ขอบเขตโรงพยาบาลและการแยกข้อมูล (Hospital Multi-Tenancy Boundary)
* **BR-HOSP-01 (ขอบเขตองค์กรอิสระ):** แต่ละแถวในตาราง `hospitals` คือตัวแทนขององค์กรโรงพยาบาลและเป็นขอบเขต Tenant แยกขาดจากกันโดยสิ้นเชิง
* **BR-HOSP-02 (การระบุตัวตนที่ไม่ซ้ำ):** โรงพยาบาลทุกแห่งต้องมี UUID Primary Key (`id`) และรหัสย่อประจำโรงพยาบาล (`code UK`) ที่ไม่ซ้ำกัน เช่น `"hospital-a"`
* **BR-HOSP-03 (ความถูกต้องของข้อมูลอ้างอิง):** ห้ามลบข้อมูลโรงพยาบาลหากยังมีประวัติเจ้าหน้าที่หรือคนไข้ผูกอยู่ (`ON DELETE RESTRICT`)

#### 2. การจัดการเจ้าหน้าที่และการควบคุมสิทธิ์ (Staff Management & Access Control)
* **BR-STAFF-01 (การสังกัดโรงพยาบาล):** เจ้าหน้าที่ทุกคนต้องสังกัดโรงพยาบาลเพียงแห่งเดียวเท่านั้น (`hospital_id NOT NULL REFERENCES hospitals(id)`) ห้ามมีบัญชีเจ้าหน้าที่ลอยที่ไม่มีสังกัด
* **BR-STAFF-02 (ชื่อบัญชีต้องไม่ซ้ำ):** เจ้าหน้าที่แต่ละคนต้องมี `username (UK)` ที่ไม่ซ้ำกันในทั้งระบบ หากมีการสมัครซ้ำระบบจะปฏิเสธด้วย HTTP 409 Conflict
* **BR-STAFF-03 (ความปลอดภัยของรหัสผ่าน):** รหัสผ่านต้องถูกเข้ารหัสด้วย bcrypt ก่อนบันทึกลงฐานข้อมูลเสมอ ห้ามจัดเก็บ Plaintext โดยเด็ดขาด
* **BR-STAFF-04 (ความสัมพันธ์กับโรงพยาบาล):** โรงพยาบาล 1 แห่ง มีเจ้าหน้าที่ได้ตั้งแต่ 0 ถึงหลายคน (`HOSPITALS ||--o{ STAFFS`) ขณะที่เจ้าหน้าที่ 1 คน สังกัดได้เพียง 1 โรงพยาบาลเท่านั้น

#### 3. ระเบียนคนไข้และการแยกสิทธิ์ข้อมูล (Patient Records & Data Isolation)
* **BR-PAT-01 (กรรมสิทธิ์ข้อมูลระดับโรงพยาบาล):** ข้อมูลคนไข้แต่ละรายการต้องเป็นกรรมสิทธิ์ของโรงพยาบาลเพียงแห่งเดียว (`hospital_id NOT NULL REFERENCES hospitals(id)`)
* **BR-PAT-02 (การแยกสิทธิ์ข้อมูลอย่างเด็ดขาด):** เจ้าหน้าที่จะสามารถค้นหาและดูข้อมูลได้เฉพาะคนไข้ที่สังกัดโรงพยาบาลเดียวกันกับตนเองเท่านั้น (`WHERE hospital_id = :authenticated_hospital_id`)
* **BR-PAT-03 (การระบุตัวตนคนไข้ตามสังกัด):** คนไข้แต่ละคนจะถูกระบุด้วย `patient_hn`, `national_id` (เลขบัตรประชาชน), หรือ `passport_id` (เลขพาสปอร์ต) หากบุคคลเดียวกันไปรับการรักษาที่ 2 โรงพยาบาล จะถูกจัดเก็บเป็น 2 ระเบียนแยกขาดจากกัน
* **BR-PAT-04 (การซิงค์ข้อมูลจาก External HIS):** เมื่อค้นหาคนไข้ในฐานข้อมูลท้องถิ่นไม่พบ ระบบจะส่งคำขอไปยัง External HIS (Hospital A API) หากพบข้อมูล จะทำการซิงค์และบันทึกลงฐานข้อมูลโดยผูกกับ `hospital_id` ของโรงพยาบาลผู้ค้นหาทันที
* **BR-PAT-05 (ความสัมพันธ์กับโรงพยาบาล):** โรงพยาบาล 1 แห่ง มีระเบียนคนไข้ได้ตั้งแต่ 0 ถึงหลายคน (`HOSPITALS ||--o{ PATIENTS`) ขณะที่ระเบียนคนไข้แต่ละรายการ ผูกอยู่กับ 1 โรงพยาบาลเท่านั้น

### 3.3 กลยุทธ์การสร้าง Composite Indexes เพื่อประสิทธิภาพการค้นหา

เนื่องจาก API ค้นหาคนไข้ (`/patient/search`) ทุกฟิลด์เป็นทางเลือก (Optional) การสร้าง Index แบบเดี่ยวอาจทำให้ระบบทำงานช้าลง เราจึงสร้าง **Composite Indexes ที่ขึ้นต้นด้วย `hospital_id` เสมอ** เพื่อให้ PostgreSQL กรองระดับโรงพยาบาลได้เร็วที่สุด:

1. `idx_patients_hospital_national_id ON patients(hospital_id, national_id)`
2. `idx_patients_hospital_passport_id ON patients(hospital_id, passport_id)`
3. `idx_patients_hospital_phone ON patients(hospital_id, phone_number)`
4. `idx_patients_hospital_email ON patients(hospital_id, email)`
5. `idx_patients_hospital_name_th ON patients(hospital_id, first_name_th, last_name_th)`
6. `idx_patients_hospital_name_en ON patients(hospital_id, first_name_en, last_name_en)`

---

## 4. กลไกการแยกสิทธิ์ข้อมูลระดับโรงพยาบาล (Hospital Data Isolation)

ข้อกำหนดสำคัญที่สุดของโจทย์คือ: **เจ้าหน้าที่แต่ละคนสามารถค้นหาได้เฉพาะข้อมูลคนไข้ในโรงพยาบาลเดียวกันกับตนเองเท่านั้น**

ระบบบังคับใช้ความปลอดภัยผ่าน **สถาปัตยกรรมป้องกัน 3 ชั้น (3-Tier Defense-in-Depth)**:

1. **ชั้นที่ 1 (Signed JWT Claims):**
   * เมื่อเจ้าหน้าที่ล็อกอินสำเร็จ (`/staff/login`) ระบบจะสร้าง JWT Token ที่เซ็นชื่อแบบเข้ารหัส HMAC-SHA256
   * Token Payload จะบรรจุ `hospital_id`, `staff_id`, และ `username`
   * ผู้ใช้ภายนอกไม่สามารถแก้ไขค่า `hospital_id` ได้ เพราะจะทำให้ลายเซ็น Token เสียทันที
2. **ชั้นที่ 2 (Context Injection จากเซิร์ฟเวอร์):**
   * `AuthMiddleware` ตรวจสอบความถูกต้องของ Token ก่อนส่งต่อ Request
   * สกัดค่า `hospital_id` ที่เชื่อถือได้จาก Token Claims แล้วฝังลงใน Gin Context (`c.Set("hospital_id", claims.HospitalID)`)
   * Client ไม่สามารถส่งค่า `hospital_id` มาทาง Query String เพื่อหลอกระบบได้
3. **ชั้นที่ 3 (บังคับเงื่อนไขในระดับ SQL Query):**
   * ใน `internal/repository/patient_repository.go` ทุกคำสั่งค้นหาถูกล็อคเงื่อนไขอย่างเด็ดขาด:
     ```sql
     SELECT * FROM patients WHERE hospital_id = :authenticated_hospital_id AND (...เงื่อนไขการค้นหา...);
     ```
   * **ผลลัพธ์:** หาก Staff ของ Hospital A ค้นหาข้อมูลคนไข้ที่มีอยู่ใน Hospital B ระบบจะส่งคืนผลลัพธ์เป็น Array ว่าง (`[]`) เสมอ ป้องกันการรั่วไหลของข้อมูลคนไข้ข้ามโรงพยาบาลได้ 100%

---

## 5. ข้อมูลจำเพาะของ API (Deliverable 1b: API Specifications)

ทุก Endpoint ตอบกลับด้วยโครงสร้าง JSON มาตรฐาน:
```json
{
  "success": true,
  "message": "คำอธิบายผลลัพธ์",
  "data": {},
  "error": null
}
```

### 5.1 Interactive Swagger UI
* **URL:** `http://localhost:5000/swagger/index.html` (ตรง) หรือ `http://localhost/swagger/index.html` (ผ่าน Nginx)
* รองรับการทดสอบจริงผ่าน Browser พร้อมระบบ Bearer Token Authorization

---

### 5.2 Root Endpoint (`GET /`)
* **Endpoint:** `GET /`
* **Response (200 OK):**
  ```json
  {
    "service": "hospital-middleware-api",
    "status": "running"
  }
  ```

---

### 5.3 Health Check (`GET /health`)
* **Endpoint:** `GET /health`
* **Response (200 OK):**
  ```json
  {
    "service": "hospital-middleware-api",
    "status": "ok",
    "uptime": "12m34s"
  }
  ```

* **Sequence Diagram ของ Endpoint:**
```mermaid
sequenceDiagram
    autonumber
    actor Client as Client / Monitoring
    participant Nginx as Nginx Proxy (:80)
    participant Router as Gin Engine (:5000)
    participant Handler as Health Handler

    Client->>Nginx: GET /health
    Nginx->>Router: Forward GET /health
    Router->>Handler: ส่งต่อให้ Health Handler
    Handler->>Handler: คำนวณ Uptime (time.Since)
    Handler-->>Nginx: 200 OK {"service": "hospital-middleware-api", "status": "ok", "uptime": "..."}
    Nginx-->>Client: 200 OK JSON Response
```

---

### 5.4 สร้างบัญชีเจ้าหน้าที่ (`POST /staff/create`)
* **Endpoint:** `POST /staff/create`
* **Content-Type:** `application/json`
* **Request Body:**
  ```json
  {
    "username": "nurse_somying",
    "password": "Password123!",
    "hospital": "Hospital A",
    "full_name": "Somying Raksa"
  }
  ```
* **Success Response (201 Created):**
  ```json
  {
    "success": true,
    "message": "Staff account created successfully",
    "data": {
      "id": "c1f7a08b-1234-4567-89ab-cdef01234567",
      "username": "nurse_somying",
      "hospital_id": "9b1deb4d-5678-4321-8765-abcdefabcdef",
      "full_name": "Somying Raksa",
      "created_at": "2026-09-15T08:00:00Z"
    }
  }
  ```
* **Error Responses:**
  * `400 Bad Request`: ข้อมูลไม่ครบถ้วน หรือรูปแบบไม่ถูกต้อง
  * `409 Conflict`: ชื่อผู้ใช้งาน (Username) มีอยู่ในระบบแล้ว

* **Sequence Diagram ของ Endpoint:**
```mermaid
sequenceDiagram
    autonumber
    actor Admin as Hospital Administrator
    participant Nginx as Nginx Proxy (:80)
    participant Handler as Staff Handler
    participant UseCase as Staff UseCase
    participant StaffRepo as Staff Repository
    participant HospRepo as Hospital Repository
    participant DB as PostgreSQL 15

    Admin->>Nginx: POST /staff/create {"username", "password", "hospital"}
    Nginx->>Handler: Forward Request (:5000)
    Handler->>Handler: ตรวจสอบความครบถ้วนของ JSON
    alt ข้อมูลไม่ครบถ้วน
        Handler-->>Admin: 400 Bad Request
    else ข้อมูลถูกต้อง
        Handler->>UseCase: CreateStaff(req)
        UseCase->>StaffRepo: FindByUsername(username)
        StaffRepo->>DB: SELECT * FROM staffs WHERE username = ?
        DB-->>StaffRepo: Result
        alt Username ซ้ำในระบบ
            StaffRepo-->>UseCase: พบชื่อผู้ใช้เดิม
            UseCase-->>Handler: ErrConflict (Username already exists)
            Handler-->>Admin: 409 Conflict
        else Username ไม่ซ้ำ
            UseCase->>HospRepo: FindOrCreateHospital(hospital_code)
            HospRepo->>DB: SELECT or INSERT INTO hospitals
            DB-->>HospRepo: Hospital Record (UUID)
            UseCase->>UseCase: แฮชรหัสผ่านด้วย bcrypt(cost=10)
            UseCase->>StaffRepo: Create(staff_entity)
            StaffRepo->>DB: INSERT INTO staffs (id, hospital_id, username, password_hash)
            DB-->>StaffRepo: บันทึกสำเร็จ
            StaffRepo-->>UseCase: Staff Saved
            UseCase-->>Handler: Staff Created DTO (ไม่ส่งรหัสผ่านกลับ)
            Handler-->>Admin: 201 Created JSON
        end
    end
```

---

### 5.5 เจ้าหน้าที่เข้าสู่ระบบ (`POST /staff/login`)
* **Endpoint:** `POST /staff/login`
* **Content-Type:** `application/json`
* **Request Body:**
  ```json
  {
    "username": "nurse_somying",
    "password": "Password123!",
    "hospital": "Hospital A"
  }
  ```
* **Success Response (200 OK):**
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
* **Error Responses:**
  * `401 Unauthorized`: ชื่อผู้ใช้, รหัสผ่าน, หรือโรงพยาบาลไม่ถูกต้อง

* **Sequence Diagram ของ Endpoint:**
```mermaid
sequenceDiagram
    autonumber
    actor Staff as Hospital Staff
    participant Nginx as Nginx Proxy (:80)
    participant Handler as Staff Handler
    participant UseCase as Staff UseCase
    participant StaffRepo as Staff Repository
    participant JWT as JWT Utility (HMAC-SHA256)
    participant DB as PostgreSQL 15

    Staff->>Nginx: POST /staff/login {"username", "password", "hospital"}
    Nginx->>Handler: Forward Request (:5000)
    Handler->>Handler: ตรวจสอบความครบถ้วนของ JSON
    Handler->>UseCase: LoginStaff(req)
    UseCase->>StaffRepo: FindByUsernameAndHospital(username, hospital_code)
    StaffRepo->>DB: SELECT s.* FROM staffs s JOIN hospitals h ON s.hospital_id = h.id WHERE s.username = ? AND h.code = ?
    DB-->>StaffRepo: Staff Record และ Hospital
    alt ไม่พบข้อมูลเจ้าหน้าที่หรือโรงพยาบาล
        StaffRepo-->>UseCase: ไม่พบข้อมูล
        UseCase-->>Handler: ErrUnauthorized
        Handler-->>Staff: 401 Unauthorized (Invalid credentials)
    else พบข้อมูลเจ้าหน้าที่
        UseCase->>UseCase: bcrypt.CompareHashAndPassword(hash, password)
        alt รหัสผ่านไม่ถูกต้อง
            UseCase-->>Handler: ErrUnauthorized
            Handler-->>Staff: 401 Unauthorized (Invalid credentials)
        else รหัสผ่านถูกต้อง
            UseCase->>JWT: GenerateToken(staff_id, username, hospital_id)
            Note over JWT: ฝัง Claims และเซ็นชื่อด้วย HMAC-SHA256 Secret
            JWT-->>UseCase: Signed JWT Token String
            UseCase-->>Handler: ข้อมูล Login และ Token
            Handler-->>Staff: 200 OK (JWT Token และข้อมูล Staff)
        end
    end
```

---

### 5.6 ค้นหาข้อมูลคนไข้ (`GET /patient/search` และ `POST /patient/search`)
* **Endpoint:** `GET /patient/search` หรือ `POST /patient/search`
* **Authentication:** ต้องแนบ `Authorization: Bearer <token>`
* **Parameters (ทุกฟิลด์เป็นทางเลือก - Optional):**
  * `national_id` (string): เลขบัตรประจำตัวประชาชน
  * `passport_id` (string): เลขหนังสือเดินทาง
  * `first_name` (string): ชื่อ (ค้นหาได้ทั้งภาษาไทยและอังกฤษ)
  * `middle_name` (string): ชื่อกลาง
  * `last_name` (string): นามสกุล
  * `date_of_birth` (string): วันเกิดในรูปแบบ YYYY-MM-DD
  * `phone_number` (string): หมายเลขโทรศัพท์
  * `email` (string): ที่อยู่อีเมล
* **Success Response (200 OK):**
  ```json
  {
    "success": true,
    "message": "Patients retrieved successfully",
    "data": [
      {
        "id": "e4d3c2b1-0000-0000-0000-111122223333",
        "hospital_id": "9b1deb4d-5678-4321-8765-abcdefabcdef",
        "patient_hn": "HN1001",
        "national_id": "1100100100101",
        "passport_id": "AA1234567",
        "first_name_th": "สมชาย",
        "middle_name_th": "",
        "last_name_th": "ใจดี",
        "first_name_en": "Somchai",
        "middle_name_en": "",
        "last_name_en": "Jaidee",
        "date_of_birth": "1990-01-15T00:00:00Z",
        "gender": "M",
        "phone_number": "0812345678",
        "email": "somchai.j@example.com",
        "created_at": "2026-09-15T08:00:00Z",
        "updated_at": "2026-09-15T08:00:00Z"
      }
    ]
  }
  ```
* **Error Responses:**
  * `401 Unauthorized`: ไม่ได้ส่ง Token, Token หมดอายุ, หรือ Token ไม่ถูกต้อง

* **Sequence Diagram ของ Endpoint:**
```mermaid
sequenceDiagram
    autonumber
    actor Staff as Hospital Staff (Authenticated)
    participant Nginx as Nginx Proxy (:80)
    participant Auth as Auth Middleware
    participant Handler as Patient Handler
    participant UseCase as Patient UseCase
    participant Repo as Patient Repository
    participant HIS as Hospital A External API
    participant DB as PostgreSQL 15

    Staff->>Nginx: GET /patient/search?national_id=... (Header: Authorization: Bearer token)
    Nginx->>Auth: ส่งต่อให้ Protected Route (:5000)
    
    rect rgb(240, 248, 255)
    Note over Auth: ขั้นตอนที่ 1: ตรวจสอบสิทธิ์และฉีด Context
    Auth->>Auth: ตรวจสอบลายเซ็น HMAC-SHA256 ของ Token
    alt ไม่มี Token หรือ Token ปลอม
        Auth-->>Staff: 401 Unauthorized
    else Token ถูกต้อง
        Auth->>Auth: ดึงค่า hospital_id จาก Claims
        Auth->>Handler: ฉีด hospital_id เข้าสู่ Gin Context (c.Set)
    end
    end

    Handler->>Handler: แยกค่าตัวกรองการค้นหา (Query Parameters)
    Handler->>UseCase: SearchPatients(ctx, hospital_id, filters)

    rect rgb(245, 255, 245)
    Note over UseCase,DB: ขั้นตอนที่ 2: ค้นหาในฐานข้อมูลโดยบังคับสิทธิ์โรงพยาบาล
    UseCase->>Repo: Search(hospital_id, filters)
    Repo->>DB: SELECT * FROM patients WHERE hospital_id = :auth_hospital_id AND (national_id = ? OR ...)
    DB-->>Repo: ผลลัพธ์จากการ Query
    end

    alt พบข้อมูลคนไข้ในฐานข้อมูลภายใน
        Repo-->>UseCase: ส่งคืนรายการคนไข้ในสังกัด
        UseCase-->>Handler: ข้อมูลคนไข้
        Handler-->>Staff: 200 OK (รายการข้อมูลคนไข้)
    else ไม่พบข้อมูลในระบบ (เข้าสู่เงื่อนไข External HIS Sync)
        Repo-->>UseCase: ส่งคืน Array ว่าง []
        
        rect rgb(255, 250, 240)
        Note over UseCase,HIS: ขั้นตอนที่ 3: ดึงข้อมูลจาก External HIS อัตโนมัติ
        UseCase->>HIS: GET https://hospital-a.api.co.th/patient/search/{id}
        alt พบข้อมูลใน Hospital A
            HIS-->>UseCase: 200 OK (ข้อมูลคนไข้ JSON)
            UseCase->>Repo: CreatePatient(ผูกกับ hospital_id ของผู้ค้นหา)
            Repo->>DB: INSERT INTO patients (id, hospital_id, national_id, hn, ...)
            DB-->>Repo: บันทึกสำเร็จ
            Repo-->>UseCase: ข้อมูลคนไข้ที่ซิงค์แล้ว
            UseCase-->>Handler: รายการคนไข้
            Handler-->>Staff: 200 OK (ข้อมูลคนไข้ที่ซิงค์จาก HIS)
        else ไม่พบใน HIS หรือเชื่อมต่อไม่ได้
            HIS-->>UseCase: 404 Not Found
            UseCase-->>Handler: Array ว่าง []
            Handler-->>Staff: 200 OK (ผลลัพธ์ว่าง: [])
        end
        end
    end
```

---

## 6. การทดสอบและการรับประกันคุณภาพ (Task 5: Unit Tests)

ชุดทดสอบของระบบถูกเขียนขึ้นเพื่อครอบคลุมทั้ง Positive และ Negative Scenarios โดยใช้ Mock ทำให้รันได้รวดเร็วโดยไม่ต้องพึ่งพาฐานข้อมูลภายนอก:

```bash
go test -v ./...
```

* **ผลการทดสอบ Unit Tests:**
  * `pkg/utils/password_test.go`: **PASS** (ทดสอบแฮชรหัสผ่านและการปฏิเสธรหัสผ่านผิด)
  * `pkg/utils/jwt_test.go`: **PASS** (ทดสอบการสร้าง Token และการตรวจสอบ Claims)
  * `internal/delivery/http/middleware/auth_middleware_test.go`: **PASS** (ทดสอบกรณีไม่มี Token, Token ผิดรูปแบบ, และดึง hospital_id ถูกต้อง)
  * `internal/usecase/staff_usecase_test.go`: **PASS** (ทดสอบสร้าง Staff, Login, และปฏิเสธ Username ซ้ำ)
  * `internal/usecase/patient_usecase_test.go`: **PASS** (ทดสอบค้นหาคนไข้ในโรงพยาบาล, **พิสูจน์การบล็อกข้อมูลข้ามโรงพยาบาล**, และการซิงค์ข้อมูลจาก Hospital A)
* **ผลสรุป:** ทุกชุดการทดสอบผ่าน 100%

---

## 7. การติดตั้งและรันด้วย Docker Compose (Deliverable 2)

ระบบถูกจัดเตรียมเป็น Containers พร้อมใช้งาน 3 ตัว:
1. **`db`:** PostgreSQL 15 กำหนดแมปพอร์ตโฮสต์เป็น `5435` (เพื่อไม่ให้ชนกับฐานข้อมูลเดิมของเครื่อง) พร้อมสร้างตารางอัตโนมัติและเก็บข้อมูลแบบ Persistent
2. **`hospital-middleware-api`:** Go Binary ที่คอมไพล์ผ่าน Multi-stage Dockerfile รันบนพอร์ต `5000`
3. **`nginx`:** Reverse Proxy พอร์ต `80` จัดการ Routing ส่งต่อไปยัง Go App

### คำสั่งเริ่มระบบในคำสั่งเดียว:
```bash
docker compose up --build -d
```
ระบบทุกตัวจะสตาร์ทขึ้นมาและพร้อมทำงานร่วมกันได้ทันที
