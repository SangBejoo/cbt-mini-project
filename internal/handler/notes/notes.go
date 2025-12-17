package notes

import (
	"context"
	"errors"
	"fmt"

	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/usecase/notes"

	"google.golang.org/protobuf/types/known/emptypb"
)

type notesGRPCHandler struct {
	base.UnimplementedNotesServiceServer
	usecase notes.NotesUseCase
}

func NewNotesHandler(usecase notes.NotesUseCase) *notesGRPCHandler {
	return &notesGRPCHandler{usecase: usecase}
}

// Implementasi CreateNote di handler
func (h *notesGRPCHandler) CreateNote(ctx context.Context, req *base.NotesRequest) (*base.NotesResponse, error) {
    noteProto := req.GetNote()
    note := &entity.Note{
        Title:   noteProto.GetTitle(),
        Content: noteProto.GetContent(),
        // CreatedAt bisa diisi jika perlu
    }
    err := h.usecase.CreateNote(ctx, note)
    if err != nil {
        fmt.Printf("[CreateNote] error: %v\n", err)
        return nil, err
    }
    // Kembalikan NotesResponse, misal ambil semua notes setelah insert
    notes, err := h.usecase.GetNotes(ctx)
    if err != nil {
        return nil, err
    }
    var protoNotes []*base.Notes
    for _, n := range notes {
        protoNotes = append(protoNotes, &base.Notes{
            Id:        int32(n.ID),
            Title:     n.Title,
            Content:   n.Content,
            // CreatedAt: ... (mapping time ke timestamp jika perlu)
        })
    }
    return &base.NotesResponse{Notes: protoNotes}, nil
}
// Implementasi GetNotes di handler
func (h *notesGRPCHandler) GetNotes(ctx context.Context, req *emptypb.Empty) (*base.NotesResponse, error) {
    notes, err := h.usecase.GetNotes(ctx)
    if err != nil {
        fmt.Printf("[GetNotes] error: %v\n", err)
        return nil, err
    }
    var protoNotes []*base.Notes
    for _, n := range notes {
        protoNotes = append(protoNotes, &base.Notes{
            Id:        int32(n.ID),
            Title:     n.Title,
            Content:   n.Content,
            // CreatedAt: ... (mapping time ke timestamp jika perlu)
        })
    }
    return &base.NotesResponse{Notes: protoNotes}, nil
}

// Implementasi UpdateNote di handler
// Implementasi UpdateNote di handler
func (h *notesGRPCHandler) UpdateNote(ctx context.Context, req *base.UpdateNoteRequest) (*base.NotesResponse, error) {
    note := &entity.Note{
        ID:      int(req.GetId()),
        Title:   req.GetTitle(),
        Content: req.GetContent(),
    }
    // Log input
    fmt.Printf("[UpdateNote] id: %d, title: %s, content: %s\n", note.ID, note.Title, note.Content)

    // Validasi sederhana
    if note.Title == "" {
        fmt.Println("[UpdateNote] error: title tidak boleh kosong")
        return nil, errors.New("title tidak boleh kosong")
    }

    // Lanjut ke usecase
    err := h.usecase.UpdateNote(ctx, note)
    if err != nil {
        return nil, err
    }
    // Setelah update, ambil semua notes dan kembalikan
    notes, err := h.usecase.GetNotes(ctx)
    if err != nil {
        return nil, err
    }
    var protoNotes []*base.Notes
    for _, n := range notes {
        protoNotes = append(protoNotes, &base.Notes{
            Id:        int32(n.ID),
            Title:     n.Title,
            Content:   n.Content,
            // CreatedAt: ... (mapping time ke timestamp jika perlu)
        })
    }
    return &base.NotesResponse{Notes: protoNotes}, nil
}

// Implementasi DeleteNote di handler
func (h *notesGRPCHandler) DeleteNote(ctx context.Context, req *base.DeleteNoteRequest) (*base.MessageStatusResponse, error) {
    id := int(req.GetId())
    err := h.usecase.DeleteNote(ctx, id)
    if err != nil {
        return &base.MessageStatusResponse{
            Status:  "ERROR",
            Message: err.Error(),
        }, err
    }
    return &base.MessageStatusResponse{
        Status:  "OK",
        Message: "Note deleted",
    }, nil
}

