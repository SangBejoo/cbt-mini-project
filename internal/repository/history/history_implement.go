package history

import (
	"cbt-test-mini-project/internal/entity"

	"gorm.io/gorm"
)

// historyRepositoryImpl implements HistoryRepository
type historyRepositoryImpl struct {
	db *gorm.DB
}

// NewHistoryRepository creates a new HistoryRepository instance
func NewHistoryRepository(db *gorm.DB) HistoryRepository {
	return &historyRepositoryImpl{db: db}
}

// Get student history
func (r *historyRepositoryImpl) GetStudentHistory(namaPeserta string, tingkatan, idMataPelajaran *int, limit, offset int) ([]entity.HistorySummary, int, error) {
	var histories []entity.HistorySummary
	var total int64

	query := r.db.Model(&entity.TestSession{}).Preload("MataPelajaran").Where("nama_peserta = ? AND status = ?", namaPeserta, entity.TestStatusCompleted)

	if tingkatan != nil {
		query = query.Where("tingkatan = ?", *tingkatan)
	}
	if idMataPelajaran != nil {
		query = query.Where("id_mata_pelajaran = ?", *idMataPelajaran)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Limit(limit).Offset(offset).Find(&histories).Error; err != nil {
		return nil, 0, err
	}

	return histories, int(total), nil
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
	var session entity.TestSession
	err := r.db.Preload("MataPelajaran").Where("session_token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *historyRepositoryImpl) getAnswersBySessionToken(token string) ([]entity.JawabanDetail, error) {
	var details []entity.JawabanDetail

	err := r.db.Table("jawaban_siswa").
		Select("test_session_soal.nomor_urut, soal.pertanyaan, soal.opsi_a, soal.opsi_b, soal.opsi_c, soal.opsi_d, jawaban_siswa.jawaban_dipilih, soal.jawaban_benar, jawaban_siswa.is_correct").
		Joins("JOIN test_session_soal ON jawaban_siswa.id_test_session_soal = test_session_soal.id").
		Joins("JOIN test_session ON test_session_soal.id_test_session = test_session.id").
		Joins("JOIN soal ON test_session_soal.id_soal = soal.id").
		Where("test_session.session_token = ?", token).
		Order("test_session_soal.nomor_urut").
		Scan(&details).Error

	return details, err
}

func (r *historyRepositoryImpl) getMateriBreakdown(token string) ([]entity.MateriBreakdown, error) {
	var breakdowns []entity.MateriBreakdown

	// This is a complex aggregation; using raw SQL for efficiency
	err := r.db.Raw(`
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
		WHERE ts.session_token = ?
		GROUP BY m.id, m.nama
	`, token).Scan(&breakdowns).Error

	return breakdowns, err
}