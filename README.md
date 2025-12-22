# CBT Mini Project ✅

Sistem Computer-Based Test (CBT) sederhana untuk institusi pendidikan — fokus pada manajemen sesi tes, soal, dan pelaporan.

---

## 🔧 Teknologi (Tech Stack)

- **Backend**: Go 1.21+, gRPC, REST Gateway
- **Database**: MySQL (GORM)
- **Frontend**: Next.js + TypeScript
- **Auth**: JWT
- **Monitoring**: Elastic APM (Elasticsearch, Kibana)
- **Deploy**: Docker & Docker Compose

---

## ✨ Fitur Utama (Singkat)

- **Autentikasi**: login/logout untuk Siswa
- **Manajemen Sesi**: buat, mulai, dan akhiri sesi tes
- **Manajemen Soal**: CRUD soal (pilihan ganda & esai)
- **Pengumpulan Jawaban**: kirim jawaban dan penilaian otomatis dasar
- **Riwayat Siswa**: rekam hasil tes dan ringkasan skor
- **Monitoring**: tracing gRPC/HTTP, APM untuk performa dan error

---

## 🚀 Quick Start

1. Clone repo dan salin environment:

```powershell
git clone <repo-url>
cd cbt-mini-project
copy .env.example .env
```

2. Jalankan infrastruktur (Docker):

```powershell
cd deployment
docker-compose up -d
```

3. Jalankan backend:

```powershell
go run main.go
```

4. Jalankan frontend (opsional):

```powershell
cd web
npm install
npm run dev
```

> Tip: Pastikan MySQL berjalan dan `DB_DSN` di `.env` sesuai.

---

## 📌 API (Student)

**Autentikasi**

- `POST /v1/auth/login` — Login (body: `{email, password}`) → Response: `{token}`
- `POST /v1/auth/logout` — Logout (butuh Authorization)

**Student** (Header: `Authorization: Bearer <token>`)

- `GET /v1/test-sessions` — Daftar sesi tersedia
- `GET /v1/test-sessions/{id}` — Detail sesi
- `GET /v1/test-sessions/{id}/questions` — Ambil soal untuk sesi
- `POST /v1/test-sessions/{id}/submit` — Kirim jawaban (body: `{answers: [{nomorUrut, jawabanDipilih}]}`)
- `GET /v1/history` — Riwayat tes (siswa)

> Semua endpoint yang membutuhkan autentikasi menggunakan header: `Authorization: Bearer <token>`.

---

## 📋 API Documentation - Student Endpoints

Berikut adalah dokumentasi lengkap untuk endpoint siswa dengan contoh request dan response JSON:

### 1. POST /v1/auth/login - Login

**Request Body:**
```json
{
  "email": "student@example.com",
  "password": "securepassword123"
}
```

**Response (Success):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "refresh_token_here",
  "user": {
    "id": 1,
    "email": "student@example.com",
    "nama": "John Doe",
    "role": "SISWA",
    "isActive": true,
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  },
  "expiresAt": "2025-12-22T12:00:00Z",
  "success": true,
  "message": "Login successful"
}
```

**Response (Error):**
```json
{
  "error": "Invalid credentials",
  "code": 401
}
```

### 2. GET /v1/test-sessions - Get Available Test Sessions

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (Success):**
```json
{
  "testSessions": [
    {
      "id": 1,
      "sessionToken": "sess_12345",
      "user": {
        "id": 1,
        "email": "student@example.com",
        "nama": "John Doe",
        "role": "SISWA",
        "isActive": true,
        "createdAt": "2025-01-01T00:00:00Z",
        "updatedAt": "2025-01-01T00:00:00Z"
      },
      "namaPeserta": "John Doe",
      "tingkat": {
        "id": 1,
        "nama": "Grade 10"
      },
      "mataPelajaran": {
        "id": 1,
        "nama": "Mathematics"
      },
      "waktuMulai": "2025-12-22T10:00:00Z",
      "waktuSelesai": "2025-12-22T11:00:00Z",
      "batasWaktu": "2025-12-22T11:00:00Z",
      "durasiMenit": 60,
      "nilaiAkhir": 0,
      "jumlahBenar": 0,
      "totalSoal": 20,
      "status": "ONGOING"
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 10,
    "totalPages": 1,
    "totalRecords": 1
  }
}
```

### 3. GET /v1/test-sessions/{id} - Get Test Session Details

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (Success):**
```json
{
  "id": 1,
  "sessionToken": "sess_12345",
  "user": {
    "id": 1,
    "email": "student@example.com",
    "nama": "John Doe",
    "role": "SISWA",
    "isActive": true,
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  },
  "namaPeserta": "John Doe",
  "tingkat": {
    "id": 1,
    "nama": "Grade 10"
  },
  "mataPelajaran": {
    "id": 1,
    "nama": "Mathematics"
  },
  "waktuMulai": "2025-12-22T10:00:00Z",
  "waktuSelesai": "2025-12-22T11:00:00Z",
  "batasWaktu": "2025-12-22T11:00:00Z",
  "durasiMenit": 60,
  "nilaiAkhir": 0,
  "jumlahBenar": 0,
  "totalSoal": 20,
  "status": "ONGOING"
}
```

### 4. GET /v1/test-sessions/{id}/questions - Get Questions for Session

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (Success):**
```json
{
  "sessionToken": "sess_12345",
  "soal": [
    {
      "id": 1,
      "nomorUrut": 1,
      "pertanyaan": "What is 2 + 2?",
      "opsiA": "3",
      "opsiB": "4",
      "opsiC": "5",
      "opsiD": "6",
      "jawabanDipilih": "JAWABAN_INVALID",
      "isAnswered": false,
      "materi": {
        "id": 1,
        "nama": "Basic Arithmetic",
        "idMataPelajaran": 1,
        "idTingkat": 1,
        "isActive": true,
        "defaultDurasiMenit": 60,
        "defaultJumlahSoal": 20
      },
      "gambar": []
    }
  ],
  "totalSoal": 20,
  "currentNomorUrut": 1,
  "dijawabCount": 0,
  "isAnsweredStatus": [false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false, false],
  "batasWaktu": "2025-12-22T11:00:00Z"
}
```

### 5. POST /v1/test-sessions/{id}/submit - Submit Answers

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "answers": [
    {
      "nomorUrut": 1,
      "jawabanDipilih": "B"
    },
    {
      "nomorUrut": 2,
      "jawabanDipilih": "A"
    }
  ]
}
```

**Response (Success):**
```json
{
  "sessionToken": "sess_12345",
  "submittedAnswers": [
    {
      "nomorUrut": 1,
      "jawabanDipilih": "B",
      "isCorrect": true,
      "dijawabPada": "2025-12-22T10:15:00Z"
    },
    {
      "nomorUrut": 2,
      "jawabanDipilih": "A",
      "isCorrect": false,
      "dijawabPada": "2025-12-22T10:20:00Z"
    }
  ],
  "totalScore": 50,
  "totalCorrect": 1,
  "totalQuestions": 2
}
```

### 6. GET /v1/history - Get Test History

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (Success):**
```json
{
  "history": [
    {
      "id": 1,
      "sessionToken": "sess_12345",
      "mataPelajaran": {
        "id": 1,
        "nama": "Mathematics"
      },
      "tingkat": {
        "id": 1,
        "nama": "Grade 10"
      },
      "waktuMulai": "2025-12-22T10:00:00Z",
      "waktuSelesai": "2025-12-22T10:45:00Z",
      "durasiPengerjaanDetik": 2700,
      "nilaiAkhir": 85,
      "jumlahBenar": 17,
      "totalSoal": 20,
      "status": "COMPLETED",
      "namaPeserta": "John Doe"
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 10,
    "totalPages": 1,
    "totalRecords": 1
  },
  "rataRataNilai": 85,
  "totalTestCompleted": 1,
  "user": {
    "id": 1,
    "email": "student@example.com",
    "nama": "John Doe",
    "role": "SISWA",
    "isActive": true,
    "createdAt": "2025-01-01T00:00:00Z",
    "updatedAt": "2025-01-01T00:00:00Z"
  }
}
```

---

## 📊 Sequence Diagrams - Student Flows

Berikut adalah sequence diagrams terpisah untuk berbagai flow interaksi siswa dalam sistem CBT:

### 1. Login Flow

```mermaid
sequenceDiagram
    participant Student
    participant Backend
    participant Database

    Student->>Backend: POST /v1/auth/login (email, password)
    Backend->>Database: Validate user credentials
    Database-->>Backend: User data & role
    Backend-->>Student: JWT token & user info
```

### 2. View Available Test Sessions Flow

```mermaid
sequenceDiagram
    participant Student
    participant Backend
    participant Database

    Student->>Backend: GET /v1/test-sessions (Authorization: Bearer <token>)
    Backend->>Database: Fetch available test sessions
    Database-->>Backend: List of test sessions
    Backend-->>Student: Test sessions list
```

### 3. Take Test Session Flow (Get Questions, Submit, Result & Discussion)

```mermaid
sequenceDiagram
    participant Student
    participant Backend
    participant Database

    %% Get Session Details
    Student->>Backend: GET /v1/test-sessions/{id} (Authorization: Bearer <token>)
    Backend->>Database: Fetch session details
    Database-->>Backend: Session details
    Backend-->>Student: Session details

    %% Get Questions for Session
    Student->>Backend: GET /v1/test-sessions/{id}/questions (Authorization: Bearer <token>)
    Backend->>Database: Fetch questions for session
    Database-->>Backend: Questions data (including discussion if available)
    Backend-->>Student: Questions list

    %% Submit Answers
    Student->>Backend: POST /v1/test-sessions/{id}/submit (answers) (Authorization: Bearer <token>)
    Backend->>Database: Save student answers & calculate score
    Database-->>Backend: Save confirmation & score
    Backend-->>Student: Submission result, score & discussion
```

### 4. View Test History Flow

```mermaid
sequenceDiagram
    participant Student
    participant Backend
    participant Database

    Student->>Backend: GET /v1/history (Authorization: Bearer <token>)
    Backend->>Database: Fetch test history
    Database-->>Backend: History data (scores, results, discussions)
    Backend-->>Student: Test history
```

---

## 📊 Sequence Diagrams - Feature Flows

Berikut adalah sequence diagrams terpisah untuk fitur-fitur utama dalam sistem CBT:

### Auth Feature

#### 1. Login Flow

```mermaid
sequenceDiagram
    participant Student
    participant Backend
    participant Database

    Student->>Backend: POST /v1/auth/login (email, password)
    Backend->>Database: Validate user credentials
    Database-->>Backend: User data & role
    Backend-->>Student: JWT token & user info
```

#### 2. Logout Flow

```mermaid
sequenceDiagram
    participant Student
    participant Backend

    Student->>Backend: POST /v1/auth/logout (Authorization: Bearer <token>)
    Backend-->>Student: Logout confirmation (invalidate token)
```

*(Note: Logout mungkin dilakukan client-side dengan menghapus token, tapi diasumsikan ada endpoint untuk invalidate.)*

### Test Session Feature

#### 1. List Available Test Sessions Flow

```mermaid
sequenceDiagram
    participant Student
    participant Backend
    participant Database

    Student->>Backend: GET /v1/test-sessions (Authorization: Bearer <token>)
    Backend->>Database: Fetch available test sessions
    Database-->>Backend: List of test sessions
    Backend-->>Student: Test sessions list
```

#### 2. Get Test Session Details Flow

```mermaid
sequenceDiagram
    participant Student
    participant Backend
    participant Database

    Student->>Backend: GET /v1/test-sessions/{session_token} (Authorization: Bearer <token>)
    Backend->>Database: Fetch session details
    Database-->>Backend: Session details
    Backend-->>Student: Session details
```

#### 3. Get Questions for Session Flow

```mermaid
sequenceDiagram
    participant Student
    participant Backend
    participant Database

    Student->>Backend: GET /v1/test-sessions/{session_token}/questions (Authorization: Bearer <token>)
    Backend->>Database: Fetch questions for session
    Database-->>Backend: Questions data
    Backend-->>Student: Questions list
```

#### 4. Submit Answers Flow

```mermaid
sequenceDiagram
    participant Student
    participant Backend
    participant Database

    Student->>Backend: POST /v1/test-sessions/{session_token}/answers (answers) (Authorization: Bearer <token>)
    Backend->>Database: Save student answers & calculate score
    Database-->>Backend: Save confirmation & score
    Backend-->>Student: Submission result
```

### History Feature

#### 1. Get Test History Flow

```mermaid
sequenceDiagram
    participant Student
    participant Backend
    participant Database

    Student->>Backend: GET /v1/history (Authorization: Bearer <token>)
    Backend->>Database: Fetch test history
    Database-->>Backend: History data
    Backend-->>Student: Test history
```

#### 2. Get Test Answer Key Flow

```mermaid
sequenceDiagram
    participant Student
    participant Backend
    participant Database

    Student->>Backend: GET /v1/history/{session_token}/detail (Authorization: Bearer <token>)
    Backend->>Database: Fetch history detail with answer key
    Database-->>Backend: Detail data including correct answers
    Backend-->>Student: History detail with answer key
```

---

## 🧰 Pengembangan & Struktur

- Entry: `main.go`
- Inisialisasi: `init/`
- Logika bisnis: `internal/` (entities, handlers, usecases)
- Frontend: `web/`
- DB migration: `databases/migration`

---

## ⚙️ Variabel Lingkungan (penting)

- DB_DSN (contoh: `root:root@tcp(localhost:3306)/cbt_test`)
- GRPC_PORT (default: `6000`)
- REST_PORT (default: `8080`)
- JWT_SECRET
- ELASTIC_APM_SERVER_URL

---

## 🙋 Kontribusi & Lisensi

- Contributions: buka issue atau PR sederhana; sertakan deskripsi singkat.
- Lisensi: MIT

---

Butuh versi lain (lebih ringkas atau lebih teknis)? Katakan preferensinya dan aku sesuaikan.
