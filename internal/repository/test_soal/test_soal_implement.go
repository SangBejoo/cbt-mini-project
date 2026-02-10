package test_soal

import (
	"cbt-test-mini-project/internal/entity"
	"database/sql"
)

// soalRepositoryImpl implements SoalRepository
type soalRepositoryImpl struct {
	db *sql.DB
}

// NewSoalRepository creates a new SoalRepository instance
func NewSoalRepository(db *sql.DB) SoalRepository {
	return &soalRepositoryImpl{db: db}
}


// Create a new soal
func (r *soalRepositoryImpl) Create(soal *entity.Soal) error {
	query := `
		INSERT INTO soal (id_materi, id_tingkat, pertanyaan, opsi_a, opsi_b, opsi_c, opsi_d, jawaban_benar, pembahasan, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`
	var pembahasan *string
	if soal.Pembahasan != nil {
		pembahasan = soal.Pembahasan
	}
	return r.db.QueryRow(query, soal.IDMateri, soal.IDTingkat, soal.Pertanyaan, soal.OpsiA, soal.OpsiB, soal.OpsiC, soal.OpsiD, string(soal.JawabanBenar), pembahasan, soal.IsActive).Scan(&soal.ID)
}

// Get soal by ID with all relations
func (r *soalRepositoryImpl) GetByID(id int) (*entity.Soal, error) {
	// Get soal with materi, mata_pelajaran, and tingkat
	soalQuery := `
		SELECT s.id, s.id_materi, s.id_tingkat, s.pertanyaan, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, s.jawaban_benar, s.pembahasan, s.is_active,
		       m.id, m.id_mata_pelajaran, m.id_tingkat, m.nama, m.is_active, m.default_durasi_menit, m.default_jumlah_soal, m.lms_module_id, m.lms_class_id,
		       mp.id, mp.nama, mp.is_active, mp.lms_subject_id, mp.lms_school_id, mp.lms_class_id,
		       t.id, t.nama, t.is_active, t.lms_level_id
		FROM soal s
		JOIN materi m ON s.id_materi = m.id
		JOIN mata_pelajaran mp ON m.id_mata_pelajaran = mp.id
		JOIN tingkat t ON m.id_tingkat = t.id
		WHERE s.id = $1`

	var soal entity.Soal
	var pembahasan *string
	err := r.db.QueryRow(soalQuery, id).Scan(
		&soal.ID, &soal.IDMateri, &soal.IDTingkat, &soal.Pertanyaan, &soal.OpsiA, &soal.OpsiB, &soal.OpsiC, &soal.OpsiD, &soal.JawabanBenar, &pembahasan, &soal.IsActive,
		&soal.Materi.ID, &soal.Materi.IDMataPelajaran, &soal.Materi.IDTingkat, &soal.Materi.Nama, &soal.Materi.IsActive, &soal.Materi.DefaultDurasiMenit, &soal.Materi.DefaultJumlahSoal, &soal.Materi.LmsModuleID, &soal.Materi.LmsClassID,
		&soal.Materi.MataPelajaran.ID, &soal.Materi.MataPelajaran.Nama, &soal.Materi.MataPelajaran.IsActive, &soal.Materi.MataPelajaran.LmsSubjectID, &soal.Materi.MataPelajaran.LmsSchoolID, &soal.Materi.MataPelajaran.LmsClassID,
		&soal.Materi.Tingkat.ID, &soal.Materi.Tingkat.Nama, &soal.Materi.Tingkat.IsActive, &soal.Materi.Tingkat.LmsLevelID,
	)
	if err != nil {
		return nil, err
	}
	soal.Pembahasan = pembahasan

	// Get gambar
	gambarQuery := `
		SELECT id, id_soal, nama_file, file_path, file_size, mime_type, urutan, keterangan, cloud_id, public_id, created_at
		FROM soal_gambar
		WHERE id_soal = $1
		ORDER BY urutan ASC`
	rows, err := r.db.Query(gambarQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var gambar entity.SoalGambar
		var keterangan, cloudId, publicId *string
		err := rows.Scan(&gambar.ID, &gambar.IDSoal, &gambar.NamaFile, &gambar.FilePath, &gambar.FileSize, &gambar.MimeType, &gambar.Urutan, &keterangan, &cloudId, &publicId, &gambar.CreatedAt)
		if err != nil {
			return nil, err
		}
		gambar.Keterangan = keterangan
		gambar.CloudId = cloudId
		gambar.PublicId = publicId
		soal.Gambar = append(soal.Gambar, gambar)
	}

	return &soal, nil
}

// Update existing soal
func (r *soalRepositoryImpl) Update(soal *entity.Soal) error {
	query := `
		UPDATE soal
		SET id_materi = $1, id_tingkat = $2, pertanyaan = $3, opsi_a = $4, opsi_b = $5, opsi_c = $6, opsi_d = $7, jawaban_benar = $8, pembahasan = $9, is_active = $10
		WHERE id = $11`
	_, err := r.db.Exec(query, soal.IDMateri, soal.IDTingkat, soal.Pertanyaan, soal.OpsiA, soal.OpsiB, soal.OpsiC, soal.OpsiD, string(soal.JawabanBenar), soal.Pembahasan, soal.IsActive, soal.ID)
	return err
}

// Delete soal by ID (soft delete)
func (r *soalRepositoryImpl) Delete(id int) error {
	query := `UPDATE soal SET is_active = false WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// List soal with filters
func (r *soalRepositoryImpl) List(idMateri, tingkatan, idMataPelajaran *int, limit, offset int) ([]entity.Soal, int, error) {
	var soals []entity.Soal

	// Build WHERE clause
	whereClause := "WHERE s.is_active = true"
	args := []interface{}{}
	argCount := 0

	if idMateri != nil {
		argCount++
		whereClause += " AND s.id_materi = $" + string(rune(argCount+'0'))
		args = append(args, *idMateri)
	}
	if tingkatan != nil {
		argCount++
		whereClause += " AND m.id_tingkat = $" + string(rune(argCount+'0'))
		args = append(args, *tingkatan)
	}
	if idMataPelajaran != nil {
		argCount++
		whereClause += " AND m.id_mata_pelajaran = $" + string(rune(argCount+'0'))
		args = append(args, *idMataPelajaran)
	}

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM soal s
		JOIN materi m ON s.id_materi = m.id
		JOIN mata_pelajaran mp ON m.id_mata_pelajaran = mp.id
		JOIN tingkat t ON m.id_tingkat = t.id
		` + whereClause

	var total int
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results with all relations
	listQuery := `
		SELECT s.id, s.id_materi, s.id_tingkat, s.pertanyaan, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, s.jawaban_benar, s.pembahasan, s.is_active,
		       m.id, m.id_mata_pelajaran, m.id_tingkat, m.nama, m.is_active, m.default_durasi_menit, m.default_jumlah_soal, m.lms_module_id, m.lms_class_id,
		       mp.id, mp.nama, mp.is_active, mp.lms_subject_id, mp.lms_school_id, mp.lms_class_id,
		       t.id, t.nama, t.is_active, t.lms_level_id
		FROM soal s
		JOIN materi m ON s.id_materi = m.id
		JOIN mata_pelajaran mp ON m.id_mata_pelajaran = mp.id
		JOIN tingkat t ON m.id_tingkat = t.id
		` + whereClause + `
		ORDER BY s.id
		LIMIT $` + string(rune(argCount+1+'0')) + ` OFFSET $` + string(rune(argCount+2+'0'))

	args = append(args, limit, offset)
	rows, err := r.db.Query(listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var soal entity.Soal
		var pembahasan *string
		err := rows.Scan(
			&soal.ID, &soal.IDMateri, &soal.IDTingkat, &soal.Pertanyaan, &soal.OpsiA, &soal.OpsiB, &soal.OpsiC, &soal.OpsiD, &soal.JawabanBenar, &pembahasan, &soal.IsActive,
			&soal.Materi.ID, &soal.Materi.IDMataPelajaran, &soal.Materi.IDTingkat, &soal.Materi.Nama, &soal.Materi.IsActive, &soal.Materi.DefaultDurasiMenit, &soal.Materi.DefaultJumlahSoal, &soal.Materi.LmsModuleID, &soal.Materi.LmsClassID,
			&soal.Materi.MataPelajaran.ID, &soal.Materi.MataPelajaran.Nama, &soal.Materi.MataPelajaran.IsActive, &soal.Materi.MataPelajaran.LmsSubjectID, &soal.Materi.MataPelajaran.LmsSchoolID, &soal.Materi.MataPelajaran.LmsClassID,
			&soal.Materi.Tingkat.ID, &soal.Materi.Tingkat.Nama, &soal.Materi.Tingkat.IsActive, &soal.Materi.Tingkat.LmsLevelID,
		)
		if err != nil {
			return nil, 0, err
		}
		soal.Pembahasan = pembahasan

		// Get gambar for this soal
		gambarQuery := `
			SELECT id, id_soal, nama_file, file_path, file_size, mime_type, urutan, keterangan, cloud_id, public_id, created_at
			FROM soal_gambar
			WHERE id_soal = $1
			ORDER BY urutan ASC`
		gambarRows, err := r.db.Query(gambarQuery, soal.ID)
		if err != nil {
			return nil, 0, err
		}

		for gambarRows.Next() {
			var gambar entity.SoalGambar
			var keterangan, cloudId, publicId *string
			err := gambarRows.Scan(&gambar.ID, &gambar.IDSoal, &gambar.NamaFile, &gambar.FilePath, &gambar.FileSize, &gambar.MimeType, &gambar.Urutan, &keterangan, &cloudId, &publicId, &gambar.CreatedAt)
			if err != nil {
				gambarRows.Close()
				return nil, 0, err
			}
			gambar.Keterangan = keterangan
			gambar.CloudId = cloudId
			gambar.PublicId = publicId
			soal.Gambar = append(soal.Gambar, gambar)
		}
		gambarRows.Close()

		soals = append(soals, soal)
	}

	return soals, total, nil
}

// Get soal by materi ID
func (r *soalRepositoryImpl) GetByMateriID(idMateri int) ([]entity.Soal, error) {
	var soals []entity.Soal

	query := `
		SELECT s.id, s.id_materi, s.id_tingkat, s.pertanyaan, s.opsi_a, s.opsi_b, s.opsi_c, s.opsi_d, s.jawaban_benar, s.pembahasan, s.is_active,
		       m.id, m.id_mata_pelajaran, m.id_tingkat, m.nama, m.is_active, m.default_durasi_menit, m.default_jumlah_soal, m.lms_module_id, m.lms_class_id,
		       mp.id, mp.nama, mp.is_active, mp.lms_subject_id, mp.lms_school_id, mp.lms_class_id,
		       t.id, t.nama, t.is_active, t.lms_level_id
		FROM soal s
		JOIN materi m ON s.id_materi = m.id
		JOIN mata_pelajaran mp ON m.id_mata_pelajaran = mp.id
		JOIN tingkat t ON m.id_tingkat = t.id
		WHERE s.id_materi = $1 AND s.is_active = true
		ORDER BY s.id`

	rows, err := r.db.Query(query, idMateri)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var soal entity.Soal
		var pembahasan *string
		err := rows.Scan(
			&soal.ID, &soal.IDMateri, &soal.IDTingkat, &soal.Pertanyaan, &soal.OpsiA, &soal.OpsiB, &soal.OpsiC, &soal.OpsiD, &soal.JawabanBenar, &pembahasan, &soal.IsActive,
			&soal.Materi.ID, &soal.Materi.IDMataPelajaran, &soal.Materi.IDTingkat, &soal.Materi.Nama, &soal.Materi.IsActive, &soal.Materi.DefaultDurasiMenit, &soal.Materi.DefaultJumlahSoal, &soal.Materi.LmsModuleID, &soal.Materi.LmsClassID,
			&soal.Materi.MataPelajaran.ID, &soal.Materi.MataPelajaran.Nama, &soal.Materi.MataPelajaran.IsActive, &soal.Materi.MataPelajaran.LmsSubjectID, &soal.Materi.MataPelajaran.LmsSchoolID, &soal.Materi.MataPelajaran.LmsClassID,
			&soal.Materi.Tingkat.ID, &soal.Materi.Tingkat.Nama, &soal.Materi.Tingkat.IsActive, &soal.Materi.Tingkat.LmsLevelID,
		)
		if err != nil {
			return nil, err
		}
		soal.Pembahasan = pembahasan

		// Get gambar for this soal
		gambarQuery := `
			SELECT id, id_soal, nama_file, file_path, file_size, mime_type, urutan, keterangan, cloud_id, public_id, created_at
			FROM soal_gambar
			WHERE id_soal = $1
			ORDER BY urutan ASC`
		gambarRows, err := r.db.Query(gambarQuery, soal.ID)
		if err != nil {
			return nil, err
		}

		for gambarRows.Next() {
			var gambar entity.SoalGambar
			var keterangan, cloudId, publicId *string
			err := gambarRows.Scan(&gambar.ID, &gambar.IDSoal, &gambar.NamaFile, &gambar.FilePath, &gambar.FileSize, &gambar.MimeType, &gambar.Urutan, &keterangan, &cloudId, &publicId, &gambar.CreatedAt)
			if err != nil {
				gambarRows.Close()
				return nil, err
			}
			gambar.Keterangan = keterangan
			gambar.CloudId = cloudId
			gambar.PublicId = publicId
			soal.Gambar = append(soal.Gambar, gambar)
		}
		gambarRows.Close()

		soals = append(soals, soal)
	}

	return soals, nil
}

// CreateGambar creates a new soal gambar
func (r *soalRepositoryImpl) CreateGambar(gambar *entity.SoalGambar) error {
	query := `
		INSERT INTO soal_gambar (id_soal, nama_file, file_path, file_size, mime_type, urutan, keterangan, cloud_id, public_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at`
	return r.db.QueryRow(query, gambar.IDSoal, gambar.NamaFile, gambar.FilePath, gambar.FileSize, gambar.MimeType, gambar.Urutan, gambar.Keterangan, gambar.CloudId, gambar.PublicId).Scan(&gambar.ID, &gambar.CreatedAt)
}

// GetGambarByID gets gambar by ID
func (r *soalRepositoryImpl) GetGambarByID(id int) (*entity.SoalGambar, error) {
	var gambar entity.SoalGambar
	var keterangan, cloudId, publicId *string
	query := `SELECT id, id_soal, nama_file, file_path, file_size, mime_type, urutan, keterangan, cloud_id, public_id, created_at FROM soal_gambar WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&gambar.ID, &gambar.IDSoal, &gambar.NamaFile, &gambar.FilePath, &gambar.FileSize, &gambar.MimeType, &gambar.Urutan, &keterangan, &cloudId, &publicId, &gambar.CreatedAt)
	if err != nil {
		return nil, err
	}
	gambar.Keterangan = keterangan
	gambar.CloudId = cloudId
	gambar.PublicId = publicId
	return &gambar, nil
}

// UpdateGambar updates gambar urutan and keterangan
func (r *soalRepositoryImpl) UpdateGambar(id int, urutan int, keterangan *string) error {
	query := `UPDATE soal_gambar SET urutan = $1, keterangan = $2 WHERE id = $3`
	_, err := r.db.Exec(query, urutan, keterangan, id)
	return err
}

// DeleteGambar deletes gambar by ID
func (r *soalRepositoryImpl) DeleteGambar(id int) error {
	query := `DELETE FROM soal_gambar WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// GetQuestionCountsByTopic returns the count of questions per topic (both MC and drag-drop)
func (r *soalRepositoryImpl) GetQuestionCountsByTopic() (map[int]int, error) {
	counts := make(map[int]int)

	// Count multiple-choice questions
	mcQuery := `
		SELECT id_materi, COUNT(*) as count
		FROM soal
		WHERE is_active = true
		GROUP BY id_materi`
	mcRows, err := r.db.Query(mcQuery)
	if err != nil {
		return nil, err
	}
	defer mcRows.Close()

	for mcRows.Next() {
		var idMateri, count int
		err := mcRows.Scan(&idMateri, &count)
		if err != nil {
			return nil, err
		}
		counts[idMateri] = count
	}

	// Count drag-drop questions
	ddQuery := `
		SELECT id_materi, COUNT(*) as count
		FROM soal_drag_drop
		WHERE is_active = true
		GROUP BY id_materi`
	ddRows, err := r.db.Query(ddQuery)
	if err != nil {
		return nil, err
	}
	defer ddRows.Close()

	for ddRows.Next() {
		var idMateri, count int
		err := ddRows.Scan(&idMateri, &count)
		if err != nil {
			return nil, err
		}
		counts[idMateri] += count // Add to existing count
	}

	return counts, nil
}