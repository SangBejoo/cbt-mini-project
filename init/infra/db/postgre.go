package db

import (
	"cbt-test-mini-project/init/config"
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

func OpenSQL(cfgMain config.Main) (db *sql.DB, err error) {
	cfg := cfgMain.Database
	db, err = sql.Open("postgres", cfg.DSN)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	configureConnectionPool(db, cfg)

	// Test the connection
	if err = db.Ping(); err != nil {
		return nil, err
	}

	if err = ensureLegacyCompatibilityViews(context.Background(), db); err != nil {
		return nil, err
	}

	return db, nil
}

func ensureLegacyCompatibilityViews(ctx context.Context, db *sql.DB) error {
	var hasEnglishSchema bool
	err := db.QueryRowContext(ctx, `
		SELECT (
			to_regclass('public.subjects') IS NOT NULL AND
			to_regclass('public.grade_levels') IS NOT NULL AND
			to_regclass('public.materials') IS NOT NULL AND
			to_regclass('public.questions') IS NOT NULL AND
			to_regclass('public.exam_sessions') IS NOT NULL
		)
	`).Scan(&hasEnglishSchema)
	if err != nil {
		return fmt.Errorf("failed checking english schema availability: %w", err)
	}
	if !hasEnglishSchema {
		return nil
	}

	type compatibilityView struct {
		name string
		sql  string
	}

	views := []compatibilityView{
		{
			name: "mata_pelajaran",
			sql: `CREATE OR REPLACE VIEW mata_pelajaran AS
				SELECT id, name AS nama, is_active, lms_subject_id, lms_school_id, lms_class_id, created_at, updated_at
				FROM subjects`,
		},
		{
			name: "tingkat",
			sql: `CREATE OR REPLACE VIEW tingkat AS
				SELECT id, name AS nama, is_active, lms_level_id, lms_school_id, created_at, updated_at
				FROM grade_levels`,
		},
		{
			name: "materi",
			sql: `CREATE OR REPLACE VIEW materi AS
				SELECT
					id,
					subject_id AS id_mata_pelajaran,
					grade_level_id AS id_tingkat,
					title AS nama,
					is_active,
					default_duration_minutes AS default_durasi_menit,
					default_question_count AS default_jumlah_soal,
					lms_module_id,
					lms_book_id,
					lms_teacher_material_id,
					lms_class_id,
					owner_user_id,
					school_id,
					labels,
					created_at,
					updated_at
				FROM materials`,
		},
		{
			name: "soal",
			sql: `CREATE OR REPLACE VIEW soal AS
				SELECT
					id,
					material_id AS id_materi,
					lms_asset_id,
					grade_level_id AS id_tingkat,
					question_text AS pertanyaan,
					question_type,
					option_a AS opsi_a,
					option_b AS opsi_b,
					option_c AS opsi_c,
					option_d AS opsi_d,
					correct_answer AS jawaban_benar,
					essay_answer_key AS jawaban_essay_key,
					explanation AS pembahasan,
					image_path,
					is_active,
					lms_class_id,
					created_at,
					updated_at
				FROM questions`,
		},
		{
			name: "soal_gambar",
			sql: `CREATE OR REPLACE VIEW soal_gambar AS
				SELECT
					id,
					question_id AS id_soal,
					file_name AS nama_file,
					file_path,
					file_size,
					mime_type,
					order_no AS urutan,
					caption AS keterangan,
					cloud_id,
					public_id,
					created_at
				FROM question_images`,
		},
		{
			name: "soal_drag_drop",
			sql: `CREATE OR REPLACE VIEW soal_drag_drop AS
				SELECT
					id,
					material_id AS id_materi,
					grade_level_id AS id_tingkat,
					question_text AS pertanyaan,
					drag_type,
					explanation AS pembahasan,
					is_active,
					lms_class_id,
					created_at,
					updated_at
				FROM drag_drop_questions`,
		},
		{
			name: "soal_drag_drop_gambar",
			sql: `CREATE OR REPLACE VIEW soal_drag_drop_gambar AS
				SELECT
					id,
					drag_drop_question_id AS id_soal_drag_drop,
					file_name AS nama_file,
					file_path,
					file_size,
					mime_type,
					order_no AS urutan,
					caption AS keterangan,
					cloud_id,
					public_id,
					created_at
				FROM drag_drop_images`,
		},
		{
			name: "drag_item",
			sql: `CREATE OR REPLACE VIEW drag_item AS
				SELECT id, drag_drop_question_id AS id_soal_drag_drop, label, image_url, order_no AS urutan, created_at
				FROM drag_items`,
		},
		{
			name: "drag_slot",
			sql: `CREATE OR REPLACE VIEW drag_slot AS
				SELECT id, drag_drop_question_id AS id_soal_drag_drop, label, image_url, order_no AS urutan, created_at
				FROM drag_slots`,
		},
		{
			name: "drag_correct_answer",
			sql: `CREATE OR REPLACE VIEW drag_correct_answer AS
				SELECT id, drag_item_id AS id_drag_item, drag_slot_id AS id_drag_slot, created_at
				FROM drag_correct_answers`,
		},
		{
			name: "test_session",
			sql: `CREATE OR REPLACE VIEW test_session AS
				SELECT
					id,
					session_token,
					student_name AS nama_peserta,
					grade_level_id AS id_tingkat,
					subject_id AS id_mata_pelajaran,
					user_id,
					started_at AS waktu_mulai,
					finished_at AS waktu_selesai,
					duration_minutes AS durasi_menit,
					final_score AS nilai_akhir,
					total_correct AS jumlah_benar,
					total_questions AS total_soal,
					status,
					lms_assignment_id,
					lms_class_id,
					created_at,
					updated_at
				FROM exam_sessions`,
		},
		{
			name: "test_session_soal",
			sql: `CREATE OR REPLACE VIEW test_session_soal AS
				SELECT
					id,
					exam_session_id AS id_test_session,
					question_id AS id_soal,
					drag_drop_question_id AS id_soal_drag_drop,
					question_type,
					order_no AS nomor_urut
				FROM exam_session_questions`,
		},
		{
			name: "jawaban_siswa",
			sql: `CREATE OR REPLACE VIEW jawaban_siswa AS
				SELECT
					id,
					exam_session_question_id AS id_test_session_soal,
					selected_option AS jawaban_dipilih,
					is_correct,
					answered_at AS dijawab_pada,
					question_type,
					drag_drop_answer AS jawaban_drag_drop,
					essay_answer_text AS jawaban_essay,
					essay_score AS nilai_essay,
					teacher_feedback AS feedback_teacher
				FROM student_answers`,
		},
	}

	for _, view := range views {
		var kind sql.NullString
		scanErr := db.QueryRowContext(ctx, `
			SELECT c.relkind::text
			FROM pg_class c
			JOIN pg_namespace n ON n.oid = c.relnamespace
			WHERE n.nspname = 'public' AND c.relname = $1
		`, view.name).Scan(&kind)
		if scanErr != nil && scanErr != sql.ErrNoRows {
			return fmt.Errorf("failed checking relation %s: %w", view.name, scanErr)
		}

		if scanErr == nil && kind.Valid && kind.String != "v" {
			continue
		}

		if _, execErr := db.ExecContext(ctx, view.sql); execErr != nil {
			return fmt.Errorf("failed creating compatibility view %s: %w", view.name, execErr)
		}
	}

	return nil
}

func configureConnectionPool(sqlDB *sql.DB, cfg config.Database) {
	// Set maximum number of open connections to the database
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	// Set maximum number of idle connections in the pool
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	// Set maximum amount of time a connection may be reused
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	// Warm up minimum idle connections
	if cfg.MinIdleConns > 0 {
		warmUpConnections(sqlDB, cfg.MinIdleConns)
	}
}

func warmUpConnections(sqlDB *sql.DB, minIdleConns int) {
	// Create minimum idle connections by pinging the database
	conns := make([]*sql.Conn, minIdleConns)
	defer func() {
		for _, conn := range conns {
			if conn != nil {
				conn.Close()
			}
		}
	}()

	for i := 0; i < minIdleConns; i++ {
		conn, err := sqlDB.Conn(context.Background())
		if err != nil {
			// Log error but don't fail startup
			continue
		}
		conns[i] = conn

		// Ping to ensure connection is valid
		if err := conn.PingContext(context.Background()); err != nil {
			conn.Close()
			conns[i] = nil
		}
	}
}
