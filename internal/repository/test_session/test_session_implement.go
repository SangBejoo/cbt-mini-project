package test_session

import (
	"cbt-test-mini-project/internal/entity"
	"errors"
	"math/rand"
	"time"

	"gorm.io/gorm"
)

// testSessionRepositoryImpl implements TestSessionRepository
type testSessionRepositoryImpl struct {
	db *gorm.DB
}

// NewTestSessionRepository creates a new TestSessionRepository instance
func NewTestSessionRepository(db *gorm.DB) TestSessionRepository {
	return &testSessionRepositoryImpl{db: db}
}

// Create a new test session
func (r *testSessionRepositoryImpl) Create(session *entity.TestSession) error {
	return r.db.Create(session).Error
}

// Get session by token
func (r *testSessionRepositoryImpl) GetByToken(token string) (*entity.TestSession, error) {
	var session entity.TestSession
	err := r.db.Preload("MataPelajaran").Preload("Tingkat").Where("session_token = ?", token).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Update existing session
func (r *testSessionRepositoryImpl) Update(session *entity.TestSession) error {
	return r.db.Save(session).Error
}

// Delete session by ID
func (r *testSessionRepositoryImpl) Delete(id int) error {
	return r.db.Delete(&entity.TestSession{}, id).Error
}

// Complete session
func (r *testSessionRepositoryImpl) CompleteSession(token string, waktuSelesai time.Time, nilaiAkhir *float64, jumlahBenar, totalSoal *int) error {
	return r.db.Model(&entity.TestSession{}).Where("session_token = ?", token).Updates(map[string]interface{}{
		"waktu_selesai": waktuSelesai,
		"nilai_akhir":   nilaiAkhir,
		"jumlah_benar":  jumlahBenar,
		"total_soal":    totalSoal,
		"status":        entity.TestStatusCompleted,
	}).Error
}

// List sessions with filters
func (r *testSessionRepositoryImpl) List(tingkatan, idMataPelajaran *int, status *entity.TestStatus, limit, offset int) ([]entity.TestSession, int, error) {
	var sessions []entity.TestSession
	var total int64

	query := r.db.Model(&entity.TestSession{}).Preload("MataPelajaran").Preload("Tingkat")

	if tingkatan != nil {
		query = query.Where("id_tingkat = ?", *tingkatan)
	}
	if idMataPelajaran != nil {
		query = query.Where("id_mata_pelajaran = ?", *idMataPelajaran)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	if err := query.Limit(limit).Offset(offset).Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, int(total), nil
}

// Get questions for session
func (r *testSessionRepositoryImpl) GetSessionQuestions(token string) ([]entity.TestSessionSoal, error) {
	var sessionSoals []entity.TestSessionSoal
	err := r.db.Preload("Soal").Preload("Soal.Materi").Joins("JOIN test_session ON test_session_soal.id_test_session = test_session.id").
		Where("test_session.session_token = ?", token).Order("nomor_urut").Find(&sessionSoals).Error
	return sessionSoals, err
}

// Get all questions for session with soal data
func (r *testSessionRepositoryImpl) GetAllQuestionsForSession(token string) ([]entity.TestSessionSoal, error) {
	var sessionSoals []entity.TestSessionSoal
	err := r.db.Preload("Soal", func(db *gorm.DB) *gorm.DB { return db.Preload("Gambar", func(db2 *gorm.DB) *gorm.DB { return db2.Order("urutan ASC") }).Preload("Materi").Preload("Materi.MataPelajaran").Preload("Materi.Tingkat") }).
		Joins("JOIN test_session ON test_session_soal.id_test_session = test_session.id").
		Where("test_session.session_token = ?", token).Order("nomor_urut").Find(&sessionSoals).Error
	return sessionSoals, err
}

// Get single question by order
func (r *testSessionRepositoryImpl) GetQuestionByOrder(token string, nomorUrut int) (*entity.Soal, error) {
	var soal entity.Soal
	err := r.db.Joins("JOIN test_session_soal ON test_session_soal.id_soal = soal.id").
		Joins("JOIN test_session ON test_session_soal.id_test_session = test_session.id").
		Where("test_session.session_token = ? AND test_session_soal.nomor_urut = ?", token, nomorUrut).
		Preload("Materi").Preload("Materi.MataPelajaran").Preload("Materi.Tingkat").Preload("Gambar", func(db *gorm.DB) *gorm.DB { return db.Order("urutan ASC") }).First(&soal).Error
	if err != nil {
		return nil, err
	}
	return &soal, nil
}

// Submit answer
func (r *testSessionRepositoryImpl) SubmitAnswer(token string, nomorUrut int, jawaban entity.JawabanOption) error {
	// Find the TestSessionSoal
	var tss entity.TestSessionSoal
	err := r.db.Joins("JOIN test_session ON test_session_soal.id_test_session = test_session.id").
		Preload("Soal").
		Where("test_session.session_token = ? AND test_session_soal.nomor_urut = ?", token, nomorUrut).First(&tss).Error
	if err != nil {
		return err
	}

	// Check if already answered
	var existing entity.JawabanSiswa
	err = r.db.Where("id_test_session_soal = ?", tss.ID).First(&existing).Error
	if err == nil {
		// Update existing
		existing.JawabanDipilih = &jawaban
		existing.IsCorrect = (*existing.JawabanDipilih == tss.Soal.JawabanBenar)
		return r.db.Save(&existing).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new
		isCorrect := (jawaban == tss.Soal.JawabanBenar)
		newAnswer := entity.JawabanSiswa{
			IDTestSessionSoal: tss.ID,
			JawabanDipilih:    &jawaban,
			IsCorrect:         isCorrect,
		}
		return r.db.Create(&newAnswer).Error
	}
	return err
}

// Clear answer
func (r *testSessionRepositoryImpl) ClearAnswer(token string, nomorUrut int) error {
	// Find the TestSessionSoal
	var tss entity.TestSessionSoal
	err := r.db.Joins("JOIN test_session ON test_session_soal.id_test_session = test_session.id").
		Where("test_session.session_token = ? AND test_session_soal.nomor_urut = ?", token, nomorUrut).First(&tss).Error
	if err != nil {
		return err
	}

	// Delete the answer if exists
	return r.db.Where("id_test_session_soal = ?", tss.ID).Delete(&entity.JawabanSiswa{}).Error
}

// Get answers for session
func (r *testSessionRepositoryImpl) GetSessionAnswers(token string) ([]entity.JawabanSiswa, error) {
	var answers []entity.JawabanSiswa
	err := r.db.Preload("TestSessionSoal").Preload("TestSessionSoal.Soal", func(db *gorm.DB) *gorm.DB { return db.Preload("Gambar", func(db2 *gorm.DB) *gorm.DB { return db2.Order("urutan ASC") }) }).
		Joins("JOIN test_session_soal ON jawaban_siswa.id_test_session_soal = test_session_soal.id").
		Joins("JOIN test_session ON test_session_soal.id_test_session = test_session.id").
		Where("test_session.session_token = ?", token).Find(&answers).Error
	return answers, err
}

// Assign random questions to session
func (r *testSessionRepositoryImpl) AssignRandomQuestions(sessionID, idMataPelajaran, tingkatan, jumlahSoal int) error {
	// Get random soal IDs for the criteria - get questions for the mata_pelajaran and tingkat
	var soalIDs []int
	err := r.db.Model(&entity.Soal{}).
		Joins("JOIN materi ON soal.id_materi = materi.id").
		Where("materi.id_mata_pelajaran = ? AND materi.id_tingkat = ?", idMataPelajaran, tingkatan). // ini disesuai dengan tingkatan kalau mau random tingkatan tinggal hapus saja kondisi ini
		Pluck("soal.id", &soalIDs).Error
	if err != nil {
		return err
	}

	if len(soalIDs) == 0 {
		return errors.New("tidak ada soal yang tersedia untuk mata pelajaran dan tingkatan ini")
	}

	// Ambil jumlah soal yang tersedia atau yang diminta (mana yang lebih kecil)
	actualJumlahSoal := jumlahSoal
	if len(soalIDs) < jumlahSoal {
		actualJumlahSoal = len(soalIDs)
	}

	// Shuffle and select
	rand.Shuffle(len(soalIDs), func(i, j int) { soalIDs[i], soalIDs[j] = soalIDs[j], soalIDs[i] })
	selectedIDs := soalIDs[:actualJumlahSoal]

	// Create TestSessionSoal entries
	for i, soalID := range selectedIDs {
		tss := entity.TestSessionSoal{
			IDTestSession: sessionID,
			IDSoal:        soalID,
			NomorUrut:     i + 1,
		}
		if err := r.db.Create(&tss).Error; err != nil {
			return err
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
	return r.db.Create(&newAnswer).Error
}