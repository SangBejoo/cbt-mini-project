package history

import (
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/repository/history"
	"errors"
)

// historyUsecaseImpl implements HistoryUsecase
type historyUsecaseImpl struct {
	repo history.HistoryRepository
}

// NewHistoryUsecase creates a new HistoryUsecase instance
func NewHistoryUsecase(repo history.HistoryRepository) HistoryUsecase {
	return &historyUsecaseImpl{repo: repo}
}

// GetStudentHistory gets student history with aggregates
func (u *historyUsecaseImpl) GetStudentHistory(namaPeserta string, tingkatan, idMataPelajaran *int, page, pageSize int) (*entity.StudentHistoryResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	histories, total, err := u.repo.GetStudentHistory(namaPeserta, tingkatan, idMataPelajaran, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}

	// Get nama peserta from first completed test if not provided
	actualNamaPeserta := namaPeserta
	if actualNamaPeserta == "" && len(histories) > 0 {
		// Fetch the actual nama peserta from the database
		if name, err := u.repo.GetSessionNameByToken(histories[0].SessionToken); err == nil && name != "" {
			actualNamaPeserta = name
		}
	}

	// Calculate aggregates
	totalCompleted := 0
	totalNilai := 0.0
	for _, h := range histories {
		if h.Status == entity.TestStatusCompleted {
			totalCompleted++
			totalNilai += h.NilaiAkhir
		}
	}

	rataRataNilai := 0.0
	if totalCompleted > 0 {
		rataRataNilai = totalNilai / float64(totalCompleted)
	}

	pagination := &entity.PaginationResponse{
		TotalCount:  total,
		TotalPages:  (total + pageSize - 1) / pageSize,
		CurrentPage: page,
		PageSize:    pageSize,
	}

	response := &entity.StudentHistoryResponse{
		NamaPeserta:       actualNamaPeserta,
		Tingkatan:         tingkatan,
		History:           histories,
		RataRataNilai:     rataRataNilai,
		TotalTestCompleted: totalCompleted,
		Pagination:        *pagination,
	}

	return response, nil
}

// GetHistoryDetail gets detailed history for a session
func (u *historyUsecaseImpl) GetHistoryDetail(sessionToken string) (*entity.HistoryDetailResponse, error) {
	if sessionToken == "" {
		return nil, errors.New("session token cannot be empty")
	}

	session, answers, breakdowns, err := u.repo.GetHistoryDetail(sessionToken)
	if err != nil {
		return nil, err
	}

	response := &entity.HistoryDetailResponse{
		SessionInfo:     session,
		DetailJawaban:   answers,
		BreakdownMateri: breakdowns,
	}

	return response, nil
}