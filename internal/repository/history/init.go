package history

import (
	"cbt-test-mini-project/internal/entity"
)

// HistoryRepository defines the interface for History repository operations
type HistoryRepository interface {
	// Get student history
	GetStudentHistory(namaPeserta string, tingkatan, idMataPelajaran *int, limit, offset int) ([]entity.HistorySummary, int, error)

	// Get history detail by session token
	GetHistoryDetail(sessionToken string) (*entity.TestSession, []entity.JawabanDetail, []entity.MateriBreakdown, error)
}