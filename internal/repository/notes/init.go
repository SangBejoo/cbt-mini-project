package notes

import (
	"context"

	"cbt-test-mini-project/internal/entity"
)

// NotesRepository adalah interface untuk operasi notes
type NotesRepository interface {
	CreateNote(ctx context.Context, note *entity.Note) error
	GetNotes(ctx context.Context) ([]*entity.Note, error)
	UpdateNote(ctx context.Context, note *entity.Note) error
	DeleteNote(ctx context.Context, id int) error
}
