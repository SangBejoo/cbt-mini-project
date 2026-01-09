package test_session

import (
	"cbt-test-mini-project/internal/entity"
	"errors"
	"math/rand"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	err := r.db.Preload("MataPelajaran").Preload("Tingkat").Preload("User").Where("session_token = ?", token).First(&session).Error
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

// UpdateSessionStatus updates only the status of a session
func (r *testSessionRepositoryImpl) UpdateSessionStatus(token string, status entity.TestStatus) error {
	return r.db.Model(&entity.TestSession{}).Where("session_token = ?", token).Update("status", status).Error
}

// List sessions with filters
func (r *testSessionRepositoryImpl) List(tingkatan, idMataPelajaran *int, status *entity.TestStatus, limit, offset int) ([]entity.TestSession, int, error) {
	var sessions []entity.TestSession
	var total int64

	// Build count query without preloads
	countQuery := r.db.Model(&entity.TestSession{})

	if tingkatan != nil {
		countQuery = countQuery.Where("id_tingkat = ?", *tingkatan)
	}
	if idMataPelajaran != nil {
		countQuery = countQuery.Where("id_mata_pelajaran = ?", *idMataPelajaran)
	}
	if status != nil {
		countQuery = countQuery.Where("status = ?", *status)
	}

	// Count total
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Build data query with preloads
	query := r.db.Model(&entity.TestSession{}).Preload("MataPelajaran").Preload("Tingkat").Preload("User")

	if tingkatan != nil {
		query = query.Where("id_tingkat = ?", *tingkatan)
	}
	if idMataPelajaran != nil {
		query = query.Where("id_mata_pelajaran = ?", *idMataPelajaran)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
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
		Preload("SoalDragDrop", func(db *gorm.DB) *gorm.DB { return db.Preload("Materi").Preload("Items", func(db2 *gorm.DB) *gorm.DB { return db2.Order("urutan ASC") }).Preload("Slots", func(db2 *gorm.DB) *gorm.DB { return db2.Order("urutan ASC") }).Preload("Gambar", func(db2 *gorm.DB) *gorm.DB { return db2.Order("urutan ASC") }) }).
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

	// Use GORM Clauses for Upsert (On Conflict)
	// Prepare the answer object
	isCorrect := (jawaban == tss.Soal.JawabanBenar)
	newAnswer := entity.JawabanSiswa{
		IDTestSessionSoal: tss.ID,
		JawabanDipilih:    &jawaban,
		IsCorrect:         isCorrect,
		QuestionType:      entity.QuestionTypeMultipleChoice,
	}

	// Upsert: If exists, update answer and correctness; if not, create new.
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id_test_session_soal"}},
		DoUpdates: clause.AssignmentColumns([]string{"jawaban_dipilih", "is_correct", "dijawab_pada"}),
	}).Create(&newAnswer).Error
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

	// Get drag-drop question IDs for the same criteria
	var dragDropIDs []int
	err = r.db.Model(&entity.SoalDragDrop{}).
		Joins("JOIN materi ON soal_drag_drop.id_materi = materi.id").
		Where("materi.id_mata_pelajaran = ? AND materi.id_tingkat = ? AND soal_drag_drop.is_active = ?", idMataPelajaran, tingkatan, true).
		Pluck("soal_drag_drop.id", &dragDropIDs).Error
	if err != nil {
		return err
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
			tss := entity.TestSessionSoal{
				IDTestSession: sessionID,
				QuestionType:  entity.QuestionTypeMultipleChoice,
				IDSoal:        &soalIDPtr,
				NomorUrut:     i + 1,
			}
			if err := r.db.Create(&tss).Error; err != nil {
				return err
			}
		case entity.QuestionTypeDragDrop:
			soalDragDropIDPtr := question.ID // Create a copy for pointer
			tss := entity.TestSessionSoal{
				IDTestSession:   sessionID,
				QuestionType:    entity.QuestionTypeDragDrop,
				IDSoalDragDrop:  &soalDragDropIDPtr,
				NomorUrut:       i + 1,
			}
			if err := r.db.Create(&tss).Error; err != nil {
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
	return r.db.Create(&newAnswer).Error
}

// GetTestSessionSoalByOrder gets TestSessionSoal by token and nomor_urut
func (r *testSessionRepositoryImpl) GetTestSessionSoalByOrder(token string, nomorUrut int) (*entity.TestSessionSoal, error) {
	var tss entity.TestSessionSoal
	err := r.db.Joins("JOIN test_session ON test_session_soal.id_test_session = test_session.id").
		Preload("Soal").
		Preload("SoalDragDrop").
		Preload("SoalDragDrop.Items").
		Preload("SoalDragDrop.Slots").
		Preload("SoalDragDrop.Gambar", func(db *gorm.DB) *gorm.DB { return db.Order("urutan ASC") }).
		Where("test_session.session_token = ? AND test_session_soal.nomor_urut = ?", token, nomorUrut).
		First(&tss).Error
	if err != nil {
		return nil, err
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

	// Use GORM Clauses for Upsert (On Conflict)
	// Assuming unique constraint on (id_test_session_soal)
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id_test_session_soal"}},
		DoUpdates: clause.AssignmentColumns([]string{"jawaban_drag_drop", "question_type", "is_correct", "dijawab_pada"}),
	}).Create(&newAnswer).Error
}

// GetDragDropCorrectAnswers gets correct answers for a drag-drop question
func (r *testSessionRepositoryImpl) GetDragDropCorrectAnswers(soalDragDropID int) ([]entity.DragCorrectAnswer, error) {
	var correctAnswers []entity.DragCorrectAnswer
	err := r.db.
		Joins("JOIN drag_item ON drag_correct_answer.id_drag_item = drag_item.id").
		Where("drag_item.id_soal_drag_drop = ?", soalDragDropID).
		Find(&correctAnswers).Error
	return correctAnswers, err
}

// GetSoalDragDropByID gets a drag-drop question by ID
func (r *testSessionRepositoryImpl) GetSoalDragDropByID(id int) (*entity.SoalDragDrop, error) {
	var soal entity.SoalDragDrop
	err := r.db.
		Preload("Materi").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("urutan ASC")
		}).
		Preload("Slots", func(db *gorm.DB) *gorm.DB {
			return db.Order("urutan ASC")
		}).
		Preload("Gambar", func(db *gorm.DB) *gorm.DB {
			return db.Order("urutan ASC")
		}).
		First(&soal, id).Error
	if err != nil {
		return nil, err
	}
	return &soal, nil
}