package notes

import (
	"context"
	"database/sql"

	"cbt-test-mini-project/internal/entity"
)

// notesRepo adalah implementasi dari NotesRepository
type notesRepo struct {
	db *sql.DB
}

func NewNotesRepository(db *sql.DB) NotesRepository {
	return &notesRepo{db: db}
}

// Implementasi CreateNote untuk notesRepo
func (r *notesRepo) CreateNote(ctx context.Context, note *entity.Note) error {
	query := `INSERT INTO notes (title, content) VALUES ($1, $2)`
	_, err := r.db.ExecContext(ctx, query, note.Title, note.Content)
	return err
}
// Implementasi GetNotes untuk notesRepo
func (r *notesRepo) GetNotes(ctx context.Context) ([]*entity.Note, error) {
	query := `SELECT id, title, content, created_at FROM notes`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*entity.Note
	for rows.Next() {
		note := &entity.Note{}
		if err := rows.Scan(&note.ID, &note.Title, &note.Content, &note.CreatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	return notes, nil
}

// Implementasi UpdateNote untuk notesRepo
func (r *notesRepo) UpdateNote(ctx context.Context, note *entity.Note) error {
	query := `UPDATE notes SET title = $1, content = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, note.Title, note.Content, note.ID)
	return err
}
// Implementasi DeleteNote untuk notesRepo
func (r *notesRepo) DeleteNote(ctx context.Context, id int) error {
	query := `DELETE FROM notes WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
