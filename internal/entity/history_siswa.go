package entity

import "time"

// MateriBreakdown for history details
type MateriBreakdown struct {
	NamaMateri      string  `json:"nama_materi"`
	JumlahSoal      int     `json:"jumlah_soal"`
	JumlahBenar     int     `json:"jumlah_benar"`
	PersentaseBenar float64 `json:"persentase_benar"`
}

// HistorySummary for student history
type HistorySummary struct {
	ID                    int           `json:"id"`
	SessionToken          string        `json:"session_token"`
	MataPelajaran         MataPelajaran `json:"mata_pelajaran"`
	Tingkat               Tingkat       `json:"tingkat"`
	WaktuMulai            time.Time     `json:"waktu_mulai"`
	WaktuSelesai          *time.Time    `json:"waktu_selesai"`
	DurasiPengerjaanDetik int           `json:"durasi_pengerjaan_detik"`
	NilaiAkhir            float64       `json:"nilai_akhir"`
	JumlahBenar           int           `json:"jumlah_benar"`
	TotalSoal             int           `json:"total_soal"`
	Status                TestStatus    `json:"status"`
}

// PaginationRequest for requests
type PaginationRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// PaginationResponse for responses
type PaginationResponse struct {
	TotalCount  int `json:"total_count"`
	TotalPages  int `json:"total_pages"`
	CurrentPage int `json:"current_page"`
	PageSize    int `json:"page_size"`
}

// StudentHistoryResponse for student history
type StudentHistoryResponse struct {
	NamaPeserta       string            `json:"nama_peserta"`
	Tingkatan         *int              `json:"tingkatan"`
	History           []HistorySummary  `json:"history"`
	RataRataNilai     float64           `json:"rata_rata_nilai"`
	TotalTestCompleted int              `json:"total_test_completed"`
	Pagination        PaginationResponse `json:"pagination"`
}

// HistoryDetailResponse for detailed history
type HistoryDetailResponse struct {
	SessionInfo     *TestSession       `json:"session_info"`
	DetailJawaban   []JawabanDetail    `json:"detail_jawaban"`
	BreakdownMateri []MateriBreakdown  `json:"breakdown_materi"`
}
