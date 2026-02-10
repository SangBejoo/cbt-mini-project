# CBT Full API Testing Guide

## Prerequisites
Both services must be running:
```powershell
# Terminal 1 - CBT
cd d:\Kerjaan\CBT && go run main.go

# Terminal 2 - LMS
cd d:\Kerjaan\LMS && go run main.go

# Terminal 3 - Redis
docker run -d --name redis -p 6379:6379 redis:latest
```

## Service Endpoints
| Service | REST | gRPC |
|---------|------|------|
| CBT | http://localhost:8080 | localhost:6001 |
| LMS | http://localhost:8000 | localhost:6000 |

---

## Complete Flow: Student Takes an Exam

### Step 1: Register Student in LMS
First, you need a school. Create one via LMS admin or database.

```bash
# Register student (requires school_code from an existing school)
curl -X POST http://localhost:8000/v1/auth/register/student \
  -H "Content-Type: application/json" \
  -d '{
    "school_code": "SCH001",
    "email": "student@test.com",
    "password": "password123",
    "full_name": "Test Student"
  }'
```

### Step 2: Login to LMS
```bash
curl -X POST http://localhost:8000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "student@test.com",
    "password": "password123"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "...",
  "user": {
    "id": 1,
    "email": "student@test.com",
    "name": "Test Student",
    "role": "student"
  }
}
```

### Step 3: Use Token in CBT
The JWT from LMS works in CBT (shared secret).

```bash
# Save token
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Get available subjects
curl -s http://localhost:8080/v1/subjects \
  -H "Authorization: Bearer $TOKEN"

# Get available levels
curl -s http://localhost:8080/v1/levels \
  -H "Authorization: Bearer $TOKEN"

# Get available materials/topics
curl -s http://localhost:8080/v1/materi \
  -H "Authorization: Bearer $TOKEN"
```

### Step 4: Start Test Session
```bash
curl -X POST http://localhost:8080/v1/test-sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id_materi": 1,
    "durasi_menit": 30,
    "jumlah_soal": 10
  }'
```

**Response:**
```json
{
  "session": {
    "session_token": "abc123def456...",
    "nama_peserta": "Test Student",
    "status": "ongoing",
    "durasi_menit": 30,
    "total_soal": 10,
    "waktu_mulai": "2026-02-06T16:00:00Z"
  }
}
```

### Step 5: Get Questions
```bash
# Get first question
curl -s "http://localhost:8080/v1/test-sessions/abc123def456/questions?nomor_urut=1" \
  -H "Authorization: Bearer $TOKEN"
```

**Response:**
```json
{
  "question": {
    "nomor_urut": 1,
    "question_type": "multiple_choice",
    "pertanyaan": "What is 2 + 2?",
    "opsi_a": "3",
    "opsi_b": "4",
    "opsi_c": "5",
    "opsi_d": "6"
  }
}
```

### Step 6: Submit Answer
```bash
# Multiple choice
curl -X POST http://localhost:8080/v1/test-sessions/abc123def456/answers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nomor_urut": 1,
    "jawaban_dipilih": "B"
  }'

# Drag-drop question
curl -X POST http://localhost:8080/v1/test-sessions/abc123def456/drag-drop-answers \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "nomor_urut": 5,
    "answers": [
      {"item_id": 1, "slot_id": 2},
      {"item_id": 2, "slot_id": 1}
    ]
  }'
```

### Step 7: Complete Exam
```bash
curl -X POST http://localhost:8080/v1/test-sessions/abc123def456/complete \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"
```

**Response:**
```json
{
  "session": {
    "session_token": "abc123def456",
    "status": "completed",
    "nilai_akhir": 85.0,
    "jumlah_benar": 17,
    "total_soal": 20
  }
}
```

### Step 8: View Results
```bash
curl -s http://localhost:8080/v1/test-sessions/abc123def456/result \
  -H "Authorization: Bearer $TOKEN"
```

### Step 9: View History
```bash
curl -s http://localhost:8080/v1/history \
  -H "Authorization: Bearer $TOKEN"
```

---

## Event Sync (Automatic)

When exam completes, CBT publishes to LMS:
```bash
# View CBT->LMS events
docker exec redis redis-cli XRANGE cbt_events - +
```

When LMS creates classes/subjects/users, CBT receives:
```bash
# View LMS->CBT events
docker exec redis redis-cli XRANGE lms_events - +
```

---

## Admin APIs

### Create Question (Admin)
```bash
curl -X POST http://localhost:8080/v1/questions \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id_materi": 1,
    "id_tingkat": 1,
    "lms_class_id": 1,
    "pertanyaan": "What is the capital of Indonesia?",
    "opsi_a": "Jakarta",
    "opsi_b": "Surabaya",
    "opsi_c": "Bandung",
    "opsi_d": "Medan",
    "jawaban_benar": "A",
    "pembahasan": "Jakarta is the capital city of Indonesia"
  }'
```

### Create Drag-Drop Question (Admin)
```bash
curl -X POST http://localhost:8080/v1/soal-drag-drop \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id_materi": 1,
    "id_tingkat": 1,
    "lms_class_id": 1,
    "pertanyaan": "Match the countries with their capitals",
    "drag_type": "matching",
    "items": [
      {"label": "Indonesia"},
      {"label": "Japan"},
      {"label": "Thailand"}
    ],
    "slots": [
      {"label": "Jakarta"},
      {"label": "Tokyo"},
      {"label": "Bangkok"}
    ],
    "correct_answers": [
      {"item_index": 0, "slot_index": 0},
      {"item_index": 1, "slot_index": 1},
      {"item_index": 2, "slot_index": 2}
    ]
  }'
```

---

## Quick Health Check
```bash
# CBT
curl http://localhost:8080/v1/health

# LMS
curl http://localhost:8000/v1/health
```
