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
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_test_session_tingkat ON test_session (id_tingkat);
CREATE INDEX idx_test_session_mata_pelajaran ON test_session (id_mata_pelajaran);
CREATE INDEX idx_test_session_user ON test_session (user_id);
CREATE INDEX idx_test_session_waktu_mulai ON test_session (waktu_mulai);
CREATE INDEX idx_test_session_status ON test_session (status);
CREATE INDEX idx_test_session_user_status ON test_session (user_id, status);
CREATE INDEX idx_test_session_status_waktu ON test_session (status, waktu_mulai);

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
CREATE TRIGGER update_tingkat_updated_at BEFORE UPDATE ON tingkat FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_materi_updated_at BEFORE UPDATE ON materi FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_soal_updated_at BEFORE UPDATE ON soal FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_soal_drag_drop_updated_at BEFORE UPDATE ON soal_drag_drop FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_test_session_updated_at BEFORE UPDATE ON test_session FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_user_limits_updated_at BEFORE UPDATE ON user_limits FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
