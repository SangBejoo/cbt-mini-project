package history

import (
	"cbt-test-mini-project/internal/entity"
)

// HistoryUsecase defines the interface for History usecase operations
type HistoryUsecase interface {
	GetStudentHistory(namaPeserta string, tingkatan, idMataPelajaran *int, page, pageSize int) (*entity.StudentHistoryResponse, error)
	GetHistoryDetail(sessionToken string) (*entity.HistoryDetailResponse, error)
}