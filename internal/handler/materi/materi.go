package materi

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/usecase/materi"
	"cbt-test-mini-project/internal/usecase/soal"
	"context"
)

// materiHandler implements base.MateriServiceServer
type materiHandler struct {
	base.UnimplementedMateriServiceServer
	usecase     materi.MateriUsecase
	soalUsecase soal.SoalUsecase
}

// NewMateriHandler creates a new MateriHandler
func NewMateriHandler(usecase materi.MateriUsecase, soalUsecase soal.SoalUsecase) base.MateriServiceServer {
	return &materiHandler{usecase: usecase, soalUsecase: soalUsecase}
}

// Helper function to convert entity.Materi to proto.Materi
func (h *materiHandler) convertToProtoMateri(m *entity.Materi, questionCount int) *base.Materi {
	return &base.Materi{
		Id:                    int32(m.ID),
		MataPelajaran:         &base.MataPelajaran{Id: int32(m.MataPelajaran.ID), Nama: m.MataPelajaran.Nama},
		Tingkat:               &base.Tingkat{Id: int32(m.Tingkat.ID), Nama: m.Tingkat.Nama},
		Nama:                  m.Nama,
		IsActive:              m.IsActive,
		DefaultDurasiMenit:    int32(m.DefaultDurasiMenit),
		DefaultJumlahSoal:     int32(m.DefaultJumlahSoal),
		JumlahSoalReal:        int32(questionCount),
	}
}

// CreateMateri creates a new materi
func (h *materiHandler) CreateMateri(ctx context.Context, req *base.CreateMateriRequest) (*base.MateriResponse, error) {
	m, err := h.usecase.CreateMateri(int(req.IdMataPelajaran), req.Nama, int(req.IdTingkat), req.IsActive, int(req.DefaultDurasiMenit), int(req.DefaultJumlahSoal))
	if err != nil {
		return nil, err
	}

	return &base.MateriResponse{
		Materi: h.convertToProtoMateri(m, 0),
	}, nil
}

// GetMateri gets materi by ID
func (h *materiHandler) GetMateri(ctx context.Context, req *base.GetMateriRequest) (*base.MateriResponse, error) {
	m, err := h.usecase.GetMateri(int(req.Id))
	if err != nil {
		return nil, err
	}

	// Get question count for this materi
	counts, _ := h.soalUsecase.GetQuestionCountsByTopic()
	questionCount := 0
	if counts != nil {
		questionCount = counts[m.ID]
	}

	return &base.MateriResponse{
		Materi: h.convertToProtoMateri(m, questionCount),
	}, nil
}

// UpdateMateri updates materi
func (h *materiHandler) UpdateMateri(ctx context.Context, req *base.UpdateMateriRequest) (*base.MateriResponse, error) {
	m, err := h.usecase.UpdateMateri(int(req.Id), int(req.IdMataPelajaran), req.Nama, int(req.IdTingkat), req.IsActive, int(req.DefaultDurasiMenit), int(req.DefaultJumlahSoal))
	if err != nil {
		return nil, err
	}

	// Get question count for this materi
	counts, _ := h.soalUsecase.GetQuestionCountsByTopic()
	questionCount := 0
	if counts != nil {
		questionCount = counts[m.ID]
	}

	return &base.MateriResponse{
		Materi: h.convertToProtoMateri(m, questionCount),
	}, nil
}

// DeleteMateri deletes materi
func (h *materiHandler) DeleteMateri(ctx context.Context, req *base.DeleteMateriRequest) (*base.MessageStatusResponse, error) {
	err := h.usecase.DeleteMateri(int(req.Id))
	if err != nil {
		return &base.MessageStatusResponse{
			Message: "Failed to delete materi",
			Status:  "error",
		}, err
	}

	return &base.MessageStatusResponse{
		Message: "Materi deleted successfully",
		Status:  "success",
	}, nil
}

// ListMateri lists materi
func (h *materiHandler) ListMateri(ctx context.Context, req *base.ListMateriRequest) (*base.ListMateriResponse, error) {
	page := 1
	pageSize := 100 // Changed from 10 to 100 for better default
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = int(req.Pagination.Page)
		}
		if req.Pagination.PageSize > 0 {
			pageSize = int(req.Pagination.PageSize)
		}
	}
	// Cap pageSize to prevent abuse
	if pageSize > 1000 {
		pageSize = 1000
	}
	materis, pagination, err := h.usecase.ListMateri(int(req.IdMataPelajaran), int(req.IdTingkat), page, pageSize)
	if err != nil {
		return nil, err
	}

	// Get all question counts in one query
	counts, _ := h.soalUsecase.GetQuestionCountsByTopic()
	if counts == nil {
		counts = make(map[int]int)
	}

	var materiList []*base.Materi
	for _, m := range materis {
		questionCount := counts[m.ID]
		materiList = append(materiList, h.convertToProtoMateri(&m, questionCount))
	}

	return &base.ListMateriResponse{
		Materi:     materiList,
		Pagination: &base.PaginationResponse{
			TotalCount:  int32(pagination.TotalCount),
			TotalPages:  int32(pagination.TotalPages),
			CurrentPage: int32(pagination.CurrentPage),
			PageSize:    int32(pagination.PageSize),
		},
	}, nil
}