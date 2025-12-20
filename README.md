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

- **Autentikasi**: login/logout untuk Admin dan Siswa
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

## 📌 API (Admin & Siswa)

**Autentikasi (semua pengguna)**

- `POST /api/auth/login` — Login (body: `{email, password}`) → Response: `{token}`
- `POST /api/auth/logout` — Logout (butuh Authorization)

**Admin** (Role: `admin`, Header: `Authorization: Bearer <token>`)

- `GET /api/admin/users` — Daftar pengguna
- `POST /api/admin/users` — Buat pengguna (body: `{name,email,role,password}`)
- `GET /api/admin/users/{id}` — Detail pengguna
- `PUT /api/admin/users/{id}` — Update pengguna
- `DELETE /api/admin/users/{id}` — Hapus pengguna

- `GET /api/admin/questions` — Daftar soal
- `POST /api/admin/questions` — Buat soal (body: `{title,type,options,answer,subject,level}`)
- `PUT /api/admin/questions/{id}` — Update soal
- `DELETE /api/admin/questions/{id}` — Hapus soal

- `GET /api/admin/test-sessions` — Daftar sesi
- `POST /api/admin/test-sessions` — Buat sesi (body: `{title,start_at,end_at,question_ids}`)
- `PUT /api/admin/test-sessions/{id}` — Update sesi
- `DELETE /api/admin/test-sessions/{id}` — Hapus sesi

- `GET /api/admin/reports` — Laporan hasil / statistik (opsional filter)

**Siswa** (Header: `Authorization: Bearer <token>`)

- `GET /api/test-sessions` — Daftar sesi tersedia
- `GET /api/test-sessions/{id}` — Detail sesi
- `GET /api/test-sessions/{id}/questions` — Ambil soal untuk sesi
- `POST /api/test-sessions/{id}/submit` — Kirim jawaban (body: `{answers: [{question_id, answer}]}`)
- `GET /api/history` — Riwayat tes (siswa)
- `GET /api/users/me` — Profil siswa

> Semua endpoint yang membutuhkan autentikasi menggunakan header: `Authorization: Bearer <token>`.

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