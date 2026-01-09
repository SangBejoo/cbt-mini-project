package soal_drag_drop

import (
	"context"

	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	usecase "cbt-test-mini-project/internal/usecase/soal_drag_drop"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// grpcHandler implements base.SoalDragDropServiceServer
type grpcHandler struct {
	base.UnimplementedSoalDragDropServiceServer
	usecase usecase.Usecase
}

// NewGrpcHandler creates a new SoalDragDropServiceServer
func NewGrpcHandler(uc usecase.Usecase) base.SoalDragDropServiceServer {
	return &grpcHandler{usecase: uc}
}

// CreateSoalDragDrop creates a new drag-drop question
func (h *grpcHandler) CreateSoalDragDrop(ctx context.Context, req *base.CreateSoalDragDropRequest) (*base.SoalDragDropResponse, error) {
	// Convert proto to usecase request
	ucReq := &usecase.CreateRequest{
		IDMateri:   int(req.IdMateri),
		IDTingkat:  int(req.IdTingkat),
		Pertanyaan: req.Pertanyaan,
		DragType:   protoToEntityDragType(req.DragType),
	}

	if req.Pembahasan != "" {
		ucReq.Pembahasan = &req.Pembahasan
	}

	for _, item := range req.Items {
		var imageURL *string
		if item.ImageUrl != "" {
			imageURL = &item.ImageUrl
		}
		ucReq.Items = append(ucReq.Items, usecase.ItemRequest{
			Label:    item.Label,
			ImageURL: imageURL,
			Urutan:   int(item.Urutan),
		})
	}

	for _, slot := range req.Slots {
		var imageURL *string
		if slot.ImageUrl != "" {
			imageURL = &slot.ImageUrl
		}
		ucReq.Slots = append(ucReq.Slots, usecase.SlotRequest{
			Label:    slot.Label,
			ImageURL: imageURL,
			Urutan:   int(slot.Urutan),
		})
	}

	for _, ca := range req.CorrectAnswers {
		ucReq.CorrectAnswers = append(ucReq.CorrectAnswers, usecase.CorrectAnswerRequest{
			ItemUrutan: int(ca.ItemUrutan),
			SlotUrutan: int(ca.SlotUrutan),
		})
	}

	soal, err := h.usecase.Create(ucReq)
	if err != nil {
		return nil, err
	}

	return h.toProtoResponse(soal)
}

// GetSoalDragDrop gets a drag-drop question by ID
func (h *grpcHandler) GetSoalDragDrop(ctx context.Context, req *base.GetSoalDragDropRequest) (*base.SoalDragDropResponse, error) {
	soal, err := h.usecase.GetByID(int(req.Id))
	if err != nil {
		return nil, err
	}
	if soal == nil {
		return nil, err
	}

	return h.toProtoResponse(soal)
}

// UpdateSoalDragDrop updates a drag-drop question
func (h *grpcHandler) UpdateSoalDragDrop(ctx context.Context, req *base.UpdateSoalDragDropRequest) (*base.SoalDragDropResponse, error) {
	ucReq := &usecase.UpdateRequest{
		IDMateri:   int(req.IdMateri),
		IDTingkat:  int(req.IdTingkat),
		Pertanyaan: req.Pertanyaan,
		DragType:   protoToEntityDragType(req.DragType),
		IsActive:   req.IsActive,
	}

	if req.Pembahasan != "" {
		ucReq.Pembahasan = &req.Pembahasan
	}

	for _, item := range req.Items {
		var imageURL *string
		if item.ImageUrl != "" {
			imageURL = &item.ImageUrl
		}
		ucReq.Items = append(ucReq.Items, usecase.ItemRequest{
			Label:    item.Label,
			ImageURL: imageURL,
			Urutan:   int(item.Urutan),
		})
	}

	for _, slot := range req.Slots {
		var imageURL *string
		if slot.ImageUrl != "" {
			imageURL = &slot.ImageUrl
		}
		ucReq.Slots = append(ucReq.Slots, usecase.SlotRequest{
			Label:    slot.Label,
			ImageURL: imageURL,
			Urutan:   int(slot.Urutan),
		})
	}

	for _, ca := range req.CorrectAnswers {
		ucReq.CorrectAnswers = append(ucReq.CorrectAnswers, usecase.CorrectAnswerRequest{
			ItemUrutan: int(ca.ItemUrutan),
			SlotUrutan: int(ca.SlotUrutan),
		})
	}

	soal, err := h.usecase.Update(int(req.Id), ucReq)
	if err != nil {
		return nil, err
	}

	return h.toProtoResponse(soal)
}

// DeleteSoalDragDrop deletes a drag-drop question
func (h *grpcHandler) DeleteSoalDragDrop(ctx context.Context, req *base.DeleteSoalDragDropRequest) (*base.MessageStatusResponse, error) {
	err := h.usecase.Delete(int(req.Id))
	if err != nil {
		return nil, err
	}

	return &base.MessageStatusResponse{
		Message: "Drag-drop question deleted successfully",
		Status:  "success",
	}, nil
}

// ListSoalDragDrop lists drag-drop questions
func (h *grpcHandler) ListSoalDragDrop(ctx context.Context, req *base.ListSoalDragDropRequest) (*base.ListSoalDragDropResponse, error) {
	page := 1
	pageSize := 100
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = int(req.Pagination.Page)
		}
		if req.Pagination.PageSize > 0 {
			pageSize = int(req.Pagination.PageSize)
		}
	}

	soals, total, err := h.usecase.List(int(req.IdMateri), int(req.IdTingkat), page, pageSize)
	if err != nil {
		return nil, err
	}

	var protoSoals []*base.SoalDragDropFull
	for _, s := range soals {
		protoSoal, err := h.entityToProto(&s)
		if err == nil {
			protoSoals = append(protoSoals, protoSoal)
		}
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	return &base.ListSoalDragDropResponse{
		Soal: protoSoals,
		Pagination: &base.PaginationResponse{
			TotalCount:  int32(total),
			TotalPages:  int32(totalPages),
			CurrentPage: int32(page),
			PageSize:    int32(pageSize),
		},
	}, nil
}

// Helper functions

func (h *grpcHandler) toProtoResponse(soal *entity.SoalDragDrop) (*base.SoalDragDropResponse, error) {
	protoSoal, err := h.entityToProto(soal)
	if err != nil {
		return nil, err
	}
	return &base.SoalDragDropResponse{Soal: protoSoal}, nil
}

func (h *grpcHandler) entityToProto(soal *entity.SoalDragDrop) (*base.SoalDragDropFull, error) {
	// Get correct answers
	_, correctAnswers, _ := h.usecase.GetByIDWithCorrectAnswers(soal.ID)

	// Convert items
	var protoItems []*base.DragItem
	for _, item := range soal.Items {
		imageURL := ""
		if item.ImageURL != nil {
			imageURL = *item.ImageURL
		}
		protoItems = append(protoItems, &base.DragItem{
			Id:       int32(item.ID),
			Label:    item.Label,
			ImageUrl: imageURL,
			Urutan:   int32(item.Urutan),
		})
	}

	// Convert slots
	var protoSlots []*base.DragSlot
	for _, slot := range soal.Slots {
		imageURL := ""
		if slot.ImageURL != nil {
			imageURL = *slot.ImageURL
		}
		protoSlots = append(protoSlots, &base.DragSlot{
			Id:       int32(slot.ID),
			Label:    slot.Label,
			ImageUrl: imageURL,
			Urutan:   int32(slot.Urutan),
		})
	}

	// Convert correct answers
	var protoAnswers []*base.DragCorrectAnswer
	for _, ca := range correctAnswers {
		protoAnswers = append(protoAnswers, &base.DragCorrectAnswer{
			ItemId: int32(ca.IDDragItem),
			SlotId: int32(ca.IDDragSlot),
		})
	}

	pembahasan := ""
	if soal.Pembahasan != nil {
		pembahasan = *soal.Pembahasan
	}

	return &base.SoalDragDropFull{
		Id: int32(soal.ID),
		Materi: &base.Materi{
			Id:   int32(soal.Materi.ID),
			Nama: soal.Materi.Nama,
			MataPelajaran: &base.MataPelajaran{
				Id:   int32(soal.Materi.MataPelajaran.ID),
				Nama: soal.Materi.MataPelajaran.Nama,
			},
			Tingkat: &base.Tingkat{
				Id:   int32(soal.Materi.Tingkat.ID),
				Nama: soal.Materi.Tingkat.Nama,
			},
		},
		Pertanyaan:     soal.Pertanyaan,
		DragType:       entityToProtoDragType(soal.DragType),
		Items:          protoItems,
		Slots:          protoSlots,
		CorrectAnswers: protoAnswers,
		Pembahasan:     pembahasan,
		IsActive:       soal.IsActive,
		CreatedAt:      timestamppb.New(soal.CreatedAt),
		UpdatedAt:      timestamppb.New(soal.UpdatedAt),
	}, nil
}

func protoToEntityDragType(dt base.DragDropType) entity.DragDropType {
	switch dt {
	case base.DragDropType_ORDERING:
		return entity.DragTypeOrdering
	case base.DragDropType_MATCHING:
		return entity.DragTypeMatching
	default:
		return entity.DragTypeOrdering
	}
}

func entityToProtoDragType(dt entity.DragDropType) base.DragDropType {
	switch dt {
	case entity.DragTypeOrdering:
		return base.DragDropType_ORDERING
	case entity.DragTypeMatching:
		return base.DragDropType_MATCHING
	default:
		return base.DragDropType_ORDERING
	}
}
