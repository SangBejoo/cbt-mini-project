package notes

import (
	"context"
	"errors"

	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/repository/notes"
)

type notesUseCase struct {
	repo notes.NotesRepository
}

func NewNotesUseCase(repo notes.NotesRepository) NotesUseCase {
	return &notesUseCase{repo: repo}
}

// Implementasi CreateNote di usecase
func (u *notesUseCase) CreateNote(ctx context.Context, note *entity.Note) error {
	// Validasi sederhana
	if note.Title == "" {
		return errors.New("title tidak boleh kosong")
	}
	return u.repo.CreateNote(ctx, note)
}

// Implementasi GetNotes di usecase
func (u *notesUseCase) GetNotes(ctx context.Context) ([]*entity.Note, error) {
	notes, err := u.repo.GetNotes(ctx)
	if err != nil {
		return nil, err
	}
	return notes, nil
}

// Implementasi UpdateNote di usecase
func (u *notesUseCase) UpdateNote(ctx context.Context, note *entity.Note) error {
	// Validasi sederhana
	if note.Title == "" {
		return errors.New("title tidak boleh kosong")
	}
	return u.repo.UpdateNote(ctx, note)
}

// Implementasi DeleteNote di usecase
func (u *notesUseCase) DeleteNote(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("id tidak valid")
	}
	return u.repo.DeleteNote(ctx, id)
}
