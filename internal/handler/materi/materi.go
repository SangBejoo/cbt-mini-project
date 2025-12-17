package materi

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/usecase/materi"
	"context"
)

// materiHandler implements base.MateriServiceServer
type materiHandler struct {
	base.UnimplementedMateriServiceServer
	usecase materi.MateriUsecase
}

// NewMateriHandler creates a new MateriHandler
func NewMateriHandler(usecase materi.MateriUsecase) base.MateriServiceServer {
	return &materiHandler{usecase: usecase}
}

// CreateMateri creates a new materi
func (h *materiHandler) CreateMateri(ctx context.Context, req *base.CreateMateriRequest) (*base.MateriResponse, error) {
	m, err := h.usecase.CreateMateri(int(req.IdMataPelajaran), req.Nama, int(req.Tingkatan))
	if err != nil {
		return nil, err
	}

	return &base.MateriResponse{
		Materi: &base.Materi{
			Id:            int32(m.ID),
			MataPelajaran: &base.MataPelajaran{Id: int32(m.MataPelajaran.ID), Nama: m.MataPelajaran.Nama},
			Nama:          m.Nama,
			Tingkatan:     int32(m.Tingkatan),
		},
	}, nil
}

// GetMateri gets materi by ID
func (h *materiHandler) GetMateri(ctx context.Context, req *base.GetMateriRequest) (*base.MateriResponse, error) {
	m, err := h.usecase.GetMateri(int(req.Id))
	if err != nil {
		return nil, err
	}

	return &base.MateriResponse{
		Materi: &base.Materi{
			Id:            int32(m.ID),
			MataPelajaran: &base.MataPelajaran{Id: int32(m.MataPelajaran.ID), Nama: m.MataPelajaran.Nama},
			Nama:          m.Nama,
			Tingkatan:     int32(m.Tingkatan),
		},
	}, nil
}

// UpdateMateri updates materi
func (h *materiHandler) UpdateMateri(ctx context.Context, req *base.UpdateMateriRequest) (*base.MateriResponse, error) {
	m, err := h.usecase.UpdateMateri(int(req.Id), int(req.IdMataPelajaran), req.Nama, int(req.Tingkatan))
	if err != nil {
		return nil, err
	}

	return &base.MateriResponse{
		Materi: &base.Materi{
			Id:            int32(m.ID),
			MataPelajaran: &base.MataPelajaran{Id: int32(m.MataPelajaran.ID), Nama: m.MataPelajaran.Nama},
			Nama:          m.Nama,
			Tingkatan:     int32(m.Tingkatan),
		},
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
	pageSize := 10
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = int(req.Pagination.Page)
		}
		if req.Pagination.PageSize > 0 {
			pageSize = int(req.Pagination.PageSize)
		}
	}
	materis, pagination, err := h.usecase.ListMateri(int(req.IdMataPelajaran), int(req.Tingkatan), page, pageSize)
	if err != nil {
		return nil, err
	}

	var materiList []*base.Materi
	for _, m := range materis {
		materiList = append(materiList, &base.Materi{
			Id:            int32(m.ID),
			MataPelajaran: &base.MataPelajaran{Id: int32(m.MataPelajaran.ID), Nama: m.MataPelajaran.Nama},
			Nama:          m.Nama,
			Tingkatan:     int32(m.Tingkatan),
		})
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