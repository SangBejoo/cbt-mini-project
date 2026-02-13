-- =============================================
-- COMPLETE CONSOLIDATED POSTGRESQL SCHEMA
-- CBT Mini Project - Computer Based Test System
-- Date: February 6, 2026
-- Converted from MySQL Schema (Optimized + Drag Drop Features)
-- =============================================

-- Settings
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SET check_function_bodies = false;
SET client_min_messages = warning;
SET row_security = off;
SET default_tablespace = '';

-- =============================================
-- ENUM TYPES
-- =============================================

CREATE TYPE user_role_enum AS ENUM ('siswa', 'admin');
CREATE TYPE test_session_status_enum AS ENUM ('ongoing', 'completed', 'timeout');
CREATE TYPE drag_type_enum AS ENUM ('ordering', 'matching');
CREATE TYPE question_type_enum AS ENUM ('multiple_choice', 'drag_drop');

-- =============================================
-- MASTER TABLES
-- =============================================

-- Table: Users
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    nama VARCHAR(100) NOT NULL,
    role user_role_enum NOT NULL DEFAULT 'siswa',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_role ON users (role);
CREATE INDEX idx_users_is_active ON users (is_active);
CREATE INDEX idx_users_active_role ON users (is_active, role);
CREATE INDEX idx_users_created_at ON users (created_at);

-- Table: Mata Pelajaran
CREATE TABLE mata_pelajaran (
    id SERIAL PRIMARY KEY,
    nama VARCHAR(50) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_mata_pelajaran_is_active ON mata_pelajaran (is_active);

-- Table: Tingkat
CREATE TABLE tingkat (
    id SERIAL PRIMARY KEY,
    nama VARCHAR(50) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tingkat_is_active ON tingkat (is_active);

-- Table: Classes
CREATE TABLE classes (
    id SERIAL PRIMARY KEY,
    teacher_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT NULL,
    enrollment_code VARCHAR(20) UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_classes_teacher ON classes (teacher_id);
CREATE INDEX idx_classes_code ON classes (enrollment_code);
CREATE INDEX idx_classes_is_active ON classes (is_active);

-- Table: Class Students
CREATE TABLE class_students (
    id SERIAL PRIMARY KEY,
    class_id INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE ON UPDATE CASCADE,
    student_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (class_id, student_id)
);

CREATE INDEX idx_class_students_class ON class_students (class_id);
CREATE INDEX idx_class_students_student ON class_students (student_id);

-- =============================================
-- CONTENT TABLES
-- =============================================

-- Table: Materi
CREATE TABLE materi (
    id SERIAL PRIMARY KEY,
    id_mata_pelajaran INTEGER NOT NULL REFERENCES mata_pelajaran(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    id_tingkat INTEGER NOT NULL REFERENCES tingkat(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    nama VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    default_durasi_menit INTEGER NOT NULL DEFAULT 60,
    default_jumlah_soal INTEGER NOT NULL DEFAULT 20,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (id_mata_pelajaran, id_tingkat, nama)
);

CREATE INDEX idx_materi_mata_pelajaran ON materi (id_mata_pelajaran);
CREATE INDEX idx_materi_tingkat ON materi (id_tingkat);
CREATE INDEX idx_materi_is_active ON materi (is_active);
CREATE INDEX idx_materi_active_tingkat ON materi (is_active, id_tingkat);
CREATE INDEX idx_materi_active_mata_pelajaran ON materi (is_active, id_mata_pelajaran);
CREATE INDEX idx_materi_composite ON materi (id_mata_pelajaran, id_tingkat, is_active);

-- Table: Soal (Multiple Choice)
CREATE TABLE soal (
    id SERIAL PRIMARY KEY,
    id_materi INTEGER NOT NULL REFERENCES materi(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    id_tingkat INTEGER NOT NULL REFERENCES tingkat(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    pertanyaan TEXT NOT NULL,
    opsi_a VARCHAR(500) NOT NULL,
    opsi_b VARCHAR(500) NOT NULL,
    opsi_c VARCHAR(500) NOT NULL,
    opsi_d VARCHAR(500) NOT NULL,
    jawaban_benar CHAR(1) NOT NULL,
    pembahasan TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    image_path VARCHAR(255) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_soal_materi ON soal (id_materi);
CREATE INDEX idx_soal_tingkat ON soal (id_tingkat);
CREATE INDEX idx_soal_is_active ON soal (is_active);
CREATE INDEX idx_soal_active_materi ON soal (is_active, id_materi);
CREATE INDEX idx_soal_active_tingkat ON soal (is_active, id_tingkat);
CREATE INDEX idx_soal_materi_active ON soal (id_materi, is_active);
CREATE INDEX idx_soal_created_at ON soal (created_at);

-- Table: Soal Gambar (Multiple Choice Images)
CREATE TABLE soal_gambar (
    id SERIAL PRIMARY KEY,
    id_soal INTEGER NOT NULL REFERENCES soal(id) ON DELETE CASCADE ON UPDATE CASCADE,
    nama_file VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size INTEGER NOT NULL,
    mime_type VARCHAR(50) NOT NULL,
    urutan SMALLINT NOT NULL DEFAULT 1,
    keterangan VARCHAR(255) NULL,
    cloud_id VARCHAR(255) NULL,
    public_id VARCHAR(500) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_soal_gambar_soal ON soal_gambar (id_soal);
CREATE INDEX idx_soal_gambar_soal_urutan ON soal_gambar (id_soal, urutan);
CREATE INDEX idx_soal_gambar_cloud_id ON soal_gambar (cloud_id);

-- Table: Soal Drag Drop
CREATE TABLE soal_drag_drop (
    id SERIAL PRIMARY KEY,
    id_materi INTEGER NOT NULL REFERENCES materi(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    id_tingkat INTEGER NOT NULL REFERENCES tingkat(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    pertanyaan TEXT NOT NULL,
    drag_type drag_type_enum NOT NULL,
    pembahasan TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_soal_drag_drop_materi ON soal_drag_drop (id_materi);
CREATE INDEX idx_soal_drag_drop_tingkat ON soal_drag_drop (id_tingkat);
CREATE INDEX idx_soal_drag_drop_is_active ON soal_drag_drop (is_active);
CREATE INDEX idx_soal_drag_drop_type ON soal_drag_drop (drag_type);
CREATE INDEX idx_soal_drag_drop_active_materi ON soal_drag_drop (is_active, id_materi);
CREATE INDEX idx_soal_drag_drop_active_tingkat ON soal_drag_drop (is_active, id_tingkat);

-- Table: Soal Drag Drop Gambar
CREATE TABLE soal_drag_drop_gambar (
    id SERIAL PRIMARY KEY,
    id_soal_drag_drop INTEGER NOT NULL REFERENCES soal_drag_drop(id) ON DELETE CASCADE,
    nama_file VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size INTEGER NOT NULL,
    mime_type VARCHAR(50) NOT NULL,
    urutan SMALLINT NOT NULL DEFAULT 1,
    keterangan VARCHAR(255) NULL,
    cloud_id VARCHAR(255) NULL,
    public_id VARCHAR(500) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_soal_drag_drop_gambar_soal ON soal_drag_drop_gambar (id_soal_drag_drop);

-- Table: Drag Item
CREATE TABLE drag_item (
    id SERIAL PRIMARY KEY,
    id_soal_drag_drop INTEGER NOT NULL REFERENCES soal_drag_drop(id) ON DELETE CASCADE ON UPDATE CASCADE,
    label VARCHAR(255) NOT NULL,
    image_url VARCHAR(500) NULL,
    urutan SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_drag_item_soal ON drag_item (id_soal_drag_drop);
CREATE INDEX idx_drag_item_urutan ON drag_item (id_soal_drag_drop, urutan);

-- Table: Drag Slot
CREATE TABLE drag_slot (
    id SERIAL PRIMARY KEY,
    id_soal_drag_drop INTEGER NOT NULL REFERENCES soal_drag_drop(id) ON DELETE CASCADE ON UPDATE CASCADE,
    label VARCHAR(255) NOT NULL,
    image_url VARCHAR(500) NULL,
    urutan SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_drag_slot_soal ON drag_slot (id_soal_drag_drop);
CREATE INDEX idx_drag_slot_urutan ON drag_slot (id_soal_drag_drop, urutan);

-- Table: Drag Correct Answer
CREATE TABLE drag_correct_answer (
    id SERIAL PRIMARY KEY,
    id_drag_item INTEGER NOT NULL REFERENCES drag_item(id) ON DELETE CASCADE ON UPDATE CASCADE,
    id_drag_slot INTEGER NOT NULL REFERENCES drag_slot(id) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (id_drag_item, id_drag_slot)
);

CREATE INDEX idx_drag_correct_item ON drag_correct_answer (id_drag_item);
CREATE INDEX idx_drag_correct_slot ON drag_correct_answer (id_drag_slot);
CREATE INDEX idx_drag_correct_answer_item_slot ON drag_correct_answer (id_drag_item, id_drag_slot);

-- =============================================
-- TEST SESSION TABLES
-- =============================================

-- Table: Test Session
CREATE TABLE test_session (
    id SERIAL PRIMARY KEY,
    session_token VARCHAR(64) NOT NULL UNIQUE,
    nama_peserta VARCHAR(100) NOT NULL,
    id_tingkat INTEGER NOT NULL REFERENCES tingkat(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    id_mata_pelajaran INTEGER NOT NULL REFERENCES mata_pelajaran(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    user_id INTEGER NULL REFERENCES users(id) ON DELETE SET NULL ON UPDATE CASCADE,
    waktu_mulai TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    waktu_selesai TIMESTAMPTZ NULL,
    durasi_menit INTEGER NOT NULL,
    nilai_akhir DECIMAL(5,2) NULL,
    jumlah_benar INTEGER NULL,
    total_soal INTEGER NULL,
    status test_session_status_enum NOT NULL DEFAULT 'ongoing',
    lms_assignment_id INTEGER NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_test_session_tingkat ON test_session (id_tingkat);
CREATE INDEX idx_test_session_mata_pelajaran ON test_session (id_mata_pelajaran);
CREATE INDEX idx_test_session_user ON test_session (user_id);
CREATE INDEX idx_test_session_waktu_mulai ON test_session (waktu_mulai);
CREATE INDEX idx_test_session_status ON test_session (status);
CREATE INDEX idx_test_session_user_status ON test_session (user_id, status);
CREATE INDEX idx_test_session_user_status ON test_session (user_id, status);
CREATE INDEX idx_test_session_status_waktu ON test_session (status, waktu_mulai);
CREATE INDEX idx_test_session_lms_assignment ON test_session (lms_assignment_id);

-- Table: Test Session Soal (Unified)
CREATE TABLE test_session_soal (
    id SERIAL PRIMARY KEY,
    id_test_session INTEGER NOT NULL REFERENCES test_session(id) ON DELETE CASCADE ON UPDATE CASCADE,
    id_soal INTEGER NULL REFERENCES soal(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    id_soal_drag_drop INTEGER NULL REFERENCES soal_drag_drop(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    question_type question_type_enum NOT NULL DEFAULT 'multiple_choice',
    nomor_urut SMALLINT NOT NULL,
    UNIQUE (id_test_session, nomor_urut)
);

CREATE INDEX idx_test_session_soal_session ON test_session_soal (id_test_session);
CREATE INDEX idx_test_session_soal_soal ON test_session_soal (id_soal);
CREATE INDEX idx_test_session_soal_drag_drop ON test_session_soal (id_soal_drag_drop);
CREATE INDEX idx_test_session_soal_type ON test_session_soal (question_type);

-- Table: Jawaban Siswa (Unified)
CREATE TABLE jawaban_siswa (
    id SERIAL PRIMARY KEY,
    id_test_session_soal INTEGER NOT NULL REFERENCES test_session_soal(id) ON DELETE CASCADE ON UPDATE CASCADE,
    jawaban_dipilih CHAR(1) NULL,
    question_type question_type_enum NOT NULL DEFAULT 'multiple_choice',
    jawaban_drag_drop JSONB NULL,
    is_correct BOOLEAN NOT NULL DEFAULT FALSE,
    dijawab_pada TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (id_test_session_soal)
);

CREATE INDEX idx_jawaban_siswa_is_correct ON jawaban_siswa (is_correct);
CREATE INDEX idx_jawaban_siswa_dijawab_pada ON jawaban_siswa (dijawab_pada);
CREATE INDEX idx_jawaban_siswa_type ON jawaban_siswa (question_type);

-- =============================================
-- USER LIMITS & USAGE TRACKING
-- =============================================

-- Table: User Limits
CREATE TABLE user_limits (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    limit_type VARCHAR(100) NOT NULL,
    limit_value INTEGER NOT NULL DEFAULT 0,
    current_used INTEGER NOT NULL DEFAULT 0,
    reset_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, limit_type)
);

CREATE INDEX idx_user_limits_user ON user_limits (user_id);
CREATE INDEX idx_user_limits_reset_at ON user_limits (reset_at);
CREATE INDEX idx_user_limits_user_type ON user_limits (user_id, limit_type);
CREATE INDEX idx_user_limits_type_reset ON user_limits (limit_type, reset_at);

-- Table: User Limit Usage
CREATE TABLE user_limit_usage (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    limit_type VARCHAR(100) NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource_id INTEGER NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_user_limit_usage_user ON user_limit_usage (user_id);
CREATE INDEX idx_user_limit_usage_type ON user_limit_usage (limit_type);
CREATE INDEX idx_user_limit_usage_created_at ON user_limit_usage (created_at);
CREATE INDEX idx_user_limit_usage_resource_id ON user_limit_usage (resource_id);
CREATE INDEX idx_user_limit_usage_user_created ON user_limit_usage (user_id, created_at);
CREATE INDEX idx_user_limit_usage_user_type_created ON user_limit_usage (user_id, limit_type, created_at);

-- =============================================
-- TRIGGER FUNCTION FOR UPDATED_AT
-- =============================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_mata_pelajaran_updated_at BEFORE UPDATE ON mata_pelajaran FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_mata_pelajaran_updated_at BEFORE UPDATE ON mata_pelajaran FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_tingkat_updated_at BEFORE UPDATE ON tingkat FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_classes_updated_at BEFORE UPDATE ON classes FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_materi_updated_at BEFORE UPDATE ON materi FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_soal_updated_at BEFORE UPDATE ON soal FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_soal_drag_drop_updated_at BEFORE UPDATE ON soal_drag_drop FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_test_session_updated_at BEFORE UPDATE ON test_session FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_user_limits_updated_at BEFORE UPDATE ON user_limits FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- =============================================
-- SEED DATA
-- =============================================

-- Insert Users
INSERT INTO users (email, password_hash, nama, role, is_active) VALUES
('admin@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Administrator', 'admin', TRUE),
('siswa1@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Siswa Satu', 'siswa', TRUE),
('siswa2@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Siswa Dua', 'siswa', TRUE),
('siswa3@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Siswa Tiga', 'siswa', TRUE);

-- Insert Mata Pelajaran
INSERT INTO mata_pelajaran (nama) VALUES
('Matematika'),
('Bahasa Indonesia'),
('Bahasa Inggris'),
('IPA'),
('IPS');

-- Insert Tingkat
INSERT INTO tingkat (nama) VALUES
('Tingkat 1'),
('Tingkat 2'),
('Tingkat 3'),
('Tingkat 4'),
('Tingkat 5');

-- Insert Materi
INSERT INTO materi (id_mata_pelajaran, id_tingkat, nama) VALUES
-- Matematika
(1, 1, 'Bilangan dan Operasi Dasar'),
(1, 2, 'Geometri Dasar'),
(1, 3, 'Pengukuran'),
(1, 4, 'Pecahan dan Desimal'),
(1, 5, 'Perbandingan dan Persentase'),
-- Bahasa Indonesia
(2, 1, 'Huruf dan Kata'),
(2, 2, 'Kalimat Sederhana'),
(2, 3, 'Cerita Pendek'),
(2, 4, 'Pantun dan Puisi'),
(2, 5, 'Teks Deskriptif'),
-- Bahasa Inggris
(3, 1, 'Alphabet and Numbers'),
(3, 2, 'Family and Friends'),
(3, 3, 'Animals and Colors'),
(3, 4, 'Daily Activities'),
(3, 5, 'Simple Sentences'),
-- IPA
(4, 1, 'Tubuh Manusia'),
(4, 2, 'Hewan dan Tumbuhan'),
(4, 3, 'Benda di Sekitar'),
(4, 4, 'Energi dan Gerak'),
(4, 5, 'Lingkungan Hidup'),
-- IPS
(5, 1, 'Diri Sendiri dan Keluarga'),
(5, 2, 'Sekolah dan Teman'),
(5, 3, 'Lingkungan Rumah'),
(5, 4, 'Kota dan Desa'),
(5, 5, 'Negara Indonesia');

-- Insert Soal (10 per materi)
INSERT INTO soal (id_materi, id_tingkat, pertanyaan, opsi_a, opsi_b, opsi_c, opsi_d, jawaban_benar) VALUES
-- Matematika Tingkat 1: Bilangan dan Operasi Dasar
(1, 1, 'Berapa 1 + 1?', '1', '2', '3', '4', 'B'),
(1, 1, 'Berapa 2 + 2?', '3', '4', '5', '6', 'B'),
(1, 1, 'Berapa 3 + 1?', '2', '3', '4', '5', 'C'),
(1, 1, 'Berapa 4 - 2?', '1', '2', '3', '4', 'B'),
(1, 1, 'Berapa 5 - 1?', '3', '4', '5', '6', 'B'),
(1, 1, 'Berapa 2 x 2?', '2', '3', '4', '5', 'C'),
(1, 1, 'Berapa 6 / 2?', '2', '3', '4', '5', 'B'),
(1, 1, 'Berapa 3 + 2?', '4', '5', '6', '7', 'B'),
(1, 1, 'Berapa 7 - 3?', '3', '4', '5', '6', 'B'),
(1, 1, 'Berapa 1 + 3?', '3', '4', '5', '6', 'B'),
-- Matematika Tingkat 2: Geometri Dasar
(2, 2, 'Bangun datar dengan 3 sisi adalah?', 'Lingkaran', 'Segitiga', 'Persegi', 'Persegi Panjang', 'B'),
(2, 2, 'Bangun datar dengan 4 sisi sama panjang adalah?', 'Segitiga', 'Persegi', 'Lingkaran', 'Trapesium', 'B'),
(2, 2, 'Berapa sisi persegi?', '3', '4', '5', '6', 'B'),
(2, 2, 'Bangun ruang dengan 6 sisi adalah?', 'Kubus', 'Balok', 'Prisma', 'Limas', 'A'),
(2, 2, 'Apa nama bangun datar bundar?', 'Segitiga', 'Persegi', 'Lingkaran', 'Trapesium', 'C'),
(2, 2, 'Berapa sisi segitiga?', '2', '3', '4', '5', 'B'),
(2, 2, 'Bangun datar dengan 4 sisi berbeda adalah?', 'Persegi', 'Persegi Panjang', 'Trapesium', 'Segitiga', 'C'),
(2, 2, 'Apa itu kubus?', 'Bangun datar', 'Bangun ruang', 'Garis', 'Titik', 'B'),
(2, 2, 'Berapa sisi balok?', '4', '6', '8', '10', 'B'),
(2, 2, 'Bangun datar dengan semua sisi sama adalah?', 'Trapesium', 'Jajar Genjang', 'Persegi', 'Segitiga', 'C'),
-- Matematika Tingkat 3: Pengukuran
(3, 3, 'Satuan panjang adalah?', 'Liter', 'Meter', 'Kg', 'Detik', 'B'),
(3, 3, 'Berapa cm dalam 1 meter?', '10', '100', '1000', '10000', 'B'),
(3, 3, 'Satuan berat adalah?', 'Meter', 'Liter', 'Kg', 'Cm', 'C'),
(3, 3, 'Berapa mm dalam 1 cm?', '10', '100', '1000', '10000', 'A'),
(3, 3, 'Satuan waktu adalah?', 'Meter', 'Liter', 'Kg', 'Jam', 'D'),
(3, 3, 'Berapa detik dalam 1 menit?', '10', '60', '100', '1000', 'B'),
(3, 3, 'Satuan volume adalah?', 'Meter', 'Liter', 'Kg', 'Cm', 'B'),
(3, 3, 'Berapa ml dalam 1 liter?', '10', '100', '1000', '10000', 'C'),
(3, 3, 'Alat ukur panjang adalah?', 'Termometer', 'Timbangan', 'Meteran', 'Jam', 'C'),
(3, 3, 'Alat ukur berat adalah?', 'Meteran', 'Timbangan', 'Termometer', 'Jam', 'B'),
-- Matematika Tingkat 4: Pecahan dan Desimal
(4, 4, '1/2 sama dengan?', '0.1', '0.2', '0.5', '1.0', 'C'),
(4, 4, '0.5 sama dengan?', '1/2', '1/3', '1/4', '1/5', 'A'),
(4, 4, 'Berapa 1/4 + 1/4?', '1/2', '1/3', '1/8', '2/4', 'A'),
(4, 4, '0.25 sama dengan?', '1/4', '1/2', '1/3', '1/5', 'A'),
(4, 4, 'Berapa 0.1 + 0.2?', '0.2', '0.3', '0.4', '0.5', 'B'),
(4, 4, '1/3 dari 9 adalah?', '3', '6', '9', '12', 'A'),
(4, 4, '0.75 sama dengan?', '3/4', '1/2', '1/4', '1/3', 'A'),
(4, 4, 'Berapa 2/5 - 1/5?', '1/5', '2/5', '3/5', '4/5', 'A'),
(4, 4, '0.4 sama dengan?', '2/5', '1/5', '3/5', '4/5', 'A'),
(4, 4, 'Berapa 0.5 x 2?', '0.5', '1.0', '1.5', '2.0', 'B'),
-- Matematika Tingkat 5: Perbandingan dan Persentase
(5, 5, 'Perbandingan 2:3 artinya?', '2 lebih besar dari 3', '3 lebih besar dari 2', 'Sama', 'Tidak ada', 'B'),
(5, 5, '50% sama dengan?', '1/2', '1/3', '1/4', '1/5', 'A'),
(5, 5, 'Perbandingan 4:2 disederhanakan menjadi?', '2:1', '4:2', '1:2', '2:4', 'A'),
(5, 5, '25% dari 100 adalah?', '25', '50', '75', '100', 'A'),
(5, 5, 'Perbandingan 3:6 disederhanakan menjadi?', '1:2', '3:6', '2:3', '1:3', 'A'),
(5, 5, '75% sama dengan?', '3/4', '1/2', '1/4', '1/3', 'A'),
(5, 5, 'Perbandingan 5:10 adalah?', '1:2', '2:5', '5:10', '10:5', 'A'),
(5, 5, '100% sama dengan?', '1', '0.5', '0.25', '0', 'A'),
(5, 5, 'Perbandingan 6:4 disederhanakan menjadi?', '3:2', '6:4', '2:3', '4:6', 'A'),
(5, 5, '10% dari 200 adalah?', '10', '20', '30', '40', 'B'),
-- Bahasa Indonesia Tingkat 1: Huruf dan Kata
(6, 1, 'Huruf pertama dalam abjad adalah?', 'A', 'B', 'C', 'D', 'A'),
(6, 1, 'Kata yang dimulai dengan huruf B adalah?', 'Apel', 'Bola', 'Cincin', 'Domba', 'B'),
(6, 1, 'Berapa jumlah huruf vokal?', '5', '10', '15', '20', 'A'),
(6, 1, 'Huruf konsonan contohnya?', 'A', 'I', 'U', 'B', 'D'),
(6, 1, 'Kata yang berarti buah adalah?', 'Meja', 'Apel', 'Buku', 'Kursi', 'B'),
(6, 1, 'Huruf terakhir dalam abjad adalah?', 'X', 'Y', 'Z', 'W', 'C'),
(6, 1, 'Kata yang dimulai dengan huruf M adalah?', 'Nasi', 'Meja', 'Bola', 'Kucing', 'B'),
(6, 1, 'Berapa huruf dalam kata "kucing"?', '5', '6', '7', '8', 'B'),
(6, 1, 'Huruf vokal adalah?', 'B', 'C', 'D', 'A', 'D'),
(6, 1, 'Kata yang berarti hewan adalah?', 'Meja', 'Buku', 'Kucing', 'Bola', 'C'),
-- Bahasa Indonesia Tingkat 2: Kalimat Sederhana
(7, 2, 'Kalimat "Saya makan nasi" adalah kalimat?', 'Tanya', 'Perintah', 'Berita', 'Seru', 'C'),
(7, 2, 'Kata kerja dalam "Dia lari" adalah?', 'Dia', 'Lari', 'Dan', 'Ke', 'B'),
(7, 2, 'Kalimat tanya diakhiri dengan?', '.', '!', '?', ',', 'C'),
(7, 2, 'Subjek dalam "Budi makan" adalah?', 'Makan', 'Budi', 'Dan', 'Ke', 'B'),
(7, 2, 'Kalimat perintah contohnya?', 'Apakah kamu lapar?', 'Makanlah nasi!', 'Saya lapar.', 'Dia makan.', 'B'),
(7, 2, 'Predikat dalam "Anak bermain" adalah?', 'Anak', 'Bermain', 'Dan', 'Ke', 'B'),
(7, 2, 'Kalimat seru diakhiri dengan?', '.', '!', '?', ',', 'B'),
(7, 2, 'Objek dalam "Saya makan apel" adalah?', 'Saya', 'Makan', 'Apel', 'Dan', 'C'),
(7, 2, 'Kalimat berita diakhiri dengan?', '.', '!', '?', ',', 'A'),
(7, 2, 'Kata sifat dalam "Bunga merah" adalah?', 'Bunga', 'Merah', 'Dan', 'Ke', 'B'),
-- Bahasa Indonesia Tingkat 3: Cerita Pendek
(8, 3, 'Tokoh utama dalam cerita adalah?', 'Penulis', 'Pembaca', 'Orang yang diceritakan', 'Hewan', 'C'),
(8, 3, 'Awal cerita disebut?', 'Akhir', 'Tengah', 'Awal', 'Penutup', 'C'),
(8, 3, 'Bagian cerita yang menyelesaikan masalah adalah?', 'Pendahuluan', 'Isi', 'Penutup', 'Pengantar', 'C'),
(8, 3, 'Cerita pendek biasanya memiliki?', 'Banyak tokoh', 'Sedikit tokoh', 'Tidak ada tokoh', 'Hanya hewan', 'B'),
(8, 3, 'Latar dalam cerita adalah?', 'Waktu dan tempat', 'Tokoh', 'Akhir', 'Awal', 'A'),
(8, 3, 'Konflik dalam cerita adalah?', 'Masalah', 'Penyelesaian', 'Akhir', 'Awal', 'A'),
(8, 3, 'Climax adalah?', 'Awal cerita', 'Puncak cerita', 'Akhir cerita', 'Pendahuluan', 'B'),
(8, 3, 'Resolusi adalah?', 'Masalah', 'Penyelesaian', 'Awal', 'Tengah', 'B'),
(8, 3, 'Cerita fiksi artinya?', 'Berdasarkan fakta', 'Tidak berdasarkan fakta', 'Hanya tentang hewan', 'Hanya tentang manusia', 'B'),
(8, 3, 'Tema cerita adalah?', 'Pesan utama', 'Tokoh', 'Latar', 'Akhir', 'A'),
-- Bahasa Indonesia Tingkat 4: Pantun dan Puisi
(9, 4, 'Pantun terdiri dari?', '2 baris', '4 baris', '6 baris', '8 baris', 'B'),
(9, 4, 'Sampiran dalam pantun adalah?', 'Baris 1 dan 2', 'Baris 3 dan 4', 'Baris 1 dan 3', 'Baris 2 dan 4', 'A'),
(9, 4, 'Isi dalam pantun adalah?', 'Baris 1 dan 2', 'Baris 3 dan 4', 'Baris 1 dan 3', 'Baris 2 dan 4', 'B'),
(9, 4, 'Puisi tidak memiliki?', 'Rima', 'Ritme', 'Prosa', 'Imajinasi', 'C'),
(9, 4, 'Rima dalam puisi adalah?', 'Pengulangan bunyi', 'Pengulangan kata', 'Pengulangan kalimat', 'Pengulangan cerita', 'A'),
(9, 4, 'Pantun biasanya?', 'Serius', 'Lucu', 'Bergantung', 'Tidak ada', 'C'),
(9, 4, 'Baris dalam puisi disebut?', 'Kalimat', 'Larik', 'Paragraf', 'Bab', 'B'),
(9, 4, 'Puisi bebas tidak memiliki?', 'Rima', 'Ritme', 'Larik', 'Kata', 'A'),
(9, 4, 'Pantun daerah adalah?', 'Pantun nasional', 'Pantun lokal', 'Pantun internasional', 'Pantun global', 'B'),
(9, 4, 'Fungsi puisi adalah?', 'Hanya hiburan', 'Ekspresi perasaan', 'Hanya pendidikan', 'Hanya informasi', 'B'),
-- Bahasa Indonesia Tingkat 5: Teks Deskriptif
(10, 5, 'Teks deskriptif bertujuan?', 'Menceritakan', 'Menggambarkan', 'Membujuk', 'Menganjurkan', 'B'),
(10, 5, 'Struktur teks deskriptif?', 'Identifikasi dan Deskripsi', 'Pendahuluan dan Penutup', 'Masalah dan Solusi', 'Kronologis', 'A'),
(10, 5, 'Bahasa dalam teks deskriptif?', 'Imperatif', 'Deklaratif', 'Interogatif', 'Eksklamatif', 'B'),
(10, 5, 'Contoh teks deskriptif?', 'Resep masak', 'Deskripsi binatang', 'Surat', 'Cerita', 'B'),
(10, 5, 'Kata adjektiva dalam deskripsi?', 'Kata kerja', 'Kata sifat', 'Kata depan', 'Kata sambung', 'B'),
(10, 5, 'Teks deskriptif fokus pada?', 'Proses', 'Ciri-ciri', 'Langkah-langkah', 'Peristiwa', 'B'),
(10, 5, 'Paragraf deskripsi dimulai dengan?', 'Kesimpulan', 'Topik sentence', 'Detail', 'Penutup', 'B'),
(10, 5, 'Bahasa deskriptif menggunakan?', 'Kata kerja aktif', 'Kata sifat', 'Kata tanya', 'Kata perintah', 'B'),
(10, 5, 'Objek deskripsi bisa?', 'Tempat', 'Orang', 'Benda', 'Semua benar', 'D'),
(10, 5, 'Tujuan deskripsi adalah?', 'Membuat pembaca membayangkan', 'Membuat pembaca bertanya', 'Membuat pembaca marah', 'Membuat pembaca tertawa', 'A'),
-- Bahasa Inggris Tingkat 1: Alphabet and Numbers
(11, 1, 'What is the first letter?', 'A', 'B', 'C', 'D', 'A'),
(11, 1, 'How many letters in alphabet?', '25', '26', '27', '28', 'B'),
(11, 1, 'What number is after 1?', '0', '2', '3', '4', 'B'),
(11, 1, 'What is 2 + 2?', '3', '4', '5', '6', 'B'),
(11, 1, 'What letter is Z?', 'Last', 'First', 'Middle', 'None', 'A'),
(11, 1, 'How many vowels?', '5', '6', '7', '8', 'A'),
(11, 1, 'What is 5 - 2?', '2', '3', '4', '5', 'B'),
(11, 1, 'What letter is before B?', 'A', 'C', 'D', 'E', 'A'),
(11, 1, 'What is 1 + 1?', '1', '2', '3', '4', 'B'),
(11, 1, 'What number is 10?', 'Nine', 'Ten', 'Eleven', 'Twelve', 'B'),
-- Bahasa Inggris Tingkat 2: Family and Friends
(12, 2, 'Who is your mother?', 'Mom', 'Dad', 'Sister', 'Brother', 'A'),
(12, 2, 'What is father called?', 'Mom', 'Dad', 'Aunt', 'Uncle', 'B'),
(12, 2, 'Sister is?', 'Boy', 'Girl', 'Man', 'Woman', 'B'),
(12, 2, 'Brother is?', 'Boy', 'Girl', 'Man', 'Woman', 'A'),
(12, 2, 'Grandmother is?', 'Grandma', 'Grandpa', 'Aunt', 'Uncle', 'A'),
(12, 2, 'Friend is?', 'Enemy', 'Family', 'Companion', 'Stranger', 'C'),
(12, 2, 'Aunt is sister of?', 'Mother', 'Father', 'Brother', 'Sister', 'A'),
(12, 2, 'Uncle is brother of?', 'Mother', 'Father', 'Aunt', 'Uncle', 'B'),
(12, 2, 'Cousin is child of?', 'Parent', 'Aunt or Uncle', 'Grandparent', 'Friend', 'B'),
(12, 2, 'Family means?', 'Friends', 'Relatives', 'Neighbors', 'Teachers', 'B'),
-- Bahasa Inggris Tingkat 3: Animals and Colors
(13, 3, 'What color is sky?', 'Red', 'Blue', 'Green', 'Yellow', 'B'),
(13, 3, 'What animal says meow?', 'Dog', 'Cat', 'Cow', 'Pig', 'B'),
(13, 3, 'What color is grass?', 'Red', 'Blue', 'Green', 'Yellow', 'C'),
(13, 3, 'What animal is big and gray?', 'Elephant', 'Lion', 'Tiger', 'Bear', 'A'),
(13, 3, 'What color is sun?', 'Red', 'Blue', 'Green', 'Yellow', 'D'),
(13, 3, 'What animal flies?', 'Fish', 'Bird', 'Dog', 'Cat', 'B'),
(13, 3, 'What color is blood?', 'Red', 'Blue', 'Green', 'Yellow', 'A'),
(13, 3, 'What animal lives in water?', 'Fish', 'Bird', 'Dog', 'Cat', 'A'),
(13, 3, 'What color is banana?', 'Red', 'Blue', 'Green', 'Yellow', 'D'),
(13, 3, 'What animal has stripes?', 'Zebra', 'Lion', 'Tiger', 'Bear', 'A'),
-- Bahasa Inggris Tingkat 4: Daily Activities
(14, 4, 'What do you do in morning?', 'Sleep', 'Eat breakfast', 'Watch TV', 'Play', 'B'),
(14, 4, 'What is brush teeth?', 'Wash face', 'Clean teeth', 'Comb hair', 'Wear clothes', 'B'),
(14, 4, 'What do you do at school?', 'Sleep', 'Study', 'Play games', 'Eat', 'B'),
(14, 4, 'What is lunch?', 'Breakfast', 'Dinner', 'Midday meal', 'Snack', 'C'),
(14, 4, 'What do you do after school?', 'Go to bed', 'Do homework', 'Eat dinner', 'Wake up', 'B'),
(14, 4, 'What is read a book?', 'Write', 'Read', 'Draw', 'Sing', 'B'),
(14, 4, 'What do you do at night?', 'Sleep', 'Eat', 'Study', 'Play', 'A'),
(14, 4, 'What is take a bath?', 'Wash body', 'Brush teeth', 'Comb hair', 'Wear clothes', 'A'),
(14, 4, 'What do you do in afternoon?', 'Sleep', 'Play', 'Eat lunch', 'Go to school', 'B'),
(14, 4, 'What is dinner?', 'Breakfast', 'Lunch', 'Evening meal', 'Snack', 'C'),
-- Bahasa Inggris Tingkat 5: Simple Sentences
(15, 5, 'I ___ a student.', 'am', 'is', 'are', 'be', 'A'),
(15, 5, 'She ___ happy.', 'am', 'is', 'are', 'be', 'B'),
(15, 5, 'We ___ friends.', 'am', 'is', 'are', 'be', 'C'),
(15, 5, 'He ___ running.', 'am', 'is', 'are', 'be', 'B'),
(15, 5, 'They ___ playing.', 'am', 'is', 'are', 'be', 'C'),
(15, 5, 'The cat ___ black.', 'am', 'is', 'are', 'be', 'B'),
(15, 5, 'I ___ eating.', 'am', 'is', 'are', 'be', 'A'),
(15, 5, 'You ___ tall.', 'am', 'is', 'are', 'be', 'C'),
(15, 5, 'It ___ raining.', 'am', 'is', 'are', 'be', 'B'),
(15, 5, 'We ___ singing.', 'am', 'is', 'are', 'be', 'C'),
-- IPA Tingkat 1: Tubuh Manusia
(16, 1, 'Bagian tubuh untuk melihat adalah?', 'Telinga', 'Mata', 'Hidung', 'Mulut', 'B'),
(16, 1, 'Bagian tubuh untuk mendengar adalah?', 'Mata', 'Telinga', 'Hidung', 'Mulut', 'B'),
(16, 1, 'Bagian tubuh untuk mencium adalah?', 'Mata', 'Telinga', 'Hidung', 'Mulut', 'C'),
(16, 1, 'Bagian tubuh untuk makan adalah?', 'Mata', 'Telinga', 'Hidung', 'Mulut', 'D'),
(16, 1, 'Berapa jumlah jari tangan?', '5', '10', '15', '20', 'B'),
(16, 1, 'Organ pernapasan adalah?', 'Jantung', 'Paru-paru', 'Hati', 'Ginjal', 'B'),
(16, 1, 'Darah mengalir melalui?', 'Pembuluh darah', 'Tulang', 'Otot', 'Kulit', 'A'),
(16, 1, 'Fungsi jantung adalah?', 'Bernapas', 'Pompa darah', 'Makan', 'Melihat', 'B'),
(16, 1, 'Tulang berfungsi untuk?', 'Dukung tubuh', 'Bernapas', 'Makan', 'Melihat', 'A'),
(16, 1, 'Kulit berfungsi untuk?', 'Lindungi tubuh', 'Bernapas', 'Makan', 'Melihat', 'A'),
-- IPA Tingkat 2: Hewan dan Tumbuhan
(17, 2, 'Hewan yang hidup di air adalah?', 'Ikan', 'Kucing', 'Anjing', 'Burung', 'A'),
(17, 2, 'Tumbuhan menghasilkan?', 'Oksigen', 'Karbon dioksida', 'Air', 'Makanan', 'A'),
(17, 2, 'Hewan pemakan daging adalah?', 'Karnivora', 'Herbivora', 'Omnivora', 'Insektivora', 'A'),
(17, 2, 'Akar tumbuhan berfungsi?', 'Menyerap air', 'Bernapas', 'Makan', 'Melihat', 'A'),
(17, 2, 'Daun tumbuhan berfungsi?', 'Fotosintesis', 'Bernapas', 'Makan', 'Melihat', 'A'),
(17, 2, 'Hewan yang hidup di darat adalah?', 'Ikan', 'Katak', 'Anjing', 'Burung', 'C'),
(17, 2, 'Bunga tumbuhan berfungsi?', 'Reproduksi', 'Bernapas', 'Makan', 'Melihat', 'A'),
(17, 2, 'Hewan herbivora makan?', 'Daging', 'Tumbuhan', 'Semua', 'Tidak makan', 'B'),
(17, 2, 'Tumbuhan butuh?', 'Matahari', 'Gelap', 'Air saja', 'Tanah saja', 'A'),
(17, 2, 'Hewan omnivora makan?', 'Daging dan tumbuhan', 'Daging saja', 'Tumbuhan saja', 'Tidak makan', 'A'),
-- IPA Tingkat 3: Benda di Sekitar
(18, 3, 'Benda padat contohnya?', 'Air', 'Es', 'Uap', 'Gas', 'B'),
(18, 3, 'Benda cair contohnya?', 'Es', 'Air', 'Uap', 'Batu', 'B'),
(18, 3, 'Benda gas contohnya?', 'Air', 'Es', 'Uap', 'Batu', 'C'),
(18, 3, 'Magnet menarik?', 'Besi', 'Kayu', 'Kertas', 'Plastik', 'A'),
(18, 3, 'Cahaya bergerak lurus?', 'Ya', 'Tidak', 'Kadang', 'Bergantung', 'A'),
(18, 3, 'Suara bergerak melalui?', 'Udara', 'Vacuum', 'Air', 'Semua', 'D'),
(18, 3, 'Benda panas memuai?', 'Ya', 'Tidak', 'Kadang', 'Bergantung', 'A'),
(18, 3, 'Listrik mengalir melalui?', 'Konduktor', 'Isolator', 'Vacuum', 'Udara', 'A'),
(18, 3, 'Benda terapung di air jika?', 'Lebih ringan', 'Lebih berat', 'Sama', 'Tidak ada', 'A'),
(18, 3, 'Cermin memantulkan?', 'Cahaya', 'Suara', 'Listrik', 'Magnet', 'A'),
-- IPA Tingkat 4: Energi dan Gerak
(19, 4, 'Energi dari matahari adalah?', 'Surya', 'Listrik', 'Kimia', 'Nuklir', 'A'),
(19, 4, 'Gerak lurus beraturan adalah?', 'Konstan kecepatan', 'Berubah', 'Berhenti', 'Tidak ada', 'A'),
(19, 4, 'Energi kinetik adalah?', 'Energi gerak', 'Energi diam', 'Energi panas', 'Energi cahaya', 'A'),
(19, 4, 'Energi potensial adalah?', 'Energi diam', 'Energi gerak', 'Energi panas', 'Energi cahaya', 'A'),
(19, 4, 'Gaya gravitasi menarik?', 'Ke bawah', 'Ke atas', 'Ke samping', 'Tidak ada', 'A'),
(19, 4, 'Kecepatan adalah?', 'Jarak per waktu', 'Waktu per jarak', 'Jarak x waktu', 'Waktu / jarak', 'A'),
(19, 4, 'Energi listrik dari?', 'Baterai', 'Matahari', 'Angin', 'Semua', 'D'),
(19, 4, 'Gerak melingkar contohnya?', 'Roda', 'Lurus', 'Berhenti', 'Tidak ada', 'A'),
(19, 4, 'Fricion adalah?', 'Gaya gesek', 'Gaya dorong', 'Gaya tarik', 'Gaya angkat', 'A'),
(19, 4, 'Energi panas dari?', 'Gesekan', 'Listrik', 'Cahaya', 'Semua', 'D'),
-- IPA Tingkat 5: Lingkungan Hidup
(20, 5, 'Pencemaran udara dari?', 'Asap pabrik', 'Air bersih', 'Tanah subur', 'Hutan hijau', 'A'),
(20, 5, 'Daur ulang membantu?', 'Kurangi sampah', 'Tambah sampah', 'Buat polusi', 'Tidak ada', 'A'),
(20, 5, 'Hutan berfungsi?', 'Serap CO2', 'Buat polusi', 'Kurangi oksigen', 'Tidak ada', 'A'),
(20, 5, 'Ekosistem adalah?', 'Sistem lingkungan', 'Sistem komputer', 'Sistem sekolah', 'Sistem rumah', 'A'),
(20, 5, 'Konservasi air artinya?', 'Hemat air', 'Buang air', 'Polusi air', 'Tidak ada', 'A'),
(20, 5, 'Pemanasan global dari?', 'Gas rumah kaca', 'Oksigen', 'Nitrogen', 'Hidrogen', 'A'),
(20, 5, 'Biodiversitas adalah?', 'Keragaman hayati', 'Keragaman manusia', 'Keragaman benda', 'Tidak ada', 'A'),
(20, 5, 'Deforestasi adalah?', 'Penebangan hutan', 'Penanaman hutan', 'Pelestarian hutan', 'Tidak ada', 'A'),
(20, 5, 'Sumber energi terbarukan?', 'Matahari', 'Batu bara', 'Minyak', 'Gas', 'A'),
(20, 5, 'Lingkungan sehat artinya?', 'Bersih dan hijau', 'Kotor dan polusi', 'Panas dan kering', 'Dingin dan basah', 'A'),
-- IPS Tingkat 1: Diri Sendiri dan Keluarga
(21, 1, 'Keluarga inti terdiri dari?', 'Orang tua dan anak', 'Kakek nenek', 'Teman', 'Tetangga', 'A'),
(21, 1, 'Nama diri sendiri adalah?', 'Nama orang lain', 'Nama saya', 'Nama hewan', 'Nama benda', 'B'),
(21, 1, 'Ayah adalah?', 'Ibu', 'Bapak', 'Kakak', 'Adik', 'B'),
(21, 1, 'Ibu adalah?', 'Ayah', 'Bapak', 'Ibu', 'Kakak', 'C'),
(21, 1, 'Kakak adalah?', 'Adik yang lebih tua', 'Adik yang lebih muda', 'Orang tua', 'Tetangga', 'A'),
(21, 1, 'Adik adalah?', 'Kakak yang lebih tua', 'Kakak yang lebih muda', 'Orang tua', 'Tetangga', 'B'),
(21, 1, 'Kakek adalah?', 'Ayah ayah', 'Ayah ibu', 'Ibu ayah', 'Ibu ibu', 'A'),
(21, 1, 'Nenek adalah?', 'Ibu ayah', 'Ibu ibu', 'Ayah ayah', 'Ayah ibu', 'B'),
(21, 1, 'Rumah adalah tempat?', 'Tinggal', 'Bermain', 'Belajar', 'Bekerja', 'A'),
(21, 1, 'Keluarga besar termasuk?', 'Kakek nenek', 'Teman', 'Tetangga', 'Guru', 'A'),
-- IPS Tingkat 2: Sekolah dan Teman
(22, 2, 'Sekolah adalah tempat?', 'Belajar', 'Bermain saja', 'Tidur', 'Makan', 'A'),
(22, 2, 'Guru mengajar?', 'Murid', 'Orang tua', 'Teman', 'Tetangga', 'A'),
(22, 2, 'Teman adalah?', 'Orang yang baik', 'Orang jahat', 'Orang asing', 'Orang tua', 'A'),
(22, 2, 'Pelajaran di sekolah?', 'Membaca', 'Menulis', 'Berhitung', 'Semua benar', 'D'),
(22, 2, 'Bermain dengan teman?', 'Baik', 'Jahat', 'Berbahaya', 'Tidak ada', 'A'),
(22, 2, 'Aturan di sekolah?', 'Tidak ada', 'Ada', 'Bergantung', 'Tidak penting', 'B'),
(22, 2, 'Membantu teman adalah?', 'Baik', 'Jahat', 'Berbahaya', 'Tidak ada', 'A'),
(22, 2, 'Belajar bersama?', 'Baik', 'Jahat', 'Berbahaya', 'Tidak ada', 'A'),
(22, 2, 'Respek guru?', 'Ya', 'Tidak', 'Kadang', 'Bergantung', 'A'),
(22, 2, 'Fungsi sekolah?', 'Pendidikan', 'Hiburan', 'Olahraga', 'Tidak ada', 'A'),
-- IPS Tingkat 3: Lingkungan Rumah
(23, 3, 'Rumah tetangga harus?', 'Dihormati', 'Dirusak', 'Diambil', 'Ditinggalkan', 'A'),
(23, 3, 'Bersihkan rumah?', 'Ya', 'Tidak', 'Kadang', 'Bergantung', 'A'),
(23, 3, 'Bantu orang tua?', 'Ya', 'Tidak', 'Kadang', 'Bergantung', 'A'),
(23, 3, 'Jaga keamanan rumah?', 'Ya', 'Tidak', 'Kadang', 'Bergantung', 'A'),
(23, 3, 'Lingkungan bersih?', 'Baik', 'Jahat', 'Berbahaya', 'Tidak ada', 'A'),
(23, 3, 'Sampah dibuang di?', 'Tempat sampah', 'Sembarangan', 'Rumah', 'Sekolah', 'A'),
(23, 3, 'Tetangga adalah?', 'Orang di sekitar', 'Orang jauh', 'Teman sekolah', 'Guru', 'A'),
(23, 3, 'Bersama tetangga?', 'Baik', 'Jahat', 'Berbahaya', 'Tidak ada', 'A'),
(23, 3, 'Rumah aman artinya?', 'Aman dari bahaya', 'Bahaya', 'Kotor', 'Rusak', 'A'),
(23, 3, 'Tanggung jawab di rumah?', 'Ada', 'Tidak ada', 'Kadang', 'Bergantung', 'A'),
-- IPS Tingkat 4: Kota dan Desa
(24, 4, 'Kota biasanya?', 'Ramai', 'Sepi', 'Hijau', 'Kecil', 'A'),
(24, 4, 'Desa biasanya?', 'Hijau', 'Ramai', 'Bising', 'Besar', 'A'),
(24, 4, 'Pemerintah kota?', 'Wali Kota', 'Bupati', 'Gubernur', 'Presiden', 'A'),
(24, 4, 'Pemerintah desa?', 'Kepala Desa', 'Wali Kota', 'Bupati', 'Gubernur', 'A'),
(24, 4, 'Transportasi di kota?', 'Mobil', 'Sepeda', 'Kuda', 'Berjalan', 'A'),
(24, 4, 'Pertanian di desa?', 'Ya', 'Tidak', 'Kadang', 'Bergantung', 'A'),
(24, 4, 'Sekolah di kota?', 'Banyak', 'Sedikit', 'Tidak ada', 'Bergantung', 'A'),
(24, 4, 'Lingkungan desa?', 'Alami', 'Buatan', 'Ramai', 'Bising', 'A'),
(24, 4, 'Kota vs Desa?', 'Berbeda', 'Sama', 'Tidak ada', 'Bergantung', 'A'),
(24, 4, 'Kebutuhan hidup?', 'Air', 'Makanan', 'Rumah', 'Semua benar', 'D'),
-- IPS Tingkat 5: Negara Indonesia
(25, 5, 'Ibukota Indonesia?', 'Jakarta', 'Surabaya', 'Bandung', 'Yogyakarta', 'A'),
(25, 5, 'Bendera Indonesia?', 'Merah Putih', 'Biru Putih', 'Hijau Putih', 'Kuning Putih', 'A'),
(25, 5, 'Lagu kebangsaan?', 'Indonesia Raya', 'Halo Halo Bandung', 'Gugur Bunga', 'Rayuan Pulau Kelapa', 'A'),
(25, 5, 'Presiden pertama?', 'Soekarno', 'Suharto', 'Habibie', 'Gus Dur', 'A'),
(25, 5, 'Bahasa resmi?', 'Bahasa Indonesia', 'Jawa', 'Sunda', 'Madura', 'A'),
(25, 5, 'Mata uang?', 'Rupiah', 'Dollar', 'Euro', 'Yen', 'A'),
(25, 5, 'Jumlah provinsi?', '34', '33', '35', '36', 'B'),
(25, 5, 'Lambang negara?', 'Garuda Pancasila', 'Macan', 'Harimau', 'Gajah', 'A'),
(25, 5, 'Hari kemerdekaan?', '17 Agustus', '1 Juni', '25 Desember', '1 Januari', 'A'),
(25, 5, 'Pancasila adalah?', 'Dasar negara', 'Lagu', 'Bendera', 'Lambang', 'A');
