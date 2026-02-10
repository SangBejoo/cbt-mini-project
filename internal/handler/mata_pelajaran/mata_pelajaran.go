package mata_pelajaran

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/usecase/mata_pelajaran"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

// mataPelajaranHandler implements base.MataPelajaranServiceServer
type mataPelajaranHandler struct {
	base.UnimplementedMataPelajaranServiceServer
	usecase mata_pelajaran.MataPelajaranUsecase
}

// NewMataPelajaranHandler creates a new MataPelajaranHandler
func NewMataPelajaranHandler(usecase mata_pelajaran.MataPelajaranUsecase) base.MataPelajaranServiceServer {
	return &mataPelajaranHandler{usecase: usecase}
}

// CreateMataPelajaran creates a new mata pelajaran
func (h *mataPelajaranHandler) CreateMataPelajaran(ctx context.Context, req *base.CreateMataPelajaranRequest) (*base.MataPelajaranResponse, error) {
	mp, err := h.usecase.CreateMataPelajaran(req.Nama)
	if err != nil {
		return nil, err
	}

	return &base.MataPelajaranResponse{
		MataPelajaran: &base.MataPelajaran{
			Id:   int32(mp.ID),
			Nama: mp.Nama,
		},
	}, nil
}

// GetMataPelajaran gets mata pelajaran by ID
func (h *mataPelajaranHandler) GetMataPelajaran(ctx context.Context, req *base.GetMataPelajaranRequest) (*base.MataPelajaranResponse, error) {
	mp, err := h.usecase.GetMataPelajaran(int(req.Id))
	if err != nil {
		return nil, err
	}

	return &base.MataPelajaranResponse{
		MataPelajaran: &base.MataPelajaran{
			Id:   int32(mp.ID),
			Nama: mp.Nama,
		},
	}, nil
}

// UpdateMataPelajaran updates mata pelajaran
func (h *mataPelajaranHandler) UpdateMataPelajaran(ctx context.Context, req *base.UpdateMataPelajaranRequest) (*base.MataPelajaranResponse, error) {
	mp, err := h.usecase.UpdateMataPelajaran(int(req.Id), req.Nama)
	if err != nil {
		return nil, err
	}

	return &base.MataPelajaranResponse{
		MataPelajaran: &base.MataPelajaran{
			Id:   int32(mp.ID),
			Nama: mp.Nama,
		},
	}, nil
}

// DeleteMataPelajaran deletes mata pelajaran
func (h *mataPelajaranHandler) DeleteMataPelajaran(ctx context.Context, req *base.DeleteMataPelajaranRequest) (*base.MessageStatusResponse, error) {
	err := h.usecase.DeleteMataPelajaran(int(req.Id))
	if err != nil {
		return &base.MessageStatusResponse{
			Message: "Failed to delete mata pelajaran",
			Status:  "error",
		}, err
	}

	return &base.MessageStatusResponse{
		Message: "Mata pelajaran deleted successfully",
		Status:  "success",
	}, nil
}

// ListMataPelajaran lists mata pelajaran
func (h *mataPelajaranHandler) ListMataPelajaran(ctx context.Context, req *emptypb.Empty) (*base.ListMataPelajaranResponse, error) {
	mps, _, err := h.usecase.ListMataPelajaran(1, 1000) // Get all by default
	if err != nil {
		return nil, err
	}

	var mataPelajarans []*base.MataPelajaran
	for _, mp := range mps {
		var lmsSubjectID int64
		if mp.LmsSubjectID != nil {
			lmsSubjectID = *mp.LmsSubjectID
		}
		mataPelajarans = append(mataPelajarans, &base.MataPelajaran{
			Id:           int32(mp.ID),
			Nama:         mp.Nama,
			LmsSubjectId: lmsSubjectID,
		})
	}

	return &base.ListMataPelajaranResponse{
		MataPelajaran: mataPelajarans,
	}, nil
}