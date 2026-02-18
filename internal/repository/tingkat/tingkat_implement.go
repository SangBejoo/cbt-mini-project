package tingkat

import (
	"cbt-test-mini-project/internal/entity"
	"database/sql"
)

// tingkatRepositoryImpl implements TingkatRepository
type tingkatRepositoryImpl struct {
	db *sql.DB
}

// NewTingkatRepository creates a new TingkatRepository instance
func NewTingkatRepository(db *sql.DB) TingkatRepository {
	return &tingkatRepositoryImpl{db: db}
}

// Create a new tingkat
func (r *tingkatRepositoryImpl) Create(t *entity.Tingkat) error {
	query := `
		INSERT INTO grade_levels (name, is_active, lms_level_id)
		VALUES ($1, $2, $3)
		RETURNING id`
	return r.db.QueryRow(query, t.Nama, t.IsActive, t.LmsLevelID).Scan(&t.ID)
}

// Get by ID
func (r *tingkatRepositoryImpl) GetByID(id int) (*entity.Tingkat, error) {
	var t entity.Tingkat
	query := `SELECT id, name, is_active, lms_level_id FROM grade_levels WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&t.ID, &t.Nama, &t.IsActive, &t.LmsLevelID)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Update existing
func (r *tingkatRepositoryImpl) Update(t *entity.Tingkat) error {
	query := `
		UPDATE grade_levels
		SET name = $1, is_active = $2, lms_level_id = $3
		WHERE id = $4`
	_, err := r.db.Exec(query, t.Nama, t.IsActive, t.LmsLevelID, t.ID)
	return err
}

// Delete by ID (soft delete)
func (r *tingkatRepositoryImpl) Delete(id int) error {
	query := `UPDATE grade_levels SET is_active = false WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// List all
func (r *tingkatRepositoryImpl) List(limit, offset int) ([]entity.Tingkat, int, error) {
	var tingkats []entity.Tingkat

	// Get total count
	var total int
	countQuery := `SELECT COUNT(*) FROM grade_levels`
	err := r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	query := `SELECT id, name, is_active, lms_level_id FROM grade_levels ORDER BY id LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var t entity.Tingkat
		err := rows.Scan(&t.ID, &t.Nama, &t.IsActive, &t.LmsLevelID)
		if err != nil {
			return nil, 0, err
		}
		tingkats = append(tingkats, t)
	}

	return tingkats, total, nil
}

// UpsertByLMSID inserts or updates a tingkat by LMS ID
func (r *tingkatRepositoryImpl) UpsertByLMSID(lmsID int64, name string) error {
	query := `
		WITH updated AS (
			UPDATE grade_levels
			SET name = $1,
				is_active = true,
				updated_at = CURRENT_TIMESTAMP
			WHERE lms_level_id = $2
			RETURNING id
		)
		INSERT INTO grade_levels (name, is_active, lms_level_id, created_at, updated_at)
		SELECT $1, true, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		WHERE NOT EXISTS (SELECT 1 FROM updated)`
	_, err := r.db.Exec(query, name, lmsID)
	return err
}

// DeleteByLMSID soft deletes a tingkat by LMS ID
func (r *tingkatRepositoryImpl) DeleteByLMSID(lmsID int64) error {
	query := `UPDATE grade_levels SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE lms_level_id = $1`
	_, err := r.db.Exec(query, lmsID)
	return err
}