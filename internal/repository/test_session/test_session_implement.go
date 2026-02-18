package test_session

import (
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/event/contracts"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// testSessionRepositoryImpl implements TestSessionRepository
type testSessionRepositoryImpl struct {
	db *sql.DB
}

// NewTestSessionRepository creates a new TestSessionRepository instance
func NewTestSessionRepository(db *sql.DB) TestSessionRepository {
	return &testSessionRepositoryImpl{db: db}
}

// Create a new test session
func (r *testSessionRepositoryImpl) Create(session *entity.TestSession) error {
	query := `
		INSERT INTO test_session (session_token, user_id, nama_peserta, id_tingkat, id_mata_pelajaran, waktu_mulai, waktu_selesai, durasi_menit, nilai_akhir, jumlah_benar, total_soal, status, lms_assignment_id, lms_class_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id`
	return r.db.QueryRow(query, session.SessionToken, session.UserID, session.NamaPeserta, session.IDTingkat, session.IDMataPelajaran, session.WaktuMulai, session.WaktuSelesai, session.DurasiMenit, session.NilaiAkhir, session.JumlahBenar, session.TotalSoal, string(session.Status), session.LMSAssignmentID, session.LMSClassID).Scan(&session.ID)
}

// Get session by token
func (r *testSessionRepositoryImpl) GetByToken(token string) (*entity.TestSession, error) {
	var session entity.TestSession
	// Initialize User pointer to avoid nil pointer dereference during scan
	session.User = &entity.User{}

	query := `
		SELECT ts.id, ts.session_token, ts.user_id, ts.nama_peserta, ts.id_tingkat, ts.id_mata_pelajaran, ts.waktu_mulai, ts.waktu_selesai, ts.durasi_menit, ts.nilai_akhir, ts.jumlah_benar, ts.total_soal, ts.status, ts.lms_assignment_id,
		       mp.id, mp.nama, mp.is_active, mp.lms_subject_id, mp.lms_school_id, mp.lms_class_id,
		       t.id, t.nama, t.is_active, t.lms_level_id,
		       u.id, u.email, u.full_name, u.role, u.is_active, u.created_at, u.updated_at, u.lms_user_id
		FROM test_session ts
		JOIN mata_pelajaran mp ON ts.id_mata_pelajaran = mp.id
		JOIN tingkat t ON ts.id_tingkat = t.id
		JOIN users u ON ts.user_id = u.id
		WHERE ts.session_token = $1`
	err := r.db.QueryRow(query, token).Scan(
		&session.ID, &session.SessionToken, &session.UserID, &session.NamaPeserta, &session.IDTingkat, &session.IDMataPelajaran, &session.WaktuMulai, &session.WaktuSelesai, &session.DurasiMenit, &session.NilaiAkhir, &session.JumlahBenar, &session.TotalSoal, &session.Status, &session.LMSAssignmentID,
		&session.MataPelajaran.ID, &session.MataPelajaran.Nama, &session.MataPelajaran.IsActive, &session.MataPelajaran.LmsSubjectID, &session.MataPelajaran.LmsSchoolID, &session.MataPelajaran.LmsClassID,
		&session.Tingkat.ID, &session.Tingkat.Nama, &session.Tingkat.IsActive, &session.Tingkat.LmsLevelID,
		&session.User.ID, &session.User.Email, &session.User.Nama, &session.User.Role, &session.User.IsActive, &session.User.CreatedAt, &session.User.UpdatedAt, &session.User.LmsUserID,
	)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Update existing session
func (r *testSessionRepositoryImpl) Update(session *entity.TestSession) error {
	query := `
		UPDATE test_session
		SET session_token = $1, user_id = $2, nama_peserta = $3, id_tingkat = $4, id_mata_pelajaran = $5, waktu_mulai = $6, waktu_selesai = $7, durasi_menit = $8, nilai_akhir = $9, jumlah_benar = $10, total_soal = $11, status = $12, lms_assignment_id = $13, lms_class_id = $14
		WHERE id = $15`
	_, err := r.db.Exec(query, session.SessionToken, session.UserID, session.NamaPeserta, session.IDTingkat, session.IDMataPelajaran, session.WaktuMulai, session.WaktuSelesai, session.DurasiMenit, session.NilaiAkhir, session.JumlahBenar, session.TotalSoal, string(session.Status), session.LMSAssignmentID, session.LMSClassID, session.ID)
	return err
}

func (r *testSessionRepositoryImpl) CreateSessionForLMSUserIfMissing(lmsAssignmentID, lmsClassID, lmsUserID int64, idMataPelajaran, idTingkat, durasiMenit int, totalSoal *int, scheduledTime time.Time, status entity.TestStatus) (bool, error) {
	var userID int
	var namaPeserta string
	var isActive bool

	if err := r.db.QueryRow(`SELECT id, full_name, is_active FROM users WHERE lms_user_id = $1`, lmsUserID).Scan(&userID, &namaPeserta, &isActive); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("active membership resolution failed: local user not found for lms_user_id=%d", lmsUserID)
		}
		return false, err
	}
	if !isActive {
		return false, fmt.Errorf("active membership resolution failed: local user is inactive for lms_user_id=%d", lmsUserID)
	}

	var existingID int
	err := r.db.QueryRow(`SELECT id FROM test_session WHERE lms_assignment_id = $1 AND user_id = $2 LIMIT 1`, lmsAssignmentID, userID).Scan(&existingID)
	if err == nil {
		return false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	sessionToken := fmt.Sprintf("sync-%d-%d-%d", lmsAssignmentID, userID, time.Now().UnixNano())
	query := `
		INSERT INTO test_session (
			session_token, nama_peserta, id_tingkat, id_mata_pelajaran, user_id,
			waktu_mulai, durasi_menit, total_soal, status, lms_assignment_id, lms_class_id,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`
	_, err = r.db.Exec(query, sessionToken, namaPeserta, idTingkat, idMataPelajaran, userID, scheduledTime, durasiMenit, totalSoal, string(status), lmsAssignmentID, lmsClassID)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *testSessionRepositoryImpl) BackfillSessionsForJoinedStudent(lmsClassID, lmsUserID int64) (int, error) {
	query := `
		WITH active_assignments AS (
			SELECT DISTINCT lms_assignment_id, lms_class_id, id_tingkat, id_mata_pelajaran, waktu_mulai, durasi_menit, total_soal,
				CASE
					WHEN status = 'ongoing' THEN 'scheduled'
					ELSE status
				END AS status
			FROM test_session
			WHERE lms_class_id = $1
				AND lms_assignment_id IS NOT NULL
				AND status IN ('scheduled', 'ongoing')
		)
		INSERT INTO test_session (
			session_token, nama_peserta, id_tingkat, id_mata_pelajaran, user_id,
			waktu_mulai, durasi_menit, total_soal, status, lms_assignment_id, lms_class_id,
			created_at, updated_at
		)
		SELECT
			md5(random()::text || clock_timestamp()::text || aa.lms_assignment_id::text || u.id::text),
			COALESCE(NULLIF(u.full_name, ''), 'Siswa ' || u.id::text),
			aa.id_tingkat,
			aa.id_mata_pelajaran,
			u.id,
			aa.waktu_mulai,
			aa.durasi_menit,
			aa.total_soal,
			aa.status,
			aa.lms_assignment_id,
			aa.lms_class_id,
			CURRENT_TIMESTAMP,
			CURRENT_TIMESTAMP
		FROM active_assignments aa
		JOIN users u ON u.lms_user_id = $2 AND u.is_active = true
		WHERE NOT EXISTS (
			SELECT 1
			FROM test_session existing
			WHERE existing.lms_assignment_id = aa.lms_assignment_id
				AND existing.user_id = u.id
		)
	`
	result, err := r.db.Exec(query, lmsClassID, lmsUserID)
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rows), nil
}

func (r *testSessionRepositoryImpl) BackfillMissingSessions(lmsClassID *int64, lmsAssignmentID *int64) (int, error) {
	where := []string{
		"ts.lms_assignment_id IS NOT NULL",
		"ts.status IN ('scheduled', 'ongoing')",
	}
	args := make([]interface{}, 0)

	if lmsClassID != nil {
		args = append(args, *lmsClassID)
		where = append(where, fmt.Sprintf("ts.lms_class_id = $%d", len(args)))
	}
	if lmsAssignmentID != nil {
		args = append(args, *lmsAssignmentID)
		where = append(where, fmt.Sprintf("ts.lms_assignment_id = $%d", len(args)))
	}

	query := fmt.Sprintf(`
		WITH active_assignments AS (
			SELECT DISTINCT ts.lms_assignment_id, ts.lms_class_id, ts.id_tingkat, ts.id_mata_pelajaran, ts.waktu_mulai, ts.durasi_menit, ts.total_soal,
				CASE
					WHEN ts.status = 'ongoing' THEN 'scheduled'
					ELSE ts.status
				END AS status
			FROM test_session ts
			WHERE %s
		)
		INSERT INTO test_session (
			session_token, nama_peserta, id_tingkat, id_mata_pelajaran, user_id,
			waktu_mulai, durasi_menit, total_soal, status, lms_assignment_id, lms_class_id,
			created_at, updated_at
		)
		SELECT
			md5(random()::text || clock_timestamp()::text || aa.lms_assignment_id::text || u.id::text || cs.lms_class_id::text),
			COALESCE(NULLIF(u.full_name, ''), 'Siswa ' || u.id::text),
			aa.id_tingkat,
			aa.id_mata_pelajaran,
			u.id,
			aa.waktu_mulai,
			aa.durasi_menit,
			aa.total_soal,
			aa.status,
			aa.lms_assignment_id,
			aa.lms_class_id,
			CURRENT_TIMESTAMP,
			CURRENT_TIMESTAMP
		FROM active_assignments aa
		JOIN class_students cs ON cs.lms_class_id = aa.lms_class_id
		JOIN users u ON u.lms_user_id = cs.lms_user_id AND u.is_active = true
		WHERE NOT EXISTS (
			SELECT 1
			FROM test_session existing
			WHERE existing.lms_assignment_id = aa.lms_assignment_id
				AND existing.user_id = u.id
		)
	`, strings.Join(where, " AND "))

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(rows), nil
}

func (r *testSessionRepositoryImpl) UpdateScheduledSessionsByAssignment(lmsAssignmentID int64, lmsClassID int64, idMataPelajaran, idTingkat, durasiMenit int, totalSoal *int, scheduledTime time.Time) (int64, error) {
	query := `
		UPDATE test_session
		SET id_mata_pelajaran = $1,
			id_tingkat = $2,
			durasi_menit = $3,
			total_soal = $4,
			waktu_mulai = $5,
			lms_class_id = $6,
			updated_at = CURRENT_TIMESTAMP
		WHERE lms_assignment_id = $7
			AND status = 'scheduled'
	`
	result, err := r.db.Exec(query, idMataPelajaran, idTingkat, durasiMenit, totalSoal, scheduledTime, lmsClassID, lmsAssignmentID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *testSessionRepositoryImpl) DeleteSessionsByAssignment(lmsAssignmentID int64) (int64, error) {
	result, err := r.db.Exec(`DELETE FROM test_session WHERE lms_assignment_id = $1`, lmsAssignmentID)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Delete session by ID
func (r *testSessionRepositoryImpl) Delete(id int) error {
	query := `DELETE FROM test_session WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// Complete session
func (r *testSessionRepositoryImpl) CompleteSession(token string, waktuSelesai time.Time, nilaiAkhir *float64, jumlahBenar, totalSoal *int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	query := `
		UPDATE test_session
		SET waktu_selesai = $1, nilai_akhir = $2, jumlah_benar = $3, total_soal = $4, status = $5, updated_at = $6
		WHERE session_token = $7
		  AND status IN ('ongoing', 'timeout')
		RETURNING id, lms_assignment_id, lms_class_id, user_id`

	var sessionID int
	var lmsAssignmentID sql.NullInt64
	var lmsClassID sql.NullInt64
	var userID sql.NullInt64

	err = tx.QueryRow(
		query,
		waktuSelesai,
		nilaiAkhir,
		jumlahBenar,
		totalSoal,
		string(entity.TestStatusCompleted),
		time.Now(),
		token,
	).Scan(&sessionID, &lmsAssignmentID, &lmsClassID, &userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("session is already completed")
		}
		return err
	}

	if lmsAssignmentID.Valid && lmsClassID.Valid && userID.Valid && nilaiAkhir != nil && jumlahBenar != nil && totalSoal != nil {
		var lmsUserID sql.NullInt64
		userQuery := `SELECT lms_user_id FROM users WHERE id = $1 AND is_active = true`
		if err := tx.QueryRow(userQuery, userID.Int64).Scan(&lmsUserID); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		} else if lmsUserID.Valid {
			payload := contracts.ExamResultPayload{
				SessionID:       sessionID,
				LMSAssignmentID: lmsAssignmentID.Int64,
				LMSUserID:       lmsUserID.Int64,
				LMSClassID:      lmsClassID.Int64,
				Score:           *nilaiAkhir,
				CorrectCount:    *jumlahBenar,
				TotalCount:      *totalSoal,
				CompletedAt:     waktuSelesai.UTC().Format(time.RFC3339),
			}

			payloadJSON, marshalErr := json.Marshal(payload)
			if marshalErr != nil {
				return marshalErr
			}

			outboxQuery := `
				INSERT INTO cbt_outbox (event_type, aggregate_type, aggregate_id, payload, status, retry_count, created_at, updated_at)
				VALUES ($1, $2, $3, $4::jsonb, 'pending', 0, NOW(), NOW())`
			if _, err := tx.Exec(outboxQuery, string(contracts.ExamResultCompleted), "test_session", sessionID, string(payloadJSON)); err != nil {
				return err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true

	return nil
}

// UpdateSessionStatus updates only the status of a session
func (r *testSessionRepositoryImpl) UpdateSessionStatus(token string, status entity.TestStatus) error {
	query := `UPDATE test_session SET status = $1, updated_at = $2 WHERE session_token = $3`
	_, err := r.db.Exec(query, string(status), time.Now(), token)
	return err
}

// List sessions with filters
func (r *testSessionRepositoryImpl) List(tingkatan, idMataPelajaran *int, status *entity.TestStatus, limit, offset int) ([]entity.TestSession, int, error) {
	var sessions []entity.TestSession
	var total int

	// Build count query
	countQuery := `SELECT COUNT(*) FROM test_session ts`
	var countArgs []interface{}
	var countConditions []string

	if tingkatan != nil {
		countConditions = append(countConditions, "ts.id_tingkat = $"+fmt.Sprintf("%d", len(countArgs)+1))
		countArgs = append(countArgs, *tingkatan)
	}
	if idMataPelajaran != nil {
		countConditions = append(countConditions, "ts.id_mata_pelajaran = $"+fmt.Sprintf("%d", len(countArgs)+1))
		countArgs = append(countArgs, *idMataPelajaran)
	}
	if status != nil {
		countConditions = append(countConditions, "ts.status = $"+fmt.Sprintf("%d", len(countArgs)+1))
		countArgs = append(countArgs, string(*status))
	}

	if len(countConditions) > 0 {
		countQuery += " WHERE " + strings.Join(countConditions, " AND ")
	}

	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Build data query with JOINs
	dataQuery := `
		SELECT ts.id, ts.session_token, ts.user_id, ts.nama_peserta, ts.id_tingkat, ts.id_mata_pelajaran, ts.waktu_mulai, ts.waktu_selesai, ts.durasi_menit, ts.nilai_akhir, ts.jumlah_benar, ts.total_soal, ts.status, ts.lms_assignment_id,
		       mp.id, mp.nama, mp.is_active, mp.lms_subject_id, mp.lms_school_id, mp.lms_class_id,
		       t.id, t.nama, t.is_active, t.lms_level_id,
		       u.id, u.email, u.full_name, u.role, u.is_active, u.created_at, u.updated_at, u.lms_user_id
		FROM test_session ts
		JOIN mata_pelajaran mp ON ts.id_mata_pelajaran = mp.id
		JOIN tingkat t ON ts.id_tingkat = t.id
		JOIN users u ON ts.user_id = u.id`
	var dataArgs []interface{}
	var dataConditions []string

	if tingkatan != nil {
		dataConditions = append(dataConditions, "ts.id_tingkat = $"+fmt.Sprintf("%d", len(dataArgs)+1))
		dataArgs = append(dataArgs, *tingkatan)
	}
	if idMataPelajaran != nil {
		dataConditions = append(dataConditions, "ts.id_mata_pelajaran = $"+fmt.Sprintf("%d", len(dataArgs)+1))
		dataArgs = append(dataArgs, *idMataPelajaran)
	}
	if status != nil {
		dataConditions = append(dataConditions, "ts.status = $"+fmt.Sprintf("%d", len(dataArgs)+1))
		dataArgs = append(dataArgs, string(*status))
	}

	if len(dataConditions) > 0 {
		dataQuery += " WHERE " + strings.Join(dataConditions, " AND ")
	}

	dataQuery += " ORDER BY ts.created_at DESC LIMIT $" + fmt.Sprintf("%d", len(dataArgs)+1) + " OFFSET $" + fmt.Sprintf("%d", len(dataArgs)+2)
	dataArgs = append(dataArgs, limit, offset)

	rows, err := r.db.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var session entity.TestSession
		// Initialize User pointer to avoid nil pointer dereference during scan
		session.User = &entity.User{}
		err := rows.Scan(
			&session.ID, &session.SessionToken, &session.UserID, &session.NamaPeserta, &session.IDTingkat, &session.IDMataPelajaran, &session.WaktuMulai, &session.WaktuSelesai, &session.DurasiMenit, &session.NilaiAkhir, &session.JumlahBenar, &session.TotalSoal, &session.Status, &session.LMSAssignmentID,
			&session.MataPelajaran.ID, &session.MataPelajaran.Nama, &session.MataPelajaran.IsActive, &session.MataPelajaran.LmsSubjectID, &session.MataPelajaran.LmsSchoolID, &session.MataPelajaran.LmsClassID,
			&session.Tingkat.ID, &session.Tingkat.Nama, &session.Tingkat.IsActive, &session.Tingkat.LmsLevelID,
			&session.User.ID, &session.User.Email, &session.User.Nama, &session.User.Role, &session.User.IsActive, &session.User.CreatedAt, &session.User.UpdatedAt, &session.User.LmsUserID,
		)
		if err != nil {
			return nil, 0, err
		}
		sessions = append(sessions, session)
	}

	return sessions, total, nil
}

func (r *testSessionRepositoryImpl) ListScheduledByUser(userID int, lmsClassID *int64, limit, offset int) ([]entity.TestSession, int, error) {
	var sessions []entity.TestSession
	var total int

	countQuery := `SELECT COUNT(*) FROM test_session ts WHERE ts.user_id = $1 AND ts.status = 'scheduled'::test_session_status_enum`
	countArgs := []interface{}{userID}
	if lmsClassID != nil {
		countQuery += " AND ts.lms_class_id = $2"
		countArgs = append(countArgs, *lmsClassID)
	}

	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	dataQuery := `
		SELECT ts.id, ts.session_token, ts.user_id, ts.nama_peserta, ts.id_tingkat, ts.id_mata_pelajaran, ts.waktu_mulai, ts.waktu_selesai, ts.durasi_menit, ts.nilai_akhir, ts.jumlah_benar, ts.total_soal, ts.status, ts.lms_assignment_id, ts.lms_class_id,
		       mp.id, mp.nama, mp.is_active, mp.lms_subject_id, mp.lms_school_id, mp.lms_class_id,
		       t.id, t.nama, t.is_active, t.lms_level_id,
		       u.id, u.email, u.full_name, u.role, u.is_active, u.created_at, u.updated_at, u.lms_user_id
		FROM test_session ts
		JOIN mata_pelajaran mp ON ts.id_mata_pelajaran = mp.id
		JOIN tingkat t ON ts.id_tingkat = t.id
		JOIN users u ON ts.user_id = u.id
		WHERE ts.user_id = $1 AND ts.status = 'scheduled'::test_session_status_enum`
	dataArgs := []interface{}{userID}
	if lmsClassID != nil {
		dataQuery += " AND ts.lms_class_id = $2"
		dataArgs = append(dataArgs, *lmsClassID)
	}

	dataQuery += " ORDER BY ts.waktu_mulai ASC LIMIT $" + fmt.Sprintf("%d", len(dataArgs)+1) + " OFFSET $" + fmt.Sprintf("%d", len(dataArgs)+2)
	dataArgs = append(dataArgs, limit, offset)

	rows, err := r.db.Query(dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var session entity.TestSession
		session.User = &entity.User{}
		err := rows.Scan(
			&session.ID, &session.SessionToken, &session.UserID, &session.NamaPeserta, &session.IDTingkat, &session.IDMataPelajaran, &session.WaktuMulai, &session.WaktuSelesai, &session.DurasiMenit, &session.NilaiAkhir, &session.JumlahBenar, &session.TotalSoal, &session.Status, &session.LMSAssignmentID, &session.LMSClassID,
			&session.MataPelajaran.ID, &session.MataPelajaran.Nama, &session.MataPelajaran.IsActive, &session.MataPelajaran.LmsSubjectID, &session.MataPelajaran.LmsSchoolID, &session.MataPelajaran.LmsClassID,
			&session.Tingkat.ID, &session.Tingkat.Nama, &session.Tingkat.IsActive, &session.Tingkat.LmsLevelID,
			&session.User.ID, &session.User.Email, &session.User.Nama, &session.User.Role, &session.User.IsActive, &session.User.CreatedAt, &session.User.UpdatedAt, &session.User.LmsUserID,
		)
		if err != nil {
			return nil, 0, err
		}
		sessions = append(sessions, session)
	}

	return sessions, total, nil
}

func (r *testSessionRepositoryImpl) StartScheduledSession(token string, startedAt time.Time) (bool, error) {
	query := `
		UPDATE test_session
		SET status = 'ongoing'::test_session_status_enum,
			waktu_mulai = $1,
			updated_at = $2
		WHERE session_token = $3
			AND status = 'scheduled'::test_session_status_enum
	`
	result, err := r.db.Exec(query, startedAt, time.Now(), token)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

// Get questions for session
func (r *testSessionRepositoryImpl) GetSessionQuestions(token string) ([]entity.TestSessionSoal, error) {
	query := `
		SELECT tss.id, tss.id_test_session, tss.question_type, tss.id_soal, tss.id_soal_drag_drop, tss.nomor_urut,
		       s.id, s.pertanyaan, s.question_type, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, s.jawaban_benar, s.jawaban_essay_key, s.id_materi,
		       m.id, m.nama, m.id_mata_pelajaran, m.id_tingkat
		FROM test_session_soal tss
		JOIN test_session ts ON tss.id_test_session = ts.id
		LEFT JOIN soal s ON tss.id_soal = s.id
		LEFT JOIN materi m ON s.id_materi = m.id
		WHERE ts.session_token = $1
		ORDER BY tss.nomor_urut`
	rows, err := r.db.Query(query, token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessionSoals []entity.TestSessionSoal
	for rows.Next() {
		var tss entity.TestSessionSoal
		var soal entity.Soal
		var materi entity.Materi

		// Use nullable types for LEFT JOIN columns
		var soalID, soalIDMateri sql.NullInt64
		var soalPertanyaan, soalQuestionType, soalOpsiA, soalOpsiB, soalOpsiC, soalOpsiD, soalJawabanBenar, soalJawabanEssayKey sql.NullString
		var materiID, materiIDMataPelajaran, materiIDTingkat sql.NullInt64
		var materiNama sql.NullString

		err := rows.Scan(
			&tss.ID, &tss.IDTestSession, &tss.QuestionType, &tss.IDSoal, &tss.IDSoalDragDrop, &tss.NomorUrut,
			&soalID, &soalPertanyaan, &soalQuestionType, &soalOpsiA, &soalOpsiB, &soalOpsiC, &soalOpsiD, &soalJawabanBenar, &soalJawabanEssayKey, &soalIDMateri,
			&materiID, &materiNama, &materiIDMataPelajaran, &materiIDTingkat,
		)
		if err != nil {
			return nil, err
		}

		// Populate soal if it exists (multiple choice question)
		if soalID.Valid {
			soal.ID = int(soalID.Int64)
			soal.Pertanyaan = soalPertanyaan.String
			soal.OpsiA = soalOpsiA.String
			soal.OpsiB = soalOpsiB.String
			soal.OpsiC = soalOpsiC.String
			soal.OpsiD = soalOpsiD.String
			soal.QuestionType = entity.QuestionType(soalQuestionType.String)
			soal.JawabanBenar = entity.JawabanOption(soalJawabanBenar.String)
			if soalJawabanEssayKey.Valid {
				soal.JawabanEssayKey = &soalJawabanEssayKey.String
			}
			if soalIDMateri.Valid {
				soal.IDMateri = int(soalIDMateri.Int64)
			}
			if materiID.Valid {
				materi.ID = int(materiID.Int64)
				materi.Nama = materiNama.String
				if materiIDMataPelajaran.Valid {
					materi.IDMataPelajaran = int(materiIDMataPelajaran.Int64)
				}
				if materiIDTingkat.Valid {
					materi.IDTingkat = int(materiIDTingkat.Int64)
				}
			}
			soal.Materi = materi
			tss.Soal = &soal
		}

		sessionSoals = append(sessionSoals, tss)
	}
	return sessionSoals, nil
}

// Get all questions for session with soal data
func (r *testSessionRepositoryImpl) GetAllQuestionsForSession(token string) ([]entity.TestSessionSoal, error) {
	query := `
		SELECT tss.id, tss.id_test_session, tss.question_type, tss.id_soal, tss.id_soal_drag_drop, tss.nomor_urut,
		       s.id, s.pertanyaan, s.question_type, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, s.jawaban_benar, s.jawaban_essay_key, s.id_materi,
		       m.id, m.nama, m.id_mata_pelajaran, m.id_tingkat, mp.id, mp.nama, mp.is_active, t.id, t.nama, t.is_active,
		       sdd.id, sdd.pertanyaan, sdd.id_materi
		FROM test_session_soal tss
		JOIN test_session ts ON tss.id_test_session = ts.id
		LEFT JOIN soal s ON tss.id_soal = s.id
		LEFT JOIN materi m ON s.id_materi = m.id
		LEFT JOIN mata_pelajaran mp ON m.id_mata_pelajaran = mp.id
		LEFT JOIN tingkat t ON m.id_tingkat = t.id
		LEFT JOIN soal_drag_drop sdd ON tss.id_soal_drag_drop = sdd.id
		WHERE ts.session_token = $1
		ORDER BY tss.nomor_urut`
	rows, err := r.db.Query(query, token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessionSoals []entity.TestSessionSoal
	for rows.Next() {
		var tss entity.TestSessionSoal
		var soal entity.Soal
		var materi entity.Materi
		var mataPelajaran entity.MataPelajaran
		var tingkat entity.Tingkat
		var soalDragDrop entity.SoalDragDrop

		// Use nullable types for LEFT JOIN columns
		var soalID, soalIDMateri sql.NullInt64
		var soalPertanyaan, soalQuestionType, soalOpsiA, soalOpsiB, soalOpsiC, soalOpsiD, soalJawabanBenar, soalJawabanEssayKey sql.NullString
		var materiID, materiIDMataPelajaran, materiIDTingkat sql.NullInt64
		var materiNama sql.NullString
		var mataPelajaranID sql.NullInt64
		var mataPelajaranNama sql.NullString
		var mataPelajaranIsActive sql.NullBool
		var tingkatID sql.NullInt64
		var tingkatNama sql.NullString
		var tingkatIsActive sql.NullBool
		var sddID, sddIDMateri sql.NullInt64
		var sddPertanyaan sql.NullString

		err := rows.Scan(
			&tss.ID, &tss.IDTestSession, &tss.QuestionType, &tss.IDSoal, &tss.IDSoalDragDrop, &tss.NomorUrut,
			&soalID, &soalPertanyaan, &soalQuestionType, &soalOpsiA, &soalOpsiB, &soalOpsiC, &soalOpsiD, &soalJawabanBenar, &soalJawabanEssayKey, &soalIDMateri,
			&materiID, &materiNama, &materiIDMataPelajaran, &materiIDTingkat, &mataPelajaranID, &mataPelajaranNama, &mataPelajaranIsActive, &tingkatID, &tingkatNama, &tingkatIsActive,
			&sddID, &sddPertanyaan, &sddIDMateri,
		)
		if err != nil {
			return nil, err
		}

		// Populate soal if it exists (multiple choice question)
		if soalID.Valid {
			soal.ID = int(soalID.Int64)
			soal.Pertanyaan = soalPertanyaan.String
			soal.OpsiA = soalOpsiA.String
			soal.OpsiB = soalOpsiB.String
			soal.OpsiC = soalOpsiC.String
			soal.OpsiD = soalOpsiD.String
			soal.QuestionType = entity.QuestionType(soalQuestionType.String)
			soal.JawabanBenar = entity.JawabanOption(soalJawabanBenar.String)
			if soalJawabanEssayKey.Valid {
				soal.JawabanEssayKey = &soalJawabanEssayKey.String
			}
			if soalIDMateri.Valid {
				soal.IDMateri = int(soalIDMateri.Int64)
			}
			if materiID.Valid {
				materi.ID = int(materiID.Int64)
				materi.Nama = materiNama.String
				if materiIDMataPelajaran.Valid {
					materi.IDMataPelajaran = int(materiIDMataPelajaran.Int64)
				}
				if materiIDTingkat.Valid {
					materi.IDTingkat = int(materiIDTingkat.Int64)
				}
				if mataPelajaranID.Valid {
					mataPelajaran.ID = int(mataPelajaranID.Int64)
					mataPelajaran.Nama = mataPelajaranNama.String
					mataPelajaran.IsActive = mataPelajaranIsActive.Bool
				}
				if tingkatID.Valid {
					tingkat.ID = int(tingkatID.Int64)
					tingkat.Nama = tingkatNama.String
					tingkat.IsActive = tingkatIsActive.Bool
				}
				materi.MataPelajaran = mataPelajaran
				materi.Tingkat = tingkat
			}
			soal.Materi = materi
			tss.Soal = &soal
		}

		// Populate soalDragDrop if it exists (drag-drop question)
		if sddID.Valid {
			soalDragDrop.ID = int(sddID.Int64)
			soalDragDrop.Pertanyaan = sddPertanyaan.String
			if sddIDMateri.Valid {
				soalDragDrop.IDMateri = int(sddIDMateri.Int64)
			}
			tss.SoalDragDrop = &soalDragDrop
		}

		sessionSoals = append(sessionSoals, tss)
	}
	return sessionSoals, nil
}

// Get single question by order
func (r *testSessionRepositoryImpl) GetQuestionByOrder(token string, nomorUrut int) (*entity.Soal, error) {
	query := `
		SELECT s.id, s.pertanyaan, s.question_type, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, s.jawaban_benar, s.jawaban_essay_key, s.id_materi,
		       m.id, m.nama, m.id_mata_pelajaran, m.id_tingkat,
		       mp.id, mp.nama, mp.is_active, mp.lms_subject_id, mp.lms_school_id, mp.lms_class_id,
		       t.id, t.nama, t.is_active, t.lms_level_id
		FROM soal s
		JOIN test_session_soal tss ON tss.id_soal = s.id
		JOIN test_session ts ON tss.id_test_session = ts.id
		JOIN materi m ON s.id_materi = m.id
		JOIN mata_pelajaran mp ON m.id_mata_pelajaran = mp.id
		JOIN tingkat t ON m.id_tingkat = t.id
		WHERE ts.session_token = $1 AND tss.nomor_urut = $2`
	var soal entity.Soal
	var materi entity.Materi
	var mataPelajaran entity.MataPelajaran
	var tingkat entity.Tingkat
	var soalQuestionType sql.NullString
	var jawabanEssayKey sql.NullString
	err := r.db.QueryRow(query, token, nomorUrut).Scan(
		&soal.ID, &soal.Pertanyaan, &soalQuestionType, &soal.OpsiA, &soal.OpsiB, &soal.OpsiC, &soal.OpsiD, &soal.JawabanBenar, &jawabanEssayKey, &soal.IDMateri,
		&materi.ID, &materi.Nama, &materi.IDMataPelajaran, &materi.IDTingkat,
		&mataPelajaran.ID, &mataPelajaran.Nama, &mataPelajaran.IsActive, &mataPelajaran.LmsSubjectID, &mataPelajaran.LmsSchoolID, &mataPelajaran.LmsClassID,
		&tingkat.ID, &tingkat.Nama, &tingkat.IsActive, &tingkat.LmsLevelID,
	)
	if err != nil {
		return nil, err
	}
	materi.MataPelajaran = mataPelajaran
	materi.Tingkat = tingkat
	soal.Materi = materi
	if soalQuestionType.Valid {
		soal.QuestionType = entity.QuestionType(soalQuestionType.String)
	}
	if jawabanEssayKey.Valid {
		soal.JawabanEssayKey = &jawabanEssayKey.String
	}
	return &soal, nil
}

// Submit answer
func (r *testSessionRepositoryImpl) SubmitAnswer(token string, nomorUrut int, jawaban entity.JawabanOption) error {
	// Find the TestSessionSoal
	var tss entity.TestSessionSoal
	var soal entity.Soal
	query := `
		SELECT tss.id, tss.id_test_session, tss.question_type, tss.id_soal, tss.id_soal_drag_drop, tss.nomor_urut,
		       s.id, s.pertanyaan, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, s.jawaban_benar, s.id_materi
		FROM test_session_soal tss
		JOIN test_session ts ON tss.id_test_session = ts.id
		JOIN soal s ON tss.id_soal = s.id
		WHERE ts.session_token = $1 AND tss.nomor_urut = $2`
	err := r.db.QueryRow(query, token, nomorUrut).Scan(
		&tss.ID, &tss.IDTestSession, &tss.QuestionType, &tss.IDSoal, &tss.IDSoalDragDrop, &tss.NomorUrut,
		&soal.ID, &soal.Pertanyaan, &soal.OpsiA, &soal.OpsiB, &soal.OpsiC, &soal.OpsiD, &soal.JawabanBenar, &soal.IDMateri,
	)
	if err != nil {
		return err
	}
	tss.Soal = &soal

	// Prepare the answer object
	if tss.QuestionType != entity.QuestionTypeMultipleChoice {
		return errors.New("this is not a multiple-choice question")
	}
	isCorrect := (jawaban == tss.Soal.JawabanBenar)
	newAnswer := entity.JawabanSiswa{
		IDTestSessionSoal: tss.ID,
		JawabanDipilih:    &jawaban,
		IsCorrect:         isCorrect,
		QuestionType:      entity.QuestionTypeMultipleChoice,
	}

	// Upsert: If exists, update answer and correctness; if not, create new.
	upsertQuery := `
		INSERT INTO jawaban_siswa (id_test_session_soal, jawaban_dipilih, is_correct, question_type, dijawab_pada)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id_test_session_soal)
		DO UPDATE SET jawaban_dipilih = EXCLUDED.jawaban_dipilih, is_correct = EXCLUDED.is_correct, dijawab_pada = EXCLUDED.dijawab_pada`
	_, err = r.db.Exec(upsertQuery, newAnswer.IDTestSessionSoal, newAnswer.JawabanDipilih, newAnswer.IsCorrect, string(newAnswer.QuestionType), time.Now())
	return err
}

// Clear answer
func (r *testSessionRepositoryImpl) ClearAnswer(token string, nomorUrut int) error {
	// Find the TestSessionSoal
	var tssID int
	query := `
		SELECT tss.id
		FROM test_session_soal tss
		JOIN test_session ts ON tss.id_test_session = ts.id
		WHERE ts.session_token = $1 AND tss.nomor_urut = $2`
	err := r.db.QueryRow(query, token, nomorUrut).Scan(&tssID)
	if err != nil {
		return err
	}

	// Delete the answer if exists
	deleteQuery := `DELETE FROM jawaban_siswa WHERE id_test_session_soal = $1`
	_, err = r.db.Exec(deleteQuery, tssID)
	return err
}

// Get answers for session
func (r *testSessionRepositoryImpl) GetSessionAnswers(token string) ([]entity.JawabanSiswa, error) {
	query := `
		SELECT js.id, js.id_test_session_soal, js.jawaban_dipilih, js.is_correct, js.question_type, js.dijawab_pada, js.jawaban_drag_drop, js.jawaban_essay, js.nilai_essay, js.feedback_teacher,
		       tss.id, tss.id_test_session, tss.question_type, tss.id_soal, tss.id_soal_drag_drop, tss.nomor_urut,
		       s.id, s.pertanyaan, s.question_type, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, s.jawaban_benar, s.jawaban_essay_key, s.id_materi
		FROM jawaban_siswa js
		JOIN test_session_soal tss ON js.id_test_session_soal = tss.id
		JOIN test_session ts ON tss.id_test_session = ts.id
		LEFT JOIN soal s ON tss.id_soal = s.id
		WHERE ts.session_token = $1`
	rows, err := r.db.Query(query, token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var answers []entity.JawabanSiswa
	for rows.Next() {
		var js entity.JawabanSiswa
		var tss entity.TestSessionSoal
		var soal entity.Soal

		// Use nullable types for LEFT JOIN soal columns
		var soalID, soalIDMateri sql.NullInt64
		var soalPertanyaan, soalQuestionType, soalOpsiA, soalOpsiB, soalOpsiC, soalOpsiD, soalJawabanBenar, soalJawabanEssayKey sql.NullString
		var nilaiEssay sql.NullFloat64

		err := rows.Scan(
			&js.ID, &js.IDTestSessionSoal, &js.JawabanDipilih, &js.IsCorrect, &js.QuestionType, &js.DijawabPada, &js.JawabanDragDrop, &js.JawabanEssay, &nilaiEssay, &js.FeedbackTeacher,
			&tss.ID, &tss.IDTestSession, &tss.QuestionType, &tss.IDSoal, &tss.IDSoalDragDrop, &tss.NomorUrut,
			&soalID, &soalPertanyaan, &soalQuestionType, &soalOpsiA, &soalOpsiB, &soalOpsiC, &soalOpsiD, &soalJawabanBenar, &soalJawabanEssayKey, &soalIDMateri,
		)
		if err != nil {
			return nil, err
		}

		// Populate soal if it exists (multiple choice question)
		if soalID.Valid {
			soal.ID = int(soalID.Int64)
			soal.Pertanyaan = soalPertanyaan.String
			soal.OpsiA = soalOpsiA.String
			soal.OpsiB = soalOpsiB.String
			soal.OpsiC = soalOpsiC.String
			soal.OpsiD = soalOpsiD.String
			soal.QuestionType = entity.QuestionType(soalQuestionType.String)
			soal.JawabanBenar = entity.JawabanOption(soalJawabanBenar.String)
			if soalJawabanEssayKey.Valid {
				soal.JawabanEssayKey = &soalJawabanEssayKey.String
			}
			if soalIDMateri.Valid {
				soal.IDMateri = int(soalIDMateri.Int64)
			}
			tss.Soal = &soal
		}
		if nilaiEssay.Valid {
			v := nilaiEssay.Float64
			js.NilaiEssay = &v
		}

		js.TestSessionSoal = tss
		answers = append(answers, js)
	}
	return answers, nil
}

// Assign random questions to session
func (r *testSessionRepositoryImpl) AssignRandomQuestions(sessionID, idMataPelajaran, tingkatan, jumlahSoal int) error {
	// Get random soal IDs for the criteria - get questions for the mata_pelajaran and tingkat
	soalQuery := `
		SELECT s.id, s.question_type
		FROM soal s
		JOIN materi m ON s.id_materi = m.id
		WHERE m.id_mata_pelajaran = $1 AND m.id_tingkat = $2`
	soalRows, err := r.db.Query(soalQuery, idMataPelajaran, tingkatan)
	if err != nil {
		return err
	}
	defer soalRows.Close()

	// Combine all question IDs
	var allQuestionIDs []struct {
		ID           int
		QuestionType entity.QuestionType
	}

	var soalIDs []int
	for soalRows.Next() {
		var id int
		var questionType sql.NullString
		soalRows.Scan(&id, &questionType)
		if strings.EqualFold(questionType.String, string(entity.QuestionTypeEssay)) {
			allQuestionIDs = append(allQuestionIDs, struct {
				ID           int
				QuestionType entity.QuestionType
			}{ID: id, QuestionType: entity.QuestionTypeEssay})
		} else {
			soalIDs = append(soalIDs, id)
		}
	}

	// Get drag-drop question IDs for the same criteria
	dragDropQuery := `
		SELECT sdd.id
		FROM soal_drag_drop sdd
		JOIN materi m ON sdd.id_materi = m.id
		WHERE m.id_mata_pelajaran = $1 AND m.id_tingkat = $2 AND sdd.is_active = $3`
	dragDropRows, err := r.db.Query(dragDropQuery, idMataPelajaran, tingkatan, true)
	if err != nil {
		return err
	}
	defer dragDropRows.Close()

	var dragDropIDs []int
	for dragDropRows.Next() {
		var id int
		dragDropRows.Scan(&id)
		dragDropIDs = append(dragDropIDs, id)
	}

	// Add multiple choice questions
	for _, id := range soalIDs {
		allQuestionIDs = append(allQuestionIDs, struct {
			ID           int
			QuestionType entity.QuestionType
		}{ID: id, QuestionType: entity.QuestionTypeMultipleChoice})
	}

	// Add drag-drop questions
	for _, id := range dragDropIDs {
		allQuestionIDs = append(allQuestionIDs, struct {
			ID           int
			QuestionType entity.QuestionType
		}{ID: id, QuestionType: entity.QuestionTypeDragDrop})
	}

	if len(allQuestionIDs) == 0 {
		return errors.New("tidak ada soal yang tersedia untuk mata pelajaran dan tingkatan ini")
	}

	// Ambil jumlah soal yang tersedia atau yang diminta (mana yang lebih kecil)
	actualJumlahSoal := jumlahSoal
	if len(allQuestionIDs) < jumlahSoal {
		actualJumlahSoal = len(allQuestionIDs)
	}

	// Shuffle and select
	rand.Shuffle(len(allQuestionIDs), func(i, j int) { allQuestionIDs[i], allQuestionIDs[j] = allQuestionIDs[j], allQuestionIDs[i] })
	selectedQuestions := allQuestionIDs[:actualJumlahSoal]

	// Create TestSessionSoal entries
	for i, question := range selectedQuestions {
		switch question.QuestionType {
		case entity.QuestionTypeMultipleChoice, entity.QuestionTypeEssay:
			soalIDPtr := question.ID // Create a copy for pointer
			insertQuery := `
				INSERT INTO test_session_soal (id_test_session, question_type, id_soal, nomor_urut)
				VALUES ($1, $2, $3, $4)`
			_, err := r.db.Exec(insertQuery, sessionID, string(question.QuestionType), soalIDPtr, i+1)
			if err != nil {
				return err
			}
		case entity.QuestionTypeDragDrop:
			soalDragDropIDPtr := question.ID // Create a copy for pointer
			insertQuery := `
				INSERT INTO test_session_soal (id_test_session, question_type, id_soal_drag_drop, nomor_urut)
				VALUES ($1, $2, $3, $4)`
			_, err := r.db.Exec(insertQuery, sessionID, string(entity.QuestionTypeDragDrop), soalDragDropIDPtr, i+1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateUnansweredRecord creates a record for unanswered question with NULL jawaban_dipilih
func (r *testSessionRepositoryImpl) CreateUnansweredRecord(sessionSoalID, testSessionID int) error {
	insertQuery := `
		INSERT INTO jawaban_siswa (id_test_session_soal, jawaban_dipilih, is_correct, question_type)
		SELECT tss.id, NULL, false, tss.question_type
		FROM test_session_soal tss
		WHERE tss.id = $1`
	_, err := r.db.Exec(insertQuery, sessionSoalID)
	return err
}

// GetTestSessionSoalByOrder gets TestSessionSoal by token and nomor_urut
func (r *testSessionRepositoryImpl) GetTestSessionSoalByOrder(token string, nomorUrut int) (*entity.TestSessionSoal, error) {
	query := `
		SELECT tss.id, tss.id_test_session, tss.question_type, tss.id_soal, tss.id_soal_drag_drop, tss.nomor_urut,
		       s.id, s.pertanyaan, s.question_type, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, s.jawaban_benar, s.jawaban_essay_key, s.id_materi,
		       sdd.id, sdd.pertanyaan, sdd.id_materi
		FROM test_session_soal tss
		JOIN test_session ts ON tss.id_test_session = ts.id
		LEFT JOIN soal s ON tss.id_soal = s.id
		LEFT JOIN soal_drag_drop sdd ON tss.id_soal_drag_drop = sdd.id
		WHERE ts.session_token = $1 AND tss.nomor_urut = $2`
	var tss entity.TestSessionSoal
	var soal entity.Soal
	var soalDragDrop entity.SoalDragDrop

	// Use nullable types for LEFT JOIN columns
	var soalID, soalIDMateri sql.NullInt64
	var soalPertanyaan, soalQuestionType, soalOpsiA, soalOpsiB, soalOpsiC, soalOpsiD, soalJawabanBenar, soalJawabanEssayKey sql.NullString
	var sddID, sddIDMateri sql.NullInt64
	var sddPertanyaan sql.NullString

	err := r.db.QueryRow(query, token, nomorUrut).Scan(
		&tss.ID, &tss.IDTestSession, &tss.QuestionType, &tss.IDSoal, &tss.IDSoalDragDrop, &tss.NomorUrut,
		&soalID, &soalPertanyaan, &soalQuestionType, &soalOpsiA, &soalOpsiB, &soalOpsiC, &soalOpsiD, &soalJawabanBenar, &soalJawabanEssayKey, &soalIDMateri,
		&sddID, &sddPertanyaan, &sddIDMateri,
	)
	if err != nil {
		return nil, err
	}

	// Populate soal if it exists (multiple choice question)
	if soalID.Valid {
		soal.ID = int(soalID.Int64)
		soal.Pertanyaan = soalPertanyaan.String
		soal.OpsiA = soalOpsiA.String
		soal.OpsiB = soalOpsiB.String
		soal.OpsiC = soalOpsiC.String
		soal.OpsiD = soalOpsiD.String
		soal.QuestionType = entity.QuestionType(soalQuestionType.String)
		soal.JawabanBenar = entity.JawabanOption(soalJawabanBenar.String)
		if soalJawabanEssayKey.Valid {
			soal.JawabanEssayKey = &soalJawabanEssayKey.String
		}
		if soalIDMateri.Valid {
			soal.IDMateri = int(soalIDMateri.Int64)
		}
		tss.Soal = &soal
	}

	// Populate soalDragDrop if it exists (drag-drop question)
	if sddID.Valid {
		soalDragDrop.ID = int(sddID.Int64)
		soalDragDrop.Pertanyaan = sddPertanyaan.String
		if sddIDMateri.Valid {
			soalDragDrop.IDMateri = int(sddIDMateri.Int64)
		}
		tss.SoalDragDrop = &soalDragDrop
	}

	return &tss, nil
}

// SubmitDragDropAnswer submits a drag-drop answer
func (r *testSessionRepositoryImpl) SubmitDragDropAnswer(token string, nomorUrut int, answer map[int]int, isCorrect bool) error {
	// Find the TestSessionSoal
	tss, err := r.GetTestSessionSoalByOrder(token, nomorUrut)
	if err != nil {
		return err
	}

	// Prepare the new answer object
	newAnswer := entity.JawabanSiswa{
		IDTestSessionSoal: tss.ID,
		QuestionType:      entity.QuestionTypeDragDrop,
		IsCorrect:         isCorrect,
		JawabanDipilih:    nil, // Explicitly nil for DragDrop
	}
	newAnswer.SetDragDropAnswer(answer)

	// Upsert: If exists, update answer and correctness; if not, create new.
	upsertQuery := `
		INSERT INTO jawaban_siswa (id_test_session_soal, question_type, is_correct, jawaban_drag_drop, dijawab_pada)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id_test_session_soal)
		DO UPDATE SET jawaban_drag_drop = EXCLUDED.jawaban_drag_drop, question_type = EXCLUDED.question_type, is_correct = EXCLUDED.is_correct, dijawab_pada = EXCLUDED.dijawab_pada`
	_, err = r.db.Exec(upsertQuery, newAnswer.IDTestSessionSoal, string(newAnswer.QuestionType), newAnswer.IsCorrect, newAnswer.JawabanDragDrop, time.Now())
	return err
}

func (r *testSessionRepositoryImpl) SubmitEssayAnswer(token string, nomorUrut int, jawabanEssay string) error {
	tss, err := r.GetTestSessionSoalByOrder(token, nomorUrut)
	if err != nil {
		return err
	}
	if tss.QuestionType != entity.QuestionTypeEssay {
		return errors.New("this is not an essay question")
	}

	upsertQuery := `
		INSERT INTO jawaban_siswa (id_test_session_soal, question_type, is_correct, jawaban_essay, dijawab_pada)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id_test_session_soal)
		DO UPDATE SET question_type = EXCLUDED.question_type, jawaban_essay = EXCLUDED.jawaban_essay, dijawab_pada = EXCLUDED.dijawab_pada`
	_, err = r.db.Exec(upsertQuery, tss.ID, string(entity.QuestionTypeEssay), false, strings.TrimSpace(jawabanEssay), time.Now())
	return err
}

func (r *testSessionRepositoryImpl) HasEssayQuestions(token string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM test_session_soal tss
		JOIN test_session ts ON tss.id_test_session = ts.id
		WHERE ts.session_token = $1 AND tss.question_type = $2`
	var count int
	err := r.db.QueryRow(query, token, string(entity.QuestionTypeEssay)).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *testSessionRepositoryImpl) GradeEssayAnswer(answerID int, score float64, feedback string) (string, error) {
	if score < 0 || score > 100 {
		return "", errors.New("score must be between 0 and 100")
	}

	isCorrect := score >= 60
	query := `
		UPDATE jawaban_siswa
		SET nilai_essay = $1,
		    feedback_teacher = $2,
		    is_correct = $3
		WHERE id = $4 AND question_type = $5`
	res, err := r.db.Exec(query, score, strings.TrimSpace(feedback), isCorrect, answerID, string(entity.QuestionTypeEssay))
	if err != nil {
		return "", err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return "", err
	}
	if affected == 0 {
		return "", errors.New("essay answer not found")
	}

	var token string
	tokenQuery := `
		SELECT ts.session_token
		FROM jawaban_siswa js
		JOIN test_session_soal tss ON js.id_test_session_soal = tss.id
		JOIN test_session ts ON tss.id_test_session = ts.id
		WHERE js.id = $1`
	if err := r.db.QueryRow(tokenQuery, answerID).Scan(&token); err != nil {
		return "", err
	}

	recalcQuery := `
		WITH score_calc AS (
			SELECT ts.id AS session_id,
				COUNT(tss.id)::int AS total_questions,
				SUM(
					CASE
						WHEN tss.question_type = 'essay' THEN COALESCE(js.nilai_essay, 0) / 100.0
						WHEN js.is_correct THEN 1.0
						ELSE 0.0
					END
				) AS total_points,
				SUM(
					CASE
						WHEN tss.question_type = 'essay' AND COALESCE(js.nilai_essay, 0) >= 60 THEN 1
						WHEN tss.question_type <> 'essay' AND js.is_correct THEN 1
						ELSE 0
					END
				)::int AS total_correct,
				SUM(CASE WHEN tss.question_type = 'essay' AND js.nilai_essay IS NULL THEN 1 ELSE 0 END)::int AS pending_essay
			FROM test_session ts
			JOIN test_session_soal tss ON tss.id_test_session = ts.id
			LEFT JOIN jawaban_siswa js ON js.id_test_session_soal = tss.id
			WHERE ts.session_token = $1
			GROUP BY ts.id
		)
		UPDATE test_session ts
		SET nilai_akhir = CASE WHEN sc.total_questions > 0 THEN (sc.total_points / sc.total_questions) * 100 ELSE 0 END,
		    jumlah_benar = sc.total_correct,
		    total_soal = sc.total_questions,
		    status = CASE WHEN sc.pending_essay > 0 THEN 'grading_in_progress' ELSE 'graded' END
		FROM score_calc sc
		WHERE ts.id = sc.session_id`
	_, err = r.db.Exec(recalcQuery, token)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetDragDropCorrectAnswers gets correct answers for a drag-drop question
func (r *testSessionRepositoryImpl) GetDragDropCorrectAnswers(soalDragDropID int) ([]entity.DragCorrectAnswer, error) {
	query := `
		SELECT dca.id, dca.id_drag_item, dca.id_drag_slot
		FROM drag_correct_answer dca
		JOIN drag_item di ON dca.id_drag_item = di.id
		WHERE di.id_soal_drag_drop = $1`
	rows, err := r.db.Query(query, soalDragDropID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var correctAnswers []entity.DragCorrectAnswer
	for rows.Next() {
		var dca entity.DragCorrectAnswer
		err := rows.Scan(&dca.ID, &dca.IDDragItem, &dca.IDDragSlot)
		if err != nil {
			return nil, err
		}
		correctAnswers = append(correctAnswers, dca)
	}
	return correctAnswers, nil
}

// GetSoalDragDropByID gets a drag-drop question by ID
func (r *testSessionRepositoryImpl) GetSoalDragDropByID(id int) (*entity.SoalDragDrop, error) {
	query := `
		SELECT sdd.id, sdd.pertanyaan, sdd.id_materi, sdd.is_active, sdd.created_at, sdd.updated_at,
		       m.id, m.nama, m.id_mata_pelajaran, m.id_tingkat
		FROM soal_drag_drop sdd
		JOIN materi m ON sdd.id_materi = m.id
		WHERE sdd.id = $1`
	var soal entity.SoalDragDrop
	var materi entity.Materi
	err := r.db.QueryRow(query, id).Scan(
		&soal.ID, &soal.Pertanyaan, &soal.IDMateri, &soal.IsActive, &soal.CreatedAt, &soal.UpdatedAt,
		&materi.ID, &materi.Nama, &materi.IDMataPelajaran, &materi.IDTingkat,
	)
	if err != nil {
		return nil, err
	}
	soal.Materi = materi
	return &soal, nil
}
