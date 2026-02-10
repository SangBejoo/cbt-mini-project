package tingkat

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/usecase/tingkat"
	"context"

	"google.golang.org/protobuf/types/known/emptypb"
)

// tingkatHandler implements base.TingkatServiceServer
type tingkatHandler struct {
	base.UnimplementedTingkatServiceServer
	usecase tingkat.TingkatUsecase
}

// NewTingkatHandler creates a new TingkatHandler
func NewTingkatHandler(usecase tingkat.TingkatUsecase) base.TingkatServiceServer {
	return &tingkatHandler{usecase: usecase}
}

// CreateTingkat creates a new tingkat
func (h *tingkatHandler) CreateTingkat(ctx context.Context, req *base.CreateTingkatRequest) (*base.TingkatResponse, error) {
	t, err := h.usecase.CreateTingkat(req.Nama)
	if err != nil {
		return nil, err
	}

	return &base.TingkatResponse{
		Tingkat: &base.Tingkat{
			Id:   int32(t.ID),
			Nama: t.Nama,
		},
	}, nil
}

// GetTingkat gets tingkat by ID
func (h *tingkatHandler) GetTingkat(ctx context.Context, req *base.GetTingkatRequest) (*base.TingkatResponse, error) {
	t, err := h.usecase.GetTingkat(int(req.Id))
	if err != nil {
		return nil, err
	}

	return &base.TingkatResponse{
		Tingkat: &base.Tingkat{
			Id:   int32(t.ID),
			Nama: t.Nama,
		},
	}, nil
}

// UpdateTingkat updates a tingkat
func (h *tingkatHandler) UpdateTingkat(ctx context.Context, req *base.UpdateTingkatRequest) (*base.TingkatResponse, error) {
	t, err := h.usecase.UpdateTingkat(int(req.Id), req.Nama)
	if err != nil {
		return nil, err
	}

	return &base.TingkatResponse{
		Tingkat: &base.Tingkat{
			Id:   int32(t.ID),
			Nama: t.Nama,
		},
	}, nil
}

// DeleteTingkat deletes a tingkat
func (h *tingkatHandler) DeleteTingkat(ctx context.Context, req *base.DeleteTingkatRequest) (*base.MessageStatusResponse, error) {
	err := h.usecase.DeleteTingkat(int(req.Id))
	if err != nil {
		return &base.MessageStatusResponse{
			Message: "Failed to delete tingkat",
			Status:  "error",
		}, err
	}

	return &base.MessageStatusResponse{
		Message: "Tingkat deleted successfully",
		Status:  "success",
	}, nil
}

// ListTingkat lists tingkat
func (h *tingkatHandler) ListTingkat(ctx context.Context, req *emptypb.Empty) (*base.ListTingkatResponse, error) {
	tingkats, _, err := h.usecase.ListTingkat(1, 1000) // Get all by default
	if err != nil {
		return nil, err
	}

	var tingkatList []*base.Tingkat
	for _, t := range tingkats {
		var lmsLevelID int64
		if t.LmsLevelID != nil {
			lmsLevelID = *t.LmsLevelID
		}
		tingkatList = append(tingkatList, &base.Tingkat{
			Id:         int32(t.ID),
			Nama:       t.Nama,
			LmsLevelId: lmsLevelID,
		})
	}

	return &base.ListTingkatResponse{
		Tingkat: tingkatList,
	}, nil
}