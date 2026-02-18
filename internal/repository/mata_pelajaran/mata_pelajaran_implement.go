package mata_pelajaran

import (
	"cbt-test-mini-project/internal/entity"
	"database/sql"
)

// mataPelajaranRepositoryImpl implements MataPelajaranRepository
type mataPelajaranRepositoryImpl struct {
	db *sql.DB
}

// NewMataPelajaranRepository creates a new MataPelajaranRepository instance
func NewMataPelajaranRepository(db *sql.DB) MataPelajaranRepository {
	return &mataPelajaranRepositoryImpl{db: db}
}

// Create a new mata pelajaran
func (r *mataPelajaranRepositoryImpl) Create(mp *entity.MataPelajaran) error {
	query := `INSERT INTO subjects (name, lms_subject_id, lms_school_id, lms_class_id, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) RETURNING id`
	err := r.db.QueryRow(query, mp.Nama, mp.LmsSubjectID, mp.LmsSchoolID, mp.LmsClassID).Scan(&mp.ID)
	return err
}

// Get by ID
func (r *mataPelajaranRepositoryImpl) GetByID(id int) (*entity.MataPelajaran, error) {
	var mp entity.MataPelajaran
	query := `SELECT id, name, lms_subject_id, lms_school_id, lms_class_id FROM subjects WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&mp.ID, &mp.Nama, &mp.LmsSubjectID, &mp.LmsSchoolID, &mp.LmsClassID)
	if err != nil {
		return nil, err
	}
	return &mp, nil
}

// Update existing
func (r *mataPelajaranRepositoryImpl) Update(mp *entity.MataPelajaran) error {
	query := `UPDATE subjects SET name = $1, lms_subject_id = $2, lms_school_id = $3, lms_class_id = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5`
	_, err := r.db.Exec(query, mp.Nama, mp.LmsSubjectID, mp.LmsSchoolID, mp.LmsClassID, mp.ID)
	return err
}

// Delete by ID (soft delete)
func (r *mataPelajaranRepositoryImpl) Delete(id int) error {
	query := `UPDATE subjects SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// List all
func (r *mataPelajaranRepositoryImpl) List(limit, offset int) ([]entity.MataPelajaran, int, error) {
	var mps []entity.MataPelajaran
	var total int

	// Get total count
	countQuery := `SELECT COUNT(*) FROM subjects WHERE is_active = true`
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get list
	listQuery := `SELECT id, name, lms_subject_id, lms_school_id, lms_class_id FROM subjects WHERE is_active = true ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(listQuery, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var mp entity.MataPelajaran
		err := rows.Scan(&mp.ID, &mp.Nama, &mp.LmsSubjectID, &mp.LmsSchoolID, &mp.LmsClassID)
		if err != nil {
			return nil, 0, err
		}
		mps = append(mps, mp)
	}

	return mps, total, nil
}

// Get by name
func (r *mataPelajaranRepositoryImpl) GetByName(name string) (*entity.MataPelajaran, error) {
	var mp entity.MataPelajaran
	query := `SELECT id, name, lms_subject_id, lms_school_id, lms_class_id FROM subjects WHERE name = $1 AND is_active = true`
	err := r.db.QueryRow(query, name).Scan(&mp.ID, &mp.Nama, &mp.LmsSubjectID, &mp.LmsSchoolID, &mp.LmsClassID)
	if err != nil {
		return nil, err
	}
	return &mp, nil
}

// UpsertByLMSID inserts or updates based on LMS ID
func (r *mataPelajaranRepositoryImpl) UpsertByLMSID(lmsID int64, name string, schoolID int64) error {
	query := `
		WITH updated AS (
			UPDATE subjects
			SET name = $2,
				lms_school_id = $3,
				is_active = true,
				updated_at = CURRENT_TIMESTAMP
			WHERE lms_subject_id = $1
			RETURNING id
		)
		INSERT INTO subjects (lms_subject_id, name, lms_school_id, is_active, created_at, updated_at)
		SELECT $1, $2, $3, true, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		WHERE NOT EXISTS (SELECT 1 FROM updated)
	`
	_, err := r.db.Exec(query, lmsID, name, schoolID)
	return err
}

// DeleteByLMSID soft deletes by LMS ID
func (r *mataPelajaranRepositoryImpl) DeleteByLMSID(lmsID int64) error {
	query := `UPDATE subjects SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE lms_subject_id = $1`
	_, err := r.db.Exec(query, lmsID)
	return err
}