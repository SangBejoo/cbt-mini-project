package test_session

import (
	"cbt-test-mini-project/internal/entity"
	"database/sql"
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
		INSERT INTO test_session (session_token, user_id, nama_peserta, id_tingkat, id_mata_pelajaran, waktu_mulai, waktu_selesai, durasi_menit, nilai_akhir, jumlah_benar, total_soal, status, lms_assignment_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`
	return r.db.QueryRow(query, session.SessionToken, session.UserID, session.NamaPeserta, session.IDTingkat, session.IDMataPelajaran, session.WaktuMulai, session.WaktuSelesai, session.DurasiMenit, session.NilaiAkhir, session.JumlahBenar, session.TotalSoal, string(session.Status), session.LMSAssignmentID).Scan(&session.ID)
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
		       u.id, u.email, u.nama, u.role, u.is_active, u.created_at, u.updated_at, u.lms_user_id
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
		SET session_token = $1, user_id = $2, nama_peserta = $3, id_tingkat = $4, id_mata_pelajaran = $5, waktu_mulai = $6, waktu_selesai = $7, durasi_menit = $8, nilai_akhir = $9, jumlah_benar = $10, total_soal = $11, status = $12, lms_assignment_id = $13
		WHERE id = $14`
	_, err := r.db.Exec(query, session.SessionToken, session.UserID, session.NamaPeserta, session.IDTingkat, session.IDMataPelajaran, session.WaktuMulai, session.WaktuSelesai, session.DurasiMenit, session.NilaiAkhir, session.JumlahBenar, session.TotalSoal, string(session.Status), session.LMSAssignmentID, session.ID)
	return err
}

// Delete session by ID
func (r *testSessionRepositoryImpl) Delete(id int) error {
	query := `DELETE FROM test_session WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// Complete session
func (r *testSessionRepositoryImpl) CompleteSession(token string, waktuSelesai time.Time, nilaiAkhir *float64, jumlahBenar, totalSoal *int) error {
	query := `
		UPDATE test_session
		SET waktu_selesai = $1, nilai_akhir = $2, jumlah_benar = $3, total_soal = $4, status = $5, updated_at = $6
		WHERE session_token = $7`
	_, err := r.db.Exec(query, waktuSelesai, nilaiAkhir, jumlahBenar, totalSoal, string(entity.TestStatusCompleted), time.Now(), token)
	return err
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
		       u.id, u.email, u.nama, u.role, u.is_active, u.created_at, u.updated_at, u.lms_user_id
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

// Get questions for session
func (r *testSessionRepositoryImpl) GetSessionQuestions(token string) ([]entity.TestSessionSoal, error) {
	query := `
		SELECT tss.id, tss.id_test_session, tss.question_type, tss.id_soal, tss.id_soal_drag_drop, tss.nomor_urut,
		       s.id, s.pertanyaan, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, s.jawaban_benar, s.id_materi,
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
		var soalPertanyaan, soalOpsiA, soalOpsiB, soalOpsiC, soalOpsiD, soalJawabanBenar sql.NullString
		var materiID, materiIDMataPelajaran, materiIDTingkat sql.NullInt64
		var materiNama sql.NullString
		
		err := rows.Scan(
			&tss.ID, &tss.IDTestSession, &tss.QuestionType, &tss.IDSoal, &tss.IDSoalDragDrop, &tss.NomorUrut,
			&soalID, &soalPertanyaan, &soalOpsiA, &soalOpsiB, &soalOpsiC, &soalOpsiD, &soalJawabanBenar, &soalIDMateri,
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
			soal.JawabanBenar = entity.JawabanOption(soalJawabanBenar.String)
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
		       s.id, s.pertanyaan, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, s.jawaban_benar, s.id_materi,
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
		var soalPertanyaan, soalOpsiA, soalOpsiB, soalOpsiC, soalOpsiD, soalJawabanBenar sql.NullString
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
			&soalID, &soalPertanyaan, &soalOpsiA, &soalOpsiB, &soalOpsiC, &soalOpsiD, &soalJawabanBenar, &soalIDMateri,
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
			soal.JawabanBenar = entity.JawabanOption(soalJawabanBenar.String)
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
		SELECT s.id, s.pertanyaan, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, s.jawaban_benar, s.id_materi,
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
	err := r.db.QueryRow(query, token, nomorUrut).Scan(
		&soal.ID, &soal.Pertanyaan, &soal.OpsiA, &soal.OpsiB, &soal.OpsiC, &soal.OpsiD, &soal.JawabanBenar, &soal.IDMateri,
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
		SELECT js.id, js.id_test_session_soal, js.jawaban_dipilih, js.is_correct, js.question_type, js.dijawab_pada, js.jawaban_drag_drop,
		       tss.id, tss.id_test_session, tss.question_type, tss.id_soal, tss.id_soal_drag_drop, tss.nomor_urut,
		       s.id, s.pertanyaan, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, s.jawaban_benar, s.id_materi
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
		var soalPertanyaan, soalOpsiA, soalOpsiB, soalOpsiC, soalOpsiD, soalJawabanBenar sql.NullString
		
		err := rows.Scan(
			&js.ID, &js.IDTestSessionSoal, &js.JawabanDipilih, &js.IsCorrect, &js.QuestionType, &js.DijawabPada, &js.JawabanDragDrop,
			&tss.ID, &tss.IDTestSession, &tss.QuestionType, &tss.IDSoal, &tss.IDSoalDragDrop, &tss.NomorUrut,
			&soalID, &soalPertanyaan, &soalOpsiA, &soalOpsiB, &soalOpsiC, &soalOpsiD, &soalJawabanBenar, &soalIDMateri,
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
			soal.JawabanBenar = entity.JawabanOption(soalJawabanBenar.String)
			if soalIDMateri.Valid {
				soal.IDMateri = int(soalIDMateri.Int64)
			}
			tss.Soal = &soal
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
		SELECT s.id
		FROM soal s
		JOIN materi m ON s.id_materi = m.id
		WHERE m.id_mata_pelajaran = $1 AND m.id_tingkat = $2`
	soalRows, err := r.db.Query(soalQuery, idMataPelajaran, tingkatan)
	if err != nil {
		return err
	}
	defer soalRows.Close()

	var soalIDs []int
	for soalRows.Next() {
		var id int
		soalRows.Scan(&id)
		soalIDs = append(soalIDs, id)
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

	// Combine all question IDs
	var allQuestionIDs []struct {
		ID           int
		QuestionType entity.QuestionType
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
		case entity.QuestionTypeMultipleChoice:
			soalIDPtr := question.ID // Create a copy for pointer
			insertQuery := `
				INSERT INTO test_session_soal (id_test_session, question_type, id_soal, nomor_urut)
				VALUES ($1, $2, $3, $4)`
			_, err := r.db.Exec(insertQuery, sessionID, string(entity.QuestionTypeMultipleChoice), soalIDPtr, i+1)
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
	newAnswer := entity.JawabanSiswa{
		IDTestSessionSoal: sessionSoalID,
		JawabanDipilih:    nil, // NULL - no answer provided
		IsCorrect:         false,
	}
	insertQuery := `
		INSERT INTO jawaban_siswa (id_test_session_soal, jawaban_dipilih, is_correct)
		VALUES ($1, $2, $3)`
	_, err := r.db.Exec(insertQuery, newAnswer.IDTestSessionSoal, newAnswer.JawabanDipilih, newAnswer.IsCorrect)
	return err
}

// GetTestSessionSoalByOrder gets TestSessionSoal by token and nomor_urut
func (r *testSessionRepositoryImpl) GetTestSessionSoalByOrder(token string, nomorUrut int) (*entity.TestSessionSoal, error) {
	query := `
		SELECT tss.id, tss.id_test_session, tss.question_type, tss.id_soal, tss.id_soal_drag_drop, tss.nomor_urut,
		       s.id, s.pertanyaan, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, s.jawaban_benar, s.id_materi,
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
	var soalPertanyaan, soalOpsiA, soalOpsiB, soalOpsiC, soalOpsiD, soalJawabanBenar sql.NullString
	var sddID, sddIDMateri sql.NullInt64
	var sddPertanyaan sql.NullString
	
	err := r.db.QueryRow(query, token, nomorUrut).Scan(
		&tss.ID, &tss.IDTestSession, &tss.QuestionType, &tss.IDSoal, &tss.IDSoalDragDrop, &tss.NomorUrut,
		&soalID, &soalPertanyaan, &soalOpsiA, &soalOpsiB, &soalOpsiC, &soalOpsiD, &soalJawabanBenar, &soalIDMateri,
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
		soal.JawabanBenar = entity.JawabanOption(soalJawabanBenar.String)
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