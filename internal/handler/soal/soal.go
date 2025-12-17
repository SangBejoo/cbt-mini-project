package soal

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/usecase/soal"
	"context"
)

// soalHandler implements base.SoalServiceServer
type soalHandler struct {
	base.UnimplementedSoalServiceServer
	usecase soal.SoalUsecase
}

// NewSoalHandler creates a new SoalHandler
func NewSoalHandler(usecase soal.SoalUsecase) base.SoalServiceServer {
	return &soalHandler{usecase: usecase}
}

// CreateSoal creates a new soal
func (h *soalHandler) CreateSoal(ctx context.Context, req *base.CreateSoalRequest) (*base.SoalResponse, error) {
	jawabanBenar := entity.JawabanOption(req.JawabanBenar.String()[0])
	s, err := h.usecase.CreateSoal(int(req.IdMateri), req.Pertanyaan, req.OpsiA, req.OpsiB, req.OpsiC, req.OpsiD, jawabanBenar)
	if err != nil {
		return nil, err
	}

	return &base.SoalResponse{
		Soal: &base.SoalFull{
			Id:            int32(s.ID),
			Materi:        &base.Materi{Id: int32(s.Materi.ID), Nama: s.Materi.Nama},
			Pertanyaan:    s.Pertanyaan,
			OpsiA:         s.OpsiA,
			OpsiB:         s.OpsiB,
			OpsiC:         s.OpsiC,
			OpsiD:         s.OpsiD,
			JawabanBenar:  base.JawabanOption(base.JawabanOption_value[string(s.JawabanBenar)]),
		},
	}, nil
}

// GetSoal gets soal by ID
func (h *soalHandler) GetSoal(ctx context.Context, req *base.GetSoalRequest) (*base.SoalResponse, error) {
	s, err := h.usecase.GetSoal(int(req.Id))
	if err != nil {
		return nil, err
	}

	return &base.SoalResponse{
		Soal: &base.SoalFull{
			Id:            int32(s.ID),
			Materi:        &base.Materi{Id: int32(s.Materi.ID), Nama: s.Materi.Nama},
			Pertanyaan:    s.Pertanyaan,
			OpsiA:         s.OpsiA,
			OpsiB:         s.OpsiB,
			OpsiC:         s.OpsiC,
			OpsiD:         s.OpsiD,
			JawabanBenar:  base.JawabanOption(base.JawabanOption_value[string(s.JawabanBenar)]),
		},
	}, nil
}

// UpdateSoal updates soal
func (h *soalHandler) UpdateSoal(ctx context.Context, req *base.UpdateSoalRequest) (*base.SoalResponse, error) {
	jawabanBenar := entity.JawabanOption(req.JawabanBenar.String()[0])
	s, err := h.usecase.UpdateSoal(int(req.Id), int(req.IdMateri), req.Pertanyaan, req.OpsiA, req.OpsiB, req.OpsiC, req.OpsiD, jawabanBenar)
	if err != nil {
		return nil, err
	}

	return &base.SoalResponse{
		Soal: &base.SoalFull{
			Id:            int32(s.ID),
			Materi:        &base.Materi{Id: int32(s.Materi.ID), Nama: s.Materi.Nama},
			Pertanyaan:    s.Pertanyaan,
			OpsiA:         s.OpsiA,
			OpsiB:         s.OpsiB,
			OpsiC:         s.OpsiC,
			OpsiD:         s.OpsiD,
			JawabanBenar:  base.JawabanOption(base.JawabanOption_value[string(s.JawabanBenar)]),
		},
	}, nil
}

// DeleteSoal deletes soal
func (h *soalHandler) DeleteSoal(ctx context.Context, req *base.DeleteSoalRequest) (*base.MessageStatusResponse, error) {
	err := h.usecase.DeleteSoal(int(req.Id))
	if err != nil {
		return &base.MessageStatusResponse{
			Message: "Failed to delete soal",
			Status:  "error",
		}, err
	}

	return &base.MessageStatusResponse{
		Message: "Soal deleted successfully",
		Status:  "success",
	}, nil
}

// ListSoal lists soal
func (h *soalHandler) ListSoal(ctx context.Context, req *base.ListSoalRequest) (*base.ListSoalResponse, error) {
	soals, pagination, err := h.usecase.ListSoal(int(req.IdMateri), int(req.Tingkatan), int(req.IdMataPelajaran), int(req.Pagination.Page), int(req.Pagination.PageSize))
	if err != nil {
		return nil, err
	}

	var soalList []*base.SoalFull
	for _, s := range soals {
		soalList = append(soalList, &base.SoalFull{
			Id:            int32(s.ID),
			Materi:        &base.Materi{Id: int32(s.Materi.ID), Nama: s.Materi.Nama},
			Pertanyaan:    s.Pertanyaan,
			OpsiA:         s.OpsiA,
			OpsiB:         s.OpsiB,
			OpsiC:         s.OpsiC,
			OpsiD:         s.OpsiD,
			JawabanBenar:  base.JawabanOption(base.JawabanOption_value[string(s.JawabanBenar)]),
		})
	}

	return &base.ListSoalResponse{
		Soal: soalList,
		Pagination: &base.PaginationResponse{
			TotalCount:  int32(pagination.TotalCount),
			TotalPages:  int32(pagination.TotalPages),
			CurrentPage: int32(pagination.CurrentPage),
			PageSize:    int32(pagination.PageSize),
		},
	}, nil
}