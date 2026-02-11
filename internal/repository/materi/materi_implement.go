package materi

import (
	"cbt-test-mini-project/internal/entity"
	"database/sql"
	"encoding/json"
	"fmt"
)

// materiRepositoryImpl implements MateriRepository
type materiRepositoryImpl struct {
	db *sql.DB
}

// NewMateriRepository creates a new MateriRepository instance
func NewMateriRepository(db *sql.DB) MateriRepository {
	return &materiRepositoryImpl{db: db}
}

// Create a new materi
func (r *materiRepositoryImpl) Create(materi *entity.Materi) error {
query := `INSERT INTO materi (id_mata_pelajaran, id_tingkat, nama, is_active, default_durasi_menit, default_jumlah_soal, lms_module_id, lms_class_id, owner_user_id, school_id, labels) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`

	labelsJSON, _ := json.Marshal(materi.Labels)
	err := r.db.QueryRow(query, materi.IDMataPelajaran, materi.IDTingkat, materi.Nama, materi.IsActive, materi.DefaultDurasiMenit, materi.DefaultJumlahSoal, materi.LmsModuleID, materi.LmsClassID, materi.OwnerUserID, materi.SchoolID, string(labelsJSON)).Scan(&materi.ID)
	return err
}

// Get by ID
func (r *materiRepositoryImpl) GetByID(id int) (*entity.Materi, error) {
	var materi entity.Materi
	query := `
		SELECT m.id, m.id_mata_pelajaran, m.id_tingkat, m.nama, m.is_active, m.default_durasi_menit, m.default_jumlah_soal, m.lms_module_id, m.lms_class_id, m.owner_user_id, m.school_id, m.labels,
		       mp.id, mp.nama, mp.lms_subject_id, mp.lms_school_id, mp.lms_class_id,
		       t.id, t.nama, t.lms_level_id
		FROM materi m
		JOIN mata_pelajaran mp ON m.id_mata_pelajaran = mp.id
		JOIN tingkat t ON m.id_tingkat = t.id
		WHERE m.id = $1 AND m.is_active = true
		`
	var labelsSQL []byte
	err := r.db.QueryRow(query, id).Scan(
		&materi.ID, &materi.IDMataPelajaran, &materi.IDTingkat, &materi.Nama, &materi.IsActive, &materi.DefaultDurasiMenit, &materi.DefaultJumlahSoal, &materi.LmsModuleID, &materi.LmsClassID, &materi.OwnerUserID, &materi.SchoolID, &labelsSQL,
		&materi.MataPelajaran.ID, &materi.MataPelajaran.Nama, &materi.MataPelajaran.LmsSubjectID, &materi.MataPelajaran.LmsSchoolID, &materi.MataPelajaran.LmsClassID,
		&materi.Tingkat.ID, &materi.Tingkat.Nama, &materi.Tingkat.LmsLevelID,
	)
	if err != nil {
		return nil, err
	}
	if labelsSQL != nil {
		var ls []string
		if err := json.Unmarshal(labelsSQL, &ls); err == nil {
			materi.Labels = ls
		}
	}
	return &materi, nil
}

// Update existing
func (r *materiRepositoryImpl) Update(materi *entity.Materi) error {
	query := `UPDATE materi SET id_mata_pelajaran = $1, id_tingkat = $2, nama = $3, is_active = $4, default_durasi_menit = $5, default_jumlah_soal = $6, lms_module_id = $7, lms_class_id = $8 WHERE id = $9`
	_, err := r.db.Exec(query, materi.IDMataPelajaran, materi.IDTingkat, materi.Nama, materi.IsActive, materi.DefaultDurasiMenit, materi.DefaultJumlahSoal, materi.LmsModuleID, materi.LmsClassID, materi.ID)
	return err
}

// Delete by ID (soft delete)
func (r *materiRepositoryImpl) Delete(id int) error {
	query := `UPDATE materi SET is_active = false WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// List with filters
func (r *materiRepositoryImpl) List(idMataPelajaran, idTingkat *int, limit, offset int) ([]entity.Materi, int, error) {
	var materis []entity.Materi
	var total int

	// Build WHERE clause
	whereClause := "m.is_active = true"
	args := []interface{}{}
	argCount := 0

	if idMataPelajaran != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND m.id_mata_pelajaran = $%d", argCount)
		args = append(args, *idMataPelajaran)
	}
	if idTingkat != nil {
		argCount++
		whereClause += fmt.Sprintf(" AND m.id_tingkat = $%d", argCount)
		args = append(args, *idTingkat)
	}

	// Count query
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) FROM materi m
		JOIN mata_pelajaran mp ON m.id_mata_pelajaran = mp.id
		JOIN tingkat t ON m.id_tingkat = t.id
		WHERE %s
	`, whereClause)
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// List query
	listQuery := fmt.Sprintf(`
		SELECT m.id, m.id_mata_pelajaran, m.id_tingkat, m.nama, m.is_active, m.default_durasi_menit, m.default_jumlah_soal, m.lms_module_id, m.lms_class_id, m.owner_user_id, m.school_id, m.labels,
		       mp.id, mp.nama, mp.lms_subject_id, mp.lms_school_id, mp.lms_class_id,
		       t.id, t.nama, t.lms_level_id
		FROM materi m
		JOIN mata_pelajaran mp ON m.id_mata_pelajaran = mp.id
		JOIN tingkat t ON m.id_tingkat = t.id
		WHERE %s
		ORDER BY m.id
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount+1, argCount+2)
	args = append(args, limit, offset)

	rows, err := r.db.Query(listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var materi entity.Materi
		var labelsSQL []byte
		err := rows.Scan(
			&materi.ID, &materi.IDMataPelajaran, &materi.IDTingkat, &materi.Nama, &materi.IsActive, &materi.DefaultDurasiMenit, &materi.DefaultJumlahSoal, &materi.LmsModuleID, &materi.LmsClassID, &materi.OwnerUserID, &materi.SchoolID, &labelsSQL,
			&materi.MataPelajaran.ID, &materi.MataPelajaran.Nama, &materi.MataPelajaran.LmsSubjectID, &materi.MataPelajaran.LmsSchoolID, &materi.MataPelajaran.LmsClassID,
			&materi.Tingkat.ID, &materi.Tingkat.Nama, &materi.Tingkat.LmsLevelID,
		)
		if err != nil {
			return nil, 0, err
		}
		if labelsSQL != nil {
			var ls []string
			if err := json.Unmarshal(labelsSQL, &ls); err == nil {
				materi.Labels = ls
			}
		}
		materis = append(materis, materi)
	}

	return materis, total, nil
}

// Get by mata pelajaran ID
func (r *materiRepositoryImpl) GetByMataPelajaranID(idMataPelajaran int) ([]entity.Materi, error) {
	var materis []entity.Materi
	query := `
		SELECT m.id, m.id_mata_pelajaran, m.id_tingkat, m.nama, m.is_active, m.default_durasi_menit, m.default_jumlah_soal, m.lms_module_id, m.lms_class_id, m.owner_user_id, m.school_id, m.labels,
		       mp.id, mp.nama, mp.lms_subject_id, mp.lms_school_id, mp.lms_class_id,
		       t.id, t.nama, t.lms_level_id
		FROM materi m
		JOIN mata_pelajaran mp ON m.id_mata_pelajaran = mp.id
		JOIN tingkat t ON m.id_tingkat = t.id
		WHERE m.id_mata_pelajaran = $1 AND m.is_active = true
		ORDER BY m.id
	`
	rows, err := r.db.Query(query, idMataPelajaran)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var materi entity.Materi
		var labelsSQL []byte
		err := rows.Scan(
			&materi.ID, &materi.IDMataPelajaran, &materi.IDTingkat, &materi.Nama, &materi.IsActive, &materi.DefaultDurasiMenit, &materi.DefaultJumlahSoal, &materi.LmsModuleID, &materi.LmsClassID, &materi.OwnerUserID, &materi.SchoolID, &labelsSQL,
			&materi.MataPelajaran.ID, &materi.MataPelajaran.Nama, &materi.MataPelajaran.LmsSubjectID, &materi.MataPelajaran.LmsSchoolID, &materi.MataPelajaran.LmsClassID,
			&materi.Tingkat.ID, &materi.Tingkat.Nama, &materi.Tingkat.LmsLevelID,
		)
		if err != nil {
			return nil, err
		}
		if labelsSQL != nil {
			var ls []string
			if err := json.Unmarshal(labelsSQL, &ls); err == nil {
				materi.Labels = ls
			}
		}
		materis = append(materis, materi)
	}

	return materis, nil
}

// UpsertByLMSID inserts or updates based on LMS ID
func (r *materiRepositoryImpl) UpsertByLMSID(lmsID int64, subjectID int64, levelID int64, name string) error {
	query := `
		INSERT INTO materi (lms_module_id, nama, id_mata_pelajaran, id_tingkat, is_active)
		VALUES ($1, $2, $3, $4, true)
		ON CONFLICT (lms_module_id)
		DO UPDATE SET
			nama = EXCLUDED.nama,
			id_mata_pelajaran = EXCLUDED.id_mata_pelajaran,
			id_tingkat = EXCLUDED.id_tingkat,
			is_active = true
	`
	_, err := r.db.Exec(query, lmsID, name, subjectID, levelID)
	return err
}

// DeleteByLMSID soft deletes by LMS ID
func (r *materiRepositoryImpl) DeleteByLMSID(lmsID int64) error {
	query := `UPDATE materi SET is_active = false WHERE lms_module_id = $1`
	_, err := r.db.Exec(query, lmsID)
	return err
}