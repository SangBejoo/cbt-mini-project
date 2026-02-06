package history

import (
	"cbt-test-mini-project/internal/entity"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// historyRepositoryImpl implements HistoryRepository
type historyRepositoryImpl struct {
	db *sql.DB
}

// NewHistoryRepository creates a new HistoryRepository instance
func NewHistoryRepository(db *sql.DB) HistoryRepository {
	return &historyRepositoryImpl{db: db}
}

// Get student history
func (r *historyRepositoryImpl) GetStudentHistory(userID int, tingkatan, idMataPelajaran *int, limit, offset int) ([]entity.HistorySummary, int, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "ts.status = $"+fmt.Sprintf("%d", len(args)+1))
	args = append(args, entity.TestStatusCompleted)

	if userID > 0 {
		conditions = append(conditions, "ts.user_id = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, userID)
	}

	if tingkatan != nil {
		conditions = append(conditions, "ts.id_tingkat = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *tingkatan)
	}
	if idMataPelajaran != nil {
		conditions = append(conditions, "ts.id_mata_pelajaran = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *idMataPelajaran)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM test_session ts WHERE %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	selectQuery := fmt.Sprintf(`
		SELECT ts.id, ts.session_token, ts.nama_peserta, ts.waktu_mulai, ts.waktu_selesai, ts.nilai_akhir, ts.jumlah_benar, ts.total_soal, ts.status,
		       mp.nama as mata_pelajaran_nama, mp.is_active as mata_pelajaran_is_active,
		       t.nama as tingkat_nama, t.is_active as tingkat_is_active
		FROM test_session ts
		JOIN mata_pelajaran mp ON ts.id_mata_pelajaran = mp.id
		JOIN tingkat t ON ts.id_tingkat = t.id
		WHERE %s
		LIMIT $%d OFFSET $%d`, whereClause, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var histories []entity.HistorySummary
	for rows.Next() {
		var h entity.HistorySummary
		var waktuMulai time.Time
		var waktuSelesai sql.NullTime
		var nilaiAkhir sql.NullFloat64
		var jumlahBenar sql.NullInt32
		var totalSoal sql.NullInt32
		var mpNama string
		var mpIsActive bool
		var tNama string
		var tIsActive bool

		err := rows.Scan(&h.ID, &h.SessionToken, &h.NamaPeserta, &waktuMulai, &waktuSelesai, &nilaiAkhir, &jumlahBenar, &totalSoal, &h.Status,
			&mpNama, &mpIsActive, &tNama, &tIsActive)
		if err != nil {
			return nil, 0, err
		}

		h.WaktuMulai = &waktuMulai
		if waktuSelesai.Valid {
			h.WaktuSelesai = &waktuSelesai.Time
		}
		if nilaiAkhir.Valid {
			h.NilaiAkhir = nilaiAkhir.Float64
		}
		if jumlahBenar.Valid {
			h.JumlahBenar = int(jumlahBenar.Int32)
		}
		if totalSoal.Valid {
			h.TotalSoal = int(totalSoal.Int32)
		}

		h.MataPelajaran = entity.MataPelajaran{Nama: mpNama, IsActive: mpIsActive}
		h.Tingkat = entity.Tingkat{Nama: tNama, IsActive: tIsActive}

		durasi := 0
		if h.WaktuSelesai != nil {
			durasi = int(h.WaktuSelesai.Sub(*h.WaktuMulai).Seconds())
		}
		h.DurasiPengerjaanDetik = durasi

		histories = append(histories, h)
	}

	return histories, total, rows.Err()
}

// Get history detail by session token
func (r *historyRepositoryImpl) GetHistoryDetail(sessionToken string) (*entity.TestSession, []entity.JawabanDetail, []entity.MateriBreakdown, error) {
	// Get session
	session, err := r.getSessionByToken(sessionToken)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get answers
	answers, err := r.getAnswersBySessionToken(sessionToken)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get materi breakdown
	breakdowns, err := r.getMateriBreakdown(sessionToken)
	if err != nil {
		return nil, nil, nil, err
	}

	return session, answers, breakdowns, nil
}

func (r *historyRepositoryImpl) getSessionByToken(token string) (*entity.TestSession, error) {
	query := `
		SELECT ts.id, ts.session_token, ts.user_id, ts.nama_peserta, ts.waktu_mulai, ts.waktu_selesai, ts.nilai_akhir, ts.jumlah_benar, ts.total_soal, ts.status, ts.id_mata_pelajaran, ts.id_tingkat,
		       mp.nama as mata_pelajaran_nama, mp.is_active as mata_pelajaran_is_active,
		       t.nama as tingkat_nama, t.is_active as tingkat_is_active
		FROM test_session ts
		JOIN mata_pelajaran mp ON ts.id_mata_pelajaran = mp.id
		JOIN tingkat t ON ts.id_tingkat = t.id
		WHERE ts.session_token = $1`
	var session entity.TestSession
	var userID sql.NullInt32
	var waktuSelesai sql.NullTime
	var nilaiAkhir sql.NullFloat64
	var jumlahBenar sql.NullInt32
	var totalSoal sql.NullInt32
	var mpNama string
	var mpIsActive bool
	var tNama string
	var tIsActive bool

	err := r.db.QueryRow(query, token).Scan(&session.ID, &session.SessionToken, &userID, &session.NamaPeserta, &session.WaktuMulai, &waktuSelesai, &nilaiAkhir, &jumlahBenar, &totalSoal, &session.Status, &session.IDMataPelajaran, &session.IDTingkat,
		&mpNama, &mpIsActive, &tNama, &tIsActive)
	if err != nil {
		return nil, err
	}

	if userID.Valid {
		uid := int(userID.Int32)
		session.UserID = &uid
	}
	if waktuSelesai.Valid {
		session.WaktuSelesai = &waktuSelesai.Time
	}
	if nilaiAkhir.Valid {
		session.NilaiAkhir = &nilaiAkhir.Float64
	}
	if jumlahBenar.Valid {
		jb := int(jumlahBenar.Int32)
		session.JumlahBenar = &jb
	}
	if totalSoal.Valid {
		ts := int(totalSoal.Int32)
		session.TotalSoal = &ts
	}

	session.MataPelajaran = entity.MataPelajaran{Nama: mpNama, IsActive: mpIsActive}
	session.Tingkat = entity.Tingkat{Nama: tNama, IsActive: tIsActive}

	return &session, nil
}

func (r *historyRepositoryImpl) getAnswersBySessionToken(token string) ([]entity.JawabanDetail, error) {
	query := `
		SELECT tss.nomor_urut, s.pertanyaan, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, js.jawaban_dipilih, s.jawaban_benar, js.is_correct, s.pembahasan,
		       CASE WHEN js.id IS NOT NULL THEN true ELSE false END as is_answered
		FROM test_session_soal tss
		JOIN test_session ts ON tss.id_test_session = ts.id
		JOIN soal s ON tss.id_soal = s.id
		LEFT JOIN jawaban_siswa js ON tss.id = js.id_test_session_soal
		WHERE ts.session_token = $1
		ORDER BY tss.nomor_urut`

	rows, err := r.db.Query(query, token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var details []entity.JawabanDetail
	for rows.Next() {
		var detail entity.JawabanDetail
		var jawabanDipilih sql.NullString
		var isCorrect sql.NullBool
		var pembahasan sql.NullString
		err := rows.Scan(&detail.NomorUrut, &detail.Pertanyaan, &detail.OpsiA, &detail.OpsiB, &detail.OpsiC, &detail.OpsiD, &jawabanDipilih, &detail.JawabanBenar, &isCorrect, &pembahasan, &detail.IsAnswered)
		if err != nil {
			return nil, err
		}
		if jawabanDipilih.Valid {
			s := entity.JawabanOption(jawabanDipilih.String)
			detail.JawabanDipilih = &s
		}
		if isCorrect.Valid {
			detail.IsCorrect = isCorrect.Bool
		}
		if pembahasan.Valid {
			detail.Pembahasan = &pembahasan.String
		}
		details = append(details, detail)
	}

	// Load gambar for each detail
	for i := range details {
		gambarQuery := `
			SELECT sg.nama_file, sg.file_path, sg.file_size, sg.mime_type, sg.urutan, sg.keterangan, sg.cloud_id, sg.public_id, sg.created_at
			FROM soal_gambar sg
			WHERE sg.id_soal = (
				SELECT s.id FROM soal s
				JOIN test_session_soal tss ON tss.id_soal = s.id
				JOIN test_session ts ON tss.id_test_session = ts.id
				WHERE ts.session_token = $1 AND tss.nomor_urut = $2
			)`
		gRows, err := r.db.Query(gambarQuery, token, details[i].NomorUrut)
		if err != nil {
			return nil, err
		}
		var gambar []entity.SoalGambar
		for gRows.Next() {
			var g entity.SoalGambar
			err := gRows.Scan(&g.NamaFile, &g.FilePath, &g.FileSize, &g.MimeType, &g.Urutan, &g.Keterangan, &g.CloudId, &g.PublicId, &g.CreatedAt)
			if err != nil {
				gRows.Close()
				return nil, err
			}
			gambar = append(gambar, g)
		}
		gRows.Close()
		details[i].Gambar = gambar
	}

	return details, rows.Err()
}

func (r *historyRepositoryImpl) getMateriBreakdown(token string) ([]entity.MateriBreakdown, error) {
	query := `
		SELECT
			m.nama as nama_materi,
			COUNT(s.id) as jumlah_soal,
			SUM(CASE WHEN js.is_correct THEN 1 ELSE 0 END) as jumlah_benar,
			ROUND((SUM(CASE WHEN js.is_correct THEN 1 ELSE 0 END) / COUNT(s.id)) * 100, 2) as persentase_benar
		FROM test_session_soal tss
		JOIN test_session ts ON tss.id_test_session = ts.id
		JOIN soal s ON tss.id_soal = s.id
		JOIN materi m ON s.id_materi = m.id
		LEFT JOIN jawaban_siswa js ON tss.id = js.id_test_session_soal
		WHERE ts.session_token = $1
		GROUP BY m.id, m.nama`

	rows, err := r.db.Query(query, token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var breakdowns []entity.MateriBreakdown
	for rows.Next() {
		var b entity.MateriBreakdown
		err := rows.Scan(&b.NamaMateri, &b.JumlahSoal, &b.JumlahBenar, &b.PersentaseBenar)
		if err != nil {
			return nil, err
		}
		breakdowns = append(breakdowns, b)
	}

	return breakdowns, rows.Err()
}

// Get user from session token
func (r *historyRepositoryImpl) GetUserFromSessionToken(sessionToken string) (*entity.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.nama, u.role, u.is_active, u.created_at, u.updated_at
		FROM users u
		JOIN test_session ts ON ts.user_id = u.id
		WHERE ts.session_token = $1`
	var user entity.User
	err := r.db.QueryRow(query, sessionToken).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Nama, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}

// ListStudentHistories lists all student histories with user info
func (r *historyRepositoryImpl) ListStudentHistories(userID, tingkatan, idMataPelajaran *int, limit, offset int) ([]entity.StudentHistoryWithUser, int, error) {
	// Build conditions for user_ids
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "status = $"+fmt.Sprintf("%d", len(args)+1))
	args = append(args, entity.TestStatusCompleted)

	if userID != nil && *userID > 0 {
		conditions = append(conditions, "user_id = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *userID)
	}
	if tingkatan != nil && *tingkatan > 0 {
		conditions = append(conditions, "id_tingkat = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *tingkatan)
	}
	if idMataPelajaran != nil && *idMataPelajaran > 0 {
		conditions = append(conditions, "id_mata_pelajaran = $"+fmt.Sprintf("%d", len(args)+1))
		args = append(args, *idMataPelajaran)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total users
	countQuery := fmt.Sprintf("SELECT COUNT(DISTINCT user_id) FROM test_session WHERE %s", whereClause)
	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get user IDs with pagination
	userIDQuery := fmt.Sprintf("SELECT DISTINCT user_id FROM test_session WHERE %s ORDER BY user_id LIMIT $%d OFFSET $%d", whereClause, len(args)+1, len(args)+2)
	args = append(args, limit, offset)
	userIDRows, err := r.db.Query(userIDQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer userIDRows.Close()

	var userIDs []int
	for userIDRows.Next() {
		var uid int
		err := userIDRows.Scan(&uid)
		if err != nil {
			return nil, 0, err
		}
		userIDs = append(userIDs, uid)
	}

	// For each user, get their histories
	var results []entity.StudentHistoryWithUser
	for _, uid := range userIDs {
		// Get user info
		userQuery := "SELECT id, email, password_hash, nama, role, is_active, created_at, updated_at FROM users WHERE id = $1"
		var user entity.User
		err := r.db.QueryRow(userQuery, uid).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Nama, &user.Role, &user.IsActive, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			continue // Skip if user not found
		}

		// Get user's completed test sessions
		sessionConditions := []string{"user_id = $1", "status = $2"}
		sessionArgs := []interface{}{uid, entity.TestStatusCompleted}

		if tingkatan != nil && *tingkatan > 0 {
			sessionConditions = append(sessionConditions, fmt.Sprintf("id_tingkat = $%d", len(sessionArgs)+1))
			sessionArgs = append(sessionArgs, *tingkatan)
		}
		if idMataPelajaran != nil && *idMataPelajaran > 0 {
			sessionConditions = append(sessionConditions, fmt.Sprintf("id_mata_pelajaran = $%d", len(sessionArgs)+1))
			sessionArgs = append(sessionArgs, *idMataPelajaran)
		}

		sessionWhere := strings.Join(sessionConditions, " AND ")
		sessionQuery := fmt.Sprintf(`
			SELECT ts.id, ts.session_token, ts.nama_peserta, ts.waktu_mulai, ts.waktu_selesai, ts.nilai_akhir, ts.jumlah_benar, ts.total_soal, ts.status,
			       mp.nama as mata_pelajaran_nama, mp.is_active as mata_pelajaran_is_active,
			       t.nama as tingkat_nama, t.is_active as tingkat_is_active
			FROM test_session ts
			JOIN mata_pelajaran mp ON ts.id_mata_pelajaran = mp.id
			JOIN tingkat t ON ts.id_tingkat = t.id
			WHERE %s`, sessionWhere)

		sessionRows, err := r.db.Query(sessionQuery, sessionArgs...)
		if err != nil {
			continue
		}

		var sessions []entity.TestSession
		for sessionRows.Next() {
			var s entity.TestSession
			var waktuMulai time.Time
			var waktuSelesai sql.NullTime
			var nilaiAkhir sql.NullFloat64
			var jumlahBenar sql.NullInt32
			var totalSoal sql.NullInt32
			var mpNama string
			var mpIsActive bool
			var tNama string
			var tIsActive bool

			err := sessionRows.Scan(&s.ID, &s.SessionToken, &s.NamaPeserta, &waktuMulai, &waktuSelesai, &nilaiAkhir, &jumlahBenar, &totalSoal, &s.Status,
				&mpNama, &mpIsActive, &tNama, &tIsActive)
			if err != nil {
				sessionRows.Close()
				continue
			}

			s.WaktuMulai = waktuMulai
			if waktuSelesai.Valid {
				s.WaktuSelesai = &waktuSelesai.Time
			}
			if nilaiAkhir.Valid {
				s.NilaiAkhir = &nilaiAkhir.Float64
			}
			if jumlahBenar.Valid {
				jb := int(jumlahBenar.Int32)
				s.JumlahBenar = &jb
			}
			if totalSoal.Valid {
				ts := int(totalSoal.Int32)
				s.TotalSoal = &ts
			}

			s.MataPelajaran = entity.MataPelajaran{Nama: mpNama, IsActive: mpIsActive}
			s.Tingkat = entity.Tingkat{Nama: tNama, IsActive: tIsActive}

			sessions = append(sessions, s)
		}
		sessionRows.Close()

		// Convert to HistorySummary
		histories := make([]entity.HistorySummary, len(sessions))
		totalNilai := 0.0
		for i, s := range sessions {
			durasi := 0
			if s.WaktuSelesai != nil && !s.WaktuMulai.IsZero() {
				durasi = int(s.WaktuSelesai.Sub(s.WaktuMulai).Seconds())
			}

			nilai := 0.0
			if s.NilaiAkhir != nil {
				nilai = *s.NilaiAkhir
				totalNilai += nilai
			}

			jumlahBenar := 0
			if s.JumlahBenar != nil {
				jumlahBenar = *s.JumlahBenar
			}

			totalSoal := 0
			if s.TotalSoal != nil {
				totalSoal = *s.TotalSoal
			}

			histories[i] = entity.HistorySummary{
				SessionToken:          s.SessionToken,
				NamaPeserta:           s.NamaPeserta,
				MataPelajaran:         s.MataPelajaran,
				Tingkat:               s.Tingkat,
				NilaiAkhir:            nilai,
				DurasiPengerjaanDetik: durasi,
				WaktuMulai:            &s.WaktuMulai,
				WaktuSelesai:          s.WaktuSelesai,
				JumlahBenar:           jumlahBenar,
				TotalSoal:             totalSoal,
				Status:                s.Status,
			}
		}

		// Calculate average
		rataRata := 0.0
		if len(histories) > 0 {
			rataRata = totalNilai / float64(len(histories))
		}

		results = append(results, entity.StudentHistoryWithUser{
			User:              user,
			History:           histories,
			RataRataNilai:     rataRata,
			TotalTestCompleted: len(histories),
		})
	}

	return results, total, nil
}