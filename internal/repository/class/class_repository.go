package class

import (
	"database/sql"

	"cbt-test-mini-project/internal/entity"
)

// ClassRepository defines the interface for class operations
type ClassRepository interface {
	UpsertByLMSID(lmsClassID, lmsSchoolID int64, name string, isActive bool) error
	GetByLMSID(lmsClassID int64) (*entity.Class, error)
	DeleteByLMSID(lmsClassID int64) error
	List() ([]entity.Class, error)
}

type classRepository struct {
	db *sql.DB
}

// NewClassRepository creates a new class repository
func NewClassRepository(db *sql.DB) ClassRepository {
	return &classRepository{db: db}
}

// UpsertByLMSID creates or updates a class by LMS ID
func (r *classRepository) UpsertByLMSID(lmsClassID, lmsSchoolID int64, name string, isActive bool) error {
	query := `
		UPDATE classes
		SET school_id = $2,
		    name = $3,
		    status = CASE WHEN $4 THEN 'active' ELSE 'inactive' END,
		    deleted_at = CASE WHEN $4 THEN NULL ELSE COALESCE(deleted_at, NOW()) END
		WHERE id = $1
	`
	_, err := r.db.Exec(query, lmsClassID, lmsSchoolID, name, isActive)
	return err
}

// GetByLMSID retrieves a class by its LMS ID
func (r *classRepository) GetByLMSID(lmsClassID int64) (*entity.Class, error) {
	var c entity.Class
	query := `
		SELECT id,
		       id AS lms_class_id,
		       school_id AS lms_school_id,
		       name,
		       (deleted_at IS NULL AND COALESCE(status, 'active') <> 'inactive') AS is_active,
		       COALESCE(created_at, NOW()) AS created_at,
		       COALESCE(created_at, NOW()) AS updated_at
		FROM classes
		WHERE id = $1`
	err := r.db.QueryRow(query, lmsClassID).Scan(&c.ID, &c.LMSClassID, &c.LMSSchoolID, &c.Name, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// DeleteByLMSID deletes a class by its LMS ID
func (r *classRepository) DeleteByLMSID(lmsClassID int64) error {
	query := `UPDATE classes SET status = 'inactive', deleted_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, lmsClassID)
	return err
}

// List retrieves all active classes
func (r *classRepository) List() ([]entity.Class, error) {
	query := `
		SELECT id,
		       id AS lms_class_id,
		       school_id AS lms_school_id,
		       name,
		       true AS is_active,
		       COALESCE(created_at, NOW()) AS created_at,
		       COALESCE(created_at, NOW()) AS updated_at
		FROM classes
		WHERE deleted_at IS NULL
		  AND COALESCE(status, 'active') <> 'inactive'
		ORDER BY name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []entity.Class
	for rows.Next() {
		var c entity.Class
		if err := rows.Scan(&c.ID, &c.LMSClassID, &c.LMSSchoolID, &c.Name, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}
	return classes, rows.Err()
}
