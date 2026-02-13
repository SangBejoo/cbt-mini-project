package entity

import "time"

// TestStatus defines valid test session statuses
type TestStatus string

const (
	TestStatusOngoing    TestStatus = "ongoing"
	TestStatusCompleted  TestStatus = "completed"
	TestStatusTimeout    TestStatus = "timeout"
	TestStatusScheduled  TestStatus = "scheduled"
)

// TestSession represents the test_session table
type TestSession struct {
	ID              int           `json:"id" gorm:"primaryKey;autoIncrement"`
	SessionToken    string        `json:"session_token" gorm:"unique;not null;size:64"`
	UserID          *int          `json:"user_id" gorm:"column:user_id"` // Nullable for backward compatibility
	User            *User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	NamaPeserta     string        `json:"nama_peserta" gorm:"not null;size:100"` // Keep for backward compatibility
	IDTingkat       int           `json:"id_tingkat" gorm:"not null"`
	Tingkat         Tingkat       `json:"tingkat" gorm:"foreignKey:IDTingkat"`
	IDMataPelajaran int           `json:"id_mata_pelajaran" gorm:"not null"`
	MataPelajaran   MataPelajaran `json:"mata_pelajaran" gorm:"foreignKey:IDMataPelajaran"`

	// Ganti default:CURRENT_TIMESTAMP dengan autoCreateTime biar dihandle GORM
	WaktuMulai   time.Time  `json:"waktu_mulai" gorm:"autoCreateTime"`
	WaktuSelesai *time.Time `json:"waktu_selesai"`

	DurasiMenit int        `json:"durasi_menit" gorm:"not null"`
	NilaiAkhir  *float64   `json:"nilai_akhir" gorm:"type:decimal(5,2)"`
	JumlahBenar *int       `json:"jumlah_benar"`
	TotalSoal   *int       `json:"total_soal"`
	Status      TestStatus `json:"status" gorm:"type:enum('ongoing','completed','timeout','scheduled');default:'ongoing'"`

	LMSAssignmentID *int64 `json:"lms_assignment_id" gorm:"column:lms_assignment_id"`
	LMSClassID      *int64 `json:"lms_class_id" gorm:"column:lms_class_id"`
}

func (TestSession) TableName() string { return "test_session" }

// BatasWaktu calculates deadline from WaktuMulai + DurasiMenit
func (ts TestSession) BatasWaktu() time.Time {
	return ts.WaktuMulai.Add(time.Duration(ts.DurasiMenit) * time.Minute)
}

// TestSessionSoal represents the test_session_soal table (supports both MC and Drag-Drop)
type TestSessionSoal struct {
	ID int `json:"id" gorm:"primaryKey;autoIncrement"`

	// Composite Unique Index: Satu sesi tidak boleh punya dua soal dengan nomor urut sama
	IDTestSession int         `json:"id_test_session" gorm:"not null;index:idx_session_urut,unique"`
	TestSession   TestSession `json:"-" gorm:"foreignKey:IDTestSession;constraint:OnDelete:CASCADE"`

	// Question type for routing
	QuestionType QuestionType `json:"question_type" gorm:"type:enum('multiple_choice','drag_drop');default:'multiple_choice'"`

	// Multiple-choice question FK (nullable when QuestionType is drag_drop)
	IDSoal *int `json:"id_soal" gorm:""`
	Soal   *Soal `json:"soal" gorm:"foreignKey:IDSoal"`

	// Drag-drop question FK (nullable when QuestionType is multiple_choice)
	IDSoalDragDrop *int          `json:"id_soal_drag_drop" gorm:""`
	SoalDragDrop   *SoalDragDrop `json:"soal_drag_drop,omitempty" gorm:"foreignKey:IDSoalDragDrop"`

	NomorUrut int `json:"nomor_urut" gorm:"not null;index:idx_session_urut,unique"`
}

func (TestSessionSoal) TableName() string { return "test_session_soal" }

// IsDragDrop returns true if this is a drag-drop question
func (tss TestSessionSoal) IsDragDrop() bool {
	return tss.QuestionType == QuestionTypeDragDrop
}

