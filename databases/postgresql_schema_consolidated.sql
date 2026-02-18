-- =============================================
-- COMPLETE CONSOLIDATED POSTGRESQL SCHEMA (ENGLISH)
-- CBT Microservice - Computer Based Test System
-- Date: February 18, 2026
-- Refactored: English Naming + Essay Support + LMS Integration
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

CREATE TYPE user_role_enum AS ENUM ('student', 'admin', 'teacher', 'superadmin');
CREATE TYPE exam_session_status_enum AS ENUM ('scheduled', 'ongoing', 'completed', 'timeout', 'grading_in_progress', 'graded');
CREATE TYPE drag_type_enum AS ENUM ('ordering', 'matching');
CREATE TYPE question_type_enum AS ENUM ('multiple_choice', 'drag_drop', 'essay');

-- =============================================
-- MASTER TABLES
-- =============================================

-- Table: Users (Cached from LMS)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    role user_role_enum NOT NULL DEFAULT 'student',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    lms_user_id BIGINT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_role ON users (role);
CREATE INDEX idx_users_is_active ON users (is_active);
CREATE INDEX idx_users_lms_id ON users (lms_user_id);

-- Table: Subjects (Mata Pelajaran)
CREATE TABLE subjects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    lms_subject_id BIGINT NULL,
    lms_school_id BIGINT NULL,
    lms_class_id BIGINT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uq_subjects_lms_subject_id ON subjects (lms_subject_id) WHERE lms_subject_id IS NOT NULL;
CREATE INDEX idx_subjects_school ON subjects (lms_school_id);

-- Table: Grade Levels (Tingkat)
CREATE TABLE grade_levels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    lms_level_id BIGINT NULL,
    lms_school_id BIGINT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uq_grade_levels_lms_level_id ON grade_levels (lms_level_id) WHERE lms_level_id IS NOT NULL;

-- Table: Classes
CREATE TABLE classes (
    id SERIAL PRIMARY KEY,
    lms_class_id BIGINT UNIQUE NOT NULL,
    lms_school_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_classes_school ON classes (lms_school_id);

-- Table: Class Students (Enrollment)
CREATE TABLE class_students (
    id SERIAL PRIMARY KEY,
    lms_class_id BIGINT NOT NULL,
    lms_user_id BIGINT NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (lms_class_id, lms_user_id)
);

-- Table: CBT Outbox
CREATE TABLE cbt_outbox (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    aggregate_type VARCHAR(100),
    aggregate_id BIGINT,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count INTEGER NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ,
    sent_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_cbt_outbox_status_next_attempt ON cbt_outbox (status, next_attempt_at, id);

-- =============================================
-- CONTENT TABLES
-- =============================================

-- Table: Materials (Materi)
-- Central content unit for an Exam/Test
CREATE TABLE materials (
    id SERIAL PRIMARY KEY,
    subject_id INTEGER NOT NULL REFERENCES subjects(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    grade_level_id INTEGER NOT NULL REFERENCES grade_levels(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    title VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    default_duration_minutes INTEGER NOT NULL DEFAULT 60,
    default_question_count INTEGER NOT NULL DEFAULT 20,
    
    -- Integration References
    lms_module_id BIGINT NULL,          -- If linked to an LMS Module
    lms_book_id BIGINT NULL,            -- [NEW] If linked to an LMS Book (Publisher Content)
    lms_teacher_material_id BIGINT NULL,-- [NEW] If derived from a Teacher's Material Collection
    lms_class_id BIGINT NULL,           -- Specific class assignment context
    
    owner_user_id INTEGER NULL,         -- ID of the creator (Teacher or Sync/Admin)
    school_id BIGINT NULL,
    labels JSONB NOT NULL DEFAULT '[]'::jsonb,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (subject_id, grade_level_id, title) -- Constraint might be too strict, consider removing UNIQUE on title later
);

CREATE INDEX idx_materials_subject ON materials (subject_id);
CREATE INDEX idx_materials_level ON materials (grade_level_id);
CREATE UNIQUE INDEX uq_materials_lms_module_id ON materials (lms_module_id) WHERE lms_module_id IS NOT NULL;
CREATE INDEX idx_materials_lms_book ON materials (lms_book_id);

-- Table: Questions (Soal)
-- Supports Multiple Choice and Essay. DragDrop is separated due to complexity.
CREATE TABLE questions (
    id SERIAL PRIMARY KEY,
    material_id INTEGER NOT NULL REFERENCES materials(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    lms_asset_id BIGINT,
    grade_level_id INTEGER NOT NULL REFERENCES grade_levels(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    
    question_text TEXT NOT NULL,
    question_type question_type_enum NOT NULL DEFAULT 'multiple_choice',
    
    -- Multiple Choice Fields
    option_a VARCHAR(500) NULL,
    option_b VARCHAR(500) NULL,
    option_c VARCHAR(500) NULL,
    option_d VARCHAR(500) NULL,
    correct_answer CHAR(1) NULL, -- 'A', 'B', 'C', 'D'
    
    -- Essay Fields [NEW]
    essay_answer_key TEXT NULL, -- Optional keywords or model answer for teacher reference
    
    explanation TEXT NULL, -- Pembahasan
    image_path VARCHAR(255) NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_questions_material ON questions (material_id);

-- Table: Question Images (Soal Gambar)
CREATE TABLE question_images (
    id SERIAL PRIMARY KEY,
    question_id INTEGER NOT NULL REFERENCES questions(id) ON DELETE CASCADE ON UPDATE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size INTEGER NOT NULL,
    mime_type VARCHAR(50) NOT NULL,
    order_no SMALLINT NOT NULL DEFAULT 1,
    caption VARCHAR(255) NULL,
    cloud_id VARCHAR(255) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table: Drag & Drop Questions (Soal Drag Drop)
CREATE TABLE drag_drop_questions (
    id SERIAL PRIMARY KEY,
    material_id INTEGER NOT NULL REFERENCES materials(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    grade_level_id INTEGER NOT NULL REFERENCES grade_levels(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    question_text TEXT NOT NULL,
    drag_type drag_type_enum NOT NULL,
    explanation TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table: Drag Drop Images
CREATE TABLE drag_drop_images (
    id SERIAL PRIMARY KEY,
    drag_drop_question_id INTEGER NOT NULL REFERENCES drag_drop_questions(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    order_no SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table: Drag Items
CREATE TABLE drag_items (
    id SERIAL PRIMARY KEY,
    drag_drop_question_id INTEGER NOT NULL REFERENCES drag_drop_questions(id) ON DELETE CASCADE ON UPDATE CASCADE,
    label VARCHAR(255) NOT NULL,
    image_url VARCHAR(500) NULL,
    order_no SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table: Drag Slots
CREATE TABLE drag_slots (
    id SERIAL PRIMARY KEY,
    drag_drop_question_id INTEGER NOT NULL REFERENCES drag_drop_questions(id) ON DELETE CASCADE ON UPDATE CASCADE,
    label VARCHAR(255) NOT NULL,
    image_url VARCHAR(500) NULL,
    order_no SMALLINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Table: Drag Correct Answers
CREATE TABLE drag_correct_answers (
    id SERIAL PRIMARY KEY,
    drag_item_id INTEGER NOT NULL REFERENCES drag_items(id) ON DELETE CASCADE ON UPDATE CASCADE,
    drag_slot_id INTEGER NOT NULL REFERENCES drag_slots(id) ON DELETE CASCADE ON UPDATE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (drag_item_id, drag_slot_id)
);

-- =============================================
-- EXAM SESSION TABLES
-- =============================================

-- Table: Exam Sessions (Test Session)
CREATE TABLE exam_sessions (
    id SERIAL PRIMARY KEY,
    session_token VARCHAR(64) NOT NULL UNIQUE,
    student_name VARCHAR(100) NOT NULL,
    grade_level_id INTEGER NOT NULL REFERENCES grade_levels(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    subject_id INTEGER NOT NULL REFERENCES subjects(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    user_id INTEGER NULL REFERENCES users(id) ON DELETE SET NULL ON UPDATE CASCADE,
    
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMPTZ NULL,
    duration_minutes INTEGER NOT NULL,
    
    final_score DECIMAL(5,2) NULL, -- nilai_akhir
    total_correct INTEGER NULL,
    total_questions INTEGER NULL,
    
    status exam_session_status_enum NOT NULL DEFAULT 'ongoing',
    lms_assignment_id BIGINT NULL,
    
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_exam_sessions_user ON exam_sessions (user_id);
CREATE INDEX idx_exam_sessions_status ON exam_sessions (status);

-- Table: Exam Session Questions (Unified Link)
CREATE TABLE exam_session_questions (
    id SERIAL PRIMARY KEY,
    exam_session_id INTEGER NOT NULL REFERENCES exam_sessions(id) ON DELETE CASCADE ON UPDATE CASCADE,
    question_id INTEGER NULL REFERENCES questions(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    drag_drop_question_id INTEGER NULL REFERENCES drag_drop_questions(id) ON DELETE RESTRICT ON UPDATE CASCADE,
    question_type question_type_enum NOT NULL DEFAULT 'multiple_choice',
    order_no SMALLINT NOT NULL,
    UNIQUE (exam_session_id, order_no)
);

-- Table: Student Answers (Jawaban Siswa)
CREATE TABLE student_answers (
    id SERIAL PRIMARY KEY,
    exam_session_question_id INTEGER NOT NULL REFERENCES exam_session_questions(id) ON DELETE CASCADE ON UPDATE CASCADE,
    
    question_type question_type_enum NOT NULL DEFAULT 'multiple_choice',
    
    -- Multiple Choice Answer
    selected_option CHAR(1) NULL,
    
    -- Drag Drop Answer
    drag_drop_answer JSONB NULL,
    
    -- Essay Answer [NEW]
    essay_answer_text TEXT NULL,
    essay_score DECIMAL(5,2) DEFAULT 0, -- Score for this specific essay
    teacher_feedback TEXT NULL,
    
    is_correct BOOLEAN NOT NULL DEFAULT FALSE, -- For auto-grading (MC/Drag)
    answered_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (exam_session_question_id)
);

-- =============================================
-- USER LIMITS
-- =============================================

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

CREATE TABLE user_limit_usage (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    limit_type VARCHAR(100) NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource_id INTEGER NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- =============================================
-- TRIGGER FUNCTION
-- =============================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_subjects_updated_at BEFORE UPDATE ON subjects FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_grade_levels_updated_at BEFORE UPDATE ON grade_levels FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_classes_updated_at BEFORE UPDATE ON classes FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_materials_updated_at BEFORE UPDATE ON materials FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_questions_updated_at BEFORE UPDATE ON questions FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_exam_sessions_updated_at BEFORE UPDATE ON exam_sessions FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
