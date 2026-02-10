# CBT ↔ LMS Event Streaming API Documentation

## Overview
CBT and LMS communicate via Redis Streams for real-time event synchronization.

| Stream | Direction | Purpose |
|--------|-----------|---------|
| `lms_events` | LMS → CBT | Sync classes, subjects, levels, modules, users |
| `cbt_events` | CBT → LMS | Sync exam results back to gradebook |

---

## LMS → CBT Events (`lms_events` stream)

### 1. `class_upsert` - Sync Class Data
Published when a class is created or updated in LMS.

```json
{
  "event": "class_upsert",
  "payload": {
    "id": 1,
    "school_id": 100,
    "name": "Class 10A",
    "is_active": true
  }
}
```

**Test via Redis CLI:**
```bash
docker exec redis redis-cli XADD lms_events "*" event class_upsert payload '{"id":1,"school_id":100,"name":"Class 10A","is_active":true}'
```

---

### 2. `class_deleted` - Delete Class
Published when a class is deleted in LMS.

```json
{
  "event": "class_deleted",
  "payload": {
    "id": 1
  }
}
```

**Test via Redis CLI:**
```bash
docker exec redis redis-cli XADD lms_events "*" event class_deleted payload '{"id":1}'
```

---

### 3. `subject_upsert` - Sync Subject (Mata Pelajaran)
```json
{
  "event": "subject_upsert",
  "payload": {
    "id": 5,
    "name": "Mathematics"
  }
}
```

---

### 4. `level_upsert` - Sync Level (Tingkat)
```json
{
  "event": "level_upsert",
  "payload": {
    "id": 10,
    "name": "Grade 10"
  }
}
```

---

### 5. `module_upsert` - Sync Module (Materi)
```json
{
  "event": "module_upsert",
  "payload": {
    "id": 15,
    "subject_id": 5,
    "level_id": 10,
    "name": "Algebra Basics"
  }
}
```

---

### 6. `user_upsert` - Sync User
```json
{
  "event": "user_upsert",
  "payload": {
    "id": 42,
    "email": "student@school.com",
    "name": "John Doe",
    "role": "siswa"
  }
}
```

---

### 7. `exam_assignment_created` - Create Test Session from Assignment
```json
{
  "event": "exam_assignment_created",
  "payload": {
    "id": 100,
    "user_id": 42,
    "module_id": 15,
    "scheduled_time": "2026-02-07T09:00:00Z",
    "duration_mins": 60,
    "question_count": 20
  }
}
```

---

## CBT → LMS Events (`cbt_events` stream)

### 1. `exam_result_completed` - Submit Exam Results
Published when a student completes an exam in CBT.

```json
{
  "event": "exam_result_completed",
  "payload": {
    "assignment_id": 100,
    "user_id": 42,
    "lms_class_id": 1,
    "score": 85.5,
    "correct_count": 17,
    "total_count": 20,
    "completed_at": "2026-02-06T10:45:00Z"
  }
}
```

**Test via Redis CLI:**
```bash
docker exec redis redis-cli XADD cbt_events "*" event exam_result_completed payload '{"assignment_id":100,"user_id":42,"lms_class_id":1,"score":85.5,"correct_count":17,"total_count":20,"completed_at":"2026-02-06T10:45:00Z"}'
```

---

## Testing Commands

### View all events in a stream:
```bash
# View LMS events
docker exec redis redis-cli XRANGE lms_events - +

# View CBT events
docker exec redis redis-cli XRANGE cbt_events - +
```

### Clear a stream:
```bash
docker exec redis redis-cli DEL lms_events
docker exec redis redis-cli DEL cbt_events
```

---

## Service Ports
| Service | gRPC | REST Gateway |
|---------|------|--------------|
| CBT | 6001 | 6009 |
| LMS | 6000 | 6008 |
