package soal_drag_drop

import (
	"cbt-test-mini-project/internal/entity"
	"database/sql"
	"errors"
)

// Create creates a new drag-drop question with items, slots, and correct answers
func (r *repository) Create(soal *entity.SoalDragDrop, items []entity.DragItem, slots []entity.DragSlot, correctAnswers []entity.DragCorrectAnswer) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create the main question
	soalQuery := `
		INSERT INTO soal_drag_drop (id_materi, id_tingkat, pertanyaan, drag_type, pembahasan, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id`
	err = tx.QueryRow(soalQuery, soal.IDMateri, soal.IDTingkat, soal.Pertanyaan, soal.DragType, soal.Pembahasan, soal.IsActive).Scan(&soal.ID)
	if err != nil {
		return err
	}

	// Create items
	for i := range items {
		items[i].IDSoalDragDrop = soal.ID
		itemQuery := `
			INSERT INTO drag_item (id_soal_drag_drop, label, image_url, urutan, created_at)
			VALUES ($1, $2, $3, $4, NOW())
			RETURNING id`
		err = tx.QueryRow(itemQuery, items[i].IDSoalDragDrop, items[i].Label, items[i].ImageURL, items[i].Urutan).Scan(&items[i].ID)
		if err != nil {
			return err
		}
	}

	// Create slots
	for i := range slots {
		slots[i].IDSoalDragDrop = soal.ID
		slotQuery := `
			INSERT INTO drag_slot (id_soal_drag_drop, label, image_url, urutan, created_at)
			VALUES ($1, $2, $3, $4, NOW())
			RETURNING id`
		err = tx.QueryRow(slotQuery, slots[i].IDSoalDragDrop, slots[i].Label, slots[i].ImageURL, slots[i].Urutan).Scan(&slots[i].ID)
		if err != nil {
			return err
		}
	}

	// Create correct answers (need to map temp IDs to real IDs)
	itemMap := make(map[int]int)  // urutan -> id
	slotMap := make(map[int]int)  // urutan -> id
	for _, item := range items {
		itemMap[item.Urutan] = item.ID
	}
	for _, slot := range slots {
		slotMap[slot.Urutan] = slot.ID
	}

	for _, ca := range correctAnswers {
		itemID, itemExists := itemMap[ca.IDDragItem]
		slotID, slotExists := slotMap[ca.IDDragSlot]

		if !itemExists || !slotExists {
			return errors.New("invalid item or slot urutan in correct answers")
		}

		correctAnswerQuery := `
			INSERT INTO drag_correct_answer (id_drag_item, id_drag_slot)
			VALUES ($1, $2)`
		_, err = tx.Exec(correctAnswerQuery, itemID, slotID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// GetByID retrieves a drag-drop question by ID with items and slots
func (r *repository) GetByID(id int) (*entity.SoalDragDrop, error) {
	// Get main soal with relations
	soalQuery := `
		SELECT sdd.id, sdd.id_materi, sdd.id_tingkat, sdd.pertanyaan, sdd.drag_type, sdd.pembahasan, sdd.is_active, sdd.created_at, sdd.updated_at,
		       m.id, m.id_mata_pelajaran, m.id_tingkat, m.nama, m.is_active, m.default_durasi_menit, m.default_jumlah_soal, m.lms_module_id, m.lms_class_id,
		       mp.id, mp.nama, mp.is_active, mp.lms_subject_id, mp.lms_school_id, mp.lms_class_id,
		       mt.id, mt.nama, mt.is_active, mt.lms_level_id,
		       t.id, t.nama, t.is_active, t.lms_level_id
		FROM soal_drag_drop sdd
		JOIN materi m ON sdd.id_materi = m.id
		JOIN mata_pelajaran mp ON m.id_mata_pelajaran = mp.id
		JOIN tingkat mt ON m.id_tingkat = mt.id
		JOIN tingkat t ON sdd.id_tingkat = t.id
		WHERE sdd.id = $1`

	var soal entity.SoalDragDrop
	var pembahasan *string
	err := r.db.QueryRow(soalQuery, id).Scan(
		&soal.ID, &soal.IDMateri, &soal.IDTingkat, &soal.Pertanyaan, &soal.DragType, &pembahasan, &soal.IsActive, &soal.CreatedAt, &soal.UpdatedAt,
		&soal.Materi.ID, &soal.Materi.IDMataPelajaran, &soal.Materi.IDTingkat, &soal.Materi.Nama, &soal.Materi.IsActive, &soal.Materi.DefaultDurasiMenit, &soal.Materi.DefaultJumlahSoal, &soal.Materi.LmsModuleID, &soal.Materi.LmsClassID,
		&soal.Materi.MataPelajaran.ID, &soal.Materi.MataPelajaran.Nama, &soal.Materi.MataPelajaran.IsActive, &soal.Materi.MataPelajaran.LmsSubjectID, &soal.Materi.MataPelajaran.LmsSchoolID, &soal.Materi.MataPelajaran.LmsClassID,
		&soal.Materi.Tingkat.ID, &soal.Materi.Tingkat.Nama, &soal.Materi.Tingkat.IsActive, &soal.Materi.Tingkat.LmsLevelID,
		&soal.Tingkat.ID, &soal.Tingkat.Nama, &soal.Tingkat.IsActive, &soal.Tingkat.LmsLevelID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	soal.Pembahasan = pembahasan

	// Get items
	itemsQuery := `
		SELECT id, id_soal_drag_drop, label, image_url, urutan, created_at
		FROM drag_item
		WHERE id_soal_drag_drop = $1
		ORDER BY urutan ASC`
	itemsRows, err := r.db.Query(itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer itemsRows.Close()

	for itemsRows.Next() {
		var item entity.DragItem
		err := itemsRows.Scan(&item.ID, &item.IDSoalDragDrop, &item.Label, &item.ImageURL, &item.Urutan, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		soal.Items = append(soal.Items, item)
	}

	// Get slots
	slotsQuery := `
		SELECT id, id_soal_drag_drop, label, image_url, urutan, created_at
		FROM drag_slot
		WHERE id_soal_drag_drop = $1
		ORDER BY urutan ASC`
	slotsRows, err := r.db.Query(slotsQuery, id)
	if err != nil {
		return nil, err
	}
	defer slotsRows.Close()

	for slotsRows.Next() {
		var slot entity.DragSlot
		err := slotsRows.Scan(&slot.ID, &slot.IDSoalDragDrop, &slot.Label, &slot.ImageURL, &slot.Urutan, &slot.CreatedAt)
		if err != nil {
			return nil, err
		}
		soal.Slots = append(soal.Slots, slot)
	}

	return &soal, nil
}

// GetByIDWithCorrectAnswers retrieves a drag-drop question with correct answers (admin only)
func (r *repository) GetByIDWithCorrectAnswers(id int) (*entity.SoalDragDrop, []entity.DragCorrectAnswer, error) {
	soal, err := r.GetByID(id)
	if err != nil || soal == nil {
		return soal, nil, err
	}

	correctAnswers, err := r.GetCorrectAnswersBySoalID(id)
	if err != nil {
		return nil, nil, err
	}

	return soal, correctAnswers, nil
}

// GetCorrectAnswersBySoalID gets correct answers for a question
func (r *repository) GetCorrectAnswersBySoalID(soalID int) ([]entity.DragCorrectAnswer, error) {
	var correctAnswers []entity.DragCorrectAnswer

	query := `
		SELECT dca.id_drag_item, dca.id_drag_slot
		FROM drag_correct_answer dca
		JOIN drag_item di ON dca.id_drag_item = di.id
		WHERE di.id_soal_drag_drop = $1`
	rows, err := r.db.Query(query, soalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ca entity.DragCorrectAnswer
		err := rows.Scan(&ca.IDDragItem, &ca.IDDragSlot)
		if err != nil {
			return nil, err
		}
		correctAnswers = append(correctAnswers, ca)
	}

	return correctAnswers, nil
}

// Update updates a drag-drop question with items, slots, and correct answers
func (r *repository) Update(soal *entity.SoalDragDrop, items []entity.DragItem, slots []entity.DragSlot, correctAnswers []entity.DragCorrectAnswer) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update main question
	updateQuery := `
		UPDATE soal_drag_drop
		SET id_materi = $1, id_tingkat = $2, pertanyaan = $3, drag_type = $4, pembahasan = $5, is_active = $6, updated_at = NOW()
		WHERE id = $7`
	_, err = tx.Exec(updateQuery, soal.IDMateri, soal.IDTingkat, soal.Pertanyaan, soal.DragType, soal.Pembahasan, soal.IsActive, soal.ID)
	if err != nil {
		return err
	}

	// Delete old items (cascade will delete correct_answers)
	deleteItemsQuery := `DELETE FROM drag_item WHERE id_soal_drag_drop = $1`
	_, err = tx.Exec(deleteItemsQuery, soal.ID)
	if err != nil {
		return err
	}

	// Delete old slots
	deleteSlotsQuery := `DELETE FROM drag_slot WHERE id_soal_drag_drop = $1`
	_, err = tx.Exec(deleteSlotsQuery, soal.ID)
	if err != nil {
		return err
	}

	// Create new items
	for i := range items {
		items[i].IDSoalDragDrop = soal.ID
		itemQuery := `
			INSERT INTO drag_item (id_soal_drag_drop, label, image_url, urutan, created_at)
			VALUES ($1, $2, $3, $4, NOW())
			RETURNING id`
		err = tx.QueryRow(itemQuery, items[i].IDSoalDragDrop, items[i].Label, items[i].ImageURL, items[i].Urutan).Scan(&items[i].ID)
		if err != nil {
			return err
		}
	}

	// Create new slots
	for i := range slots {
		slots[i].IDSoalDragDrop = soal.ID
		slotQuery := `
			INSERT INTO drag_slot (id_soal_drag_drop, label, image_url, urutan, created_at)
			VALUES ($1, $2, $3, $4, NOW())
			RETURNING id`
		err = tx.QueryRow(slotQuery, slots[i].IDSoalDragDrop, slots[i].Label, slots[i].ImageURL, slots[i].Urutan).Scan(&slots[i].ID)
		if err != nil {
			return err
		}
	}

	// Create maps for urutan to ID mapping
	itemMap := make(map[int]int)
	for _, item := range items {
		itemMap[item.Urutan] = item.ID
	}

	slotMap := make(map[int]int)
	for _, slot := range slots {
		slotMap[slot.Urutan] = slot.ID
	}

	// Create new correct answers using urutan mapping
	for _, ca := range correctAnswers {
		itemID, itemExists := itemMap[ca.IDDragItem]
		slotID, slotExists := slotMap[ca.IDDragSlot]

		if !itemExists || !slotExists {
			return errors.New("invalid item or slot urutan in correct answers")
		}

		correctAnswerQuery := `
			INSERT INTO drag_correct_answer (id_drag_item, id_drag_slot)
			VALUES ($1, $2)`
		_, err = tx.Exec(correctAnswerQuery, itemID, slotID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Delete soft-deletes a drag-drop question by setting is_active to false
func (r *repository) Delete(id int) error {
	query := `UPDATE soal_drag_drop SET is_active = false WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// List retrieves drag-drop questions with pagination and optional filters
func (r *repository) List(idMateri, idTingkat int, page, pageSize int) ([]entity.SoalDragDrop, int64, error) {
	var soals []entity.SoalDragDrop

	// Build WHERE clause
	whereClause := "WHERE sdd.is_active = true"
	args := []interface{}{}
	argCount := 0

	if idMateri > 0 {
		argCount++
		whereClause += " AND sdd.id_materi = $" + string(rune(argCount+'0'))
		args = append(args, idMateri)
	}
	if idTingkat > 0 {
		argCount++
		whereClause += " AND sdd.id_tingkat = $" + string(rune(argCount+'0'))
		args = append(args, idTingkat)
	}

	// Count total
	countQuery := `
		SELECT COUNT(*)
		FROM soal_drag_drop sdd
		` + whereClause

	var total int64
	err := r.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	listQuery := `
		SELECT sdd.id, sdd.id_materi, sdd.id_tingkat, sdd.pertanyaan, sdd.drag_type, sdd.pembahasan, sdd.is_active, sdd.created_at, sdd.updated_at,
		       m.id, m.id_mata_pelajaran, m.id_tingkat, m.nama, m.is_active, m.default_durasi_menit, m.default_jumlah_soal, m.lms_module_id, m.lms_class_id,
		       mp.id, mp.nama, mp.is_active, mp.lms_subject_id, mp.lms_school_id, mp.lms_class_id,
		       mt.id, mt.nama, mt.is_active, mt.lms_level_id,
		       t.id, t.nama, t.is_active, t.lms_level_id
		FROM soal_drag_drop sdd
		JOIN materi m ON sdd.id_materi = m.id
		JOIN mata_pelajaran mp ON m.id_mata_pelajaran = mp.id
		JOIN tingkat mt ON m.id_tingkat = mt.id
		JOIN tingkat t ON sdd.id_tingkat = t.id
		` + whereClause + `
		ORDER BY sdd.created_at DESC
		LIMIT $` + string(rune(argCount+1+'0')) + ` OFFSET $` + string(rune(argCount+2+'0'))

	args = append(args, pageSize, offset)
	rows, err := r.db.Query(listQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var soal entity.SoalDragDrop
		var pembahasan *string
		err := rows.Scan(
			&soal.ID, &soal.IDMateri, &soal.IDTingkat, &soal.Pertanyaan, &soal.DragType, &pembahasan, &soal.IsActive, &soal.CreatedAt, &soal.UpdatedAt,
			&soal.Materi.ID, &soal.Materi.IDMataPelajaran, &soal.Materi.IDTingkat, &soal.Materi.Nama, &soal.Materi.IsActive, &soal.Materi.DefaultDurasiMenit, &soal.Materi.DefaultJumlahSoal, &soal.Materi.LmsModuleID, &soal.Materi.LmsClassID,
			&soal.Materi.MataPelajaran.ID, &soal.Materi.MataPelajaran.Nama, &soal.Materi.MataPelajaran.IsActive, &soal.Materi.MataPelajaran.LmsSubjectID, &soal.Materi.MataPelajaran.LmsSchoolID, &soal.Materi.MataPelajaran.LmsClassID,
			&soal.Materi.Tingkat.ID, &soal.Materi.Tingkat.Nama, &soal.Materi.Tingkat.IsActive, &soal.Materi.Tingkat.LmsLevelID,
			&soal.Tingkat.ID, &soal.Tingkat.Nama, &soal.Tingkat.IsActive, &soal.Tingkat.LmsLevelID,
		)
		if err != nil {
			return nil, 0, err
		}
		soal.Pembahasan = pembahasan

		// Get items for this soal
		itemsQuery := `
			SELECT id, id_soal_drag_drop, label, image_url, urutan, created_at
			FROM drag_item
			WHERE id_soal_drag_drop = $1
			ORDER BY urutan ASC`
		itemsRows, err := r.db.Query(itemsQuery, soal.ID)
		if err != nil {
			return nil, 0, err
		}

		for itemsRows.Next() {
			var item entity.DragItem
			err := itemsRows.Scan(&item.ID, &item.IDSoalDragDrop, &item.Label, &item.ImageURL, &item.Urutan, &item.CreatedAt)
			if err != nil {
				itemsRows.Close()
				return nil, 0, err
			}
			soal.Items = append(soal.Items, item)
		}
		itemsRows.Close()

		// Get slots for this soal
		slotsQuery := `
			SELECT id, id_soal_drag_drop, label, image_url, urutan, created_at
			FROM drag_slot
			WHERE id_soal_drag_drop = $1
			ORDER BY urutan ASC`
		slotsRows, err := r.db.Query(slotsQuery, soal.ID)
		if err != nil {
			return nil, 0, err
		}

		for slotsRows.Next() {
			var slot entity.DragSlot
			err := slotsRows.Scan(&slot.ID, &slot.IDSoalDragDrop, &slot.Label, &slot.ImageURL, &slot.Urutan, &slot.CreatedAt)
			if err != nil {
				slotsRows.Close()
				return nil, 0, err
			}
			soal.Slots = append(soal.Slots, slot)
		}
		slotsRows.Close()

		soals = append(soals, soal)
	}

	return soals, total, nil
}

// GetActiveByMateri retrieves active drag-drop questions for a materi (for test sessions)
func (r *repository) GetActiveByMateri(idMateri int, limit int) ([]entity.SoalDragDrop, error) {
	var soals []entity.SoalDragDrop

	query := `
		SELECT sdd.id, sdd.id_materi, sdd.id_tingkat, sdd.pertanyaan, sdd.drag_type, sdd.pembahasan, sdd.is_active, sdd.created_at, sdd.updated_at,
		       m.id, m.id_mata_pelajaran, m.id_tingkat, m.nama, m.is_active, m.default_durasi_menit, m.default_jumlah_soal, m.lms_module_id, m.lms_class_id
		FROM soal_drag_drop sdd
		JOIN materi m ON sdd.id_materi = m.id
		WHERE sdd.is_active = true AND sdd.id_materi = $1
		ORDER BY RANDOM()`

	if limit > 0 {
		query += ` LIMIT $2`
		rows, err := r.db.Query(query, idMateri, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var soal entity.SoalDragDrop
			var pembahasan *string
			err := rows.Scan(
				&soal.ID, &soal.IDMateri, &soal.IDTingkat, &soal.Pertanyaan, &soal.DragType, &pembahasan, &soal.IsActive, &soal.CreatedAt, &soal.UpdatedAt,
				&soal.Materi.ID, &soal.Materi.IDMataPelajaran, &soal.Materi.IDTingkat, &soal.Materi.Nama, &soal.Materi.IsActive, &soal.Materi.DefaultDurasiMenit, &soal.Materi.DefaultJumlahSoal, &soal.Materi.LmsModuleID, &soal.Materi.LmsClassID,
			)
			if err != nil {
				return nil, err
			}
			soal.Pembahasan = pembahasan

			// Get items and slots for this soal
			r.loadItemsAndSlotsForSoal(&soal)
			soals = append(soals, soal)
		}
	} else {
		rows, err := r.db.Query(query, idMateri)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var soal entity.SoalDragDrop
			var pembahasan *string
			err := rows.Scan(
				&soal.ID, &soal.IDMateri, &soal.IDTingkat, &soal.Pertanyaan, &soal.DragType, &pembahasan, &soal.IsActive, &soal.CreatedAt, &soal.UpdatedAt,
				&soal.Materi.ID, &soal.Materi.IDMataPelajaran, &soal.Materi.IDTingkat, &soal.Materi.Nama, &soal.Materi.IsActive, &soal.Materi.DefaultDurasiMenit, &soal.Materi.DefaultJumlahSoal, &soal.Materi.LmsModuleID, &soal.Materi.LmsClassID,
			)
			if err != nil {
				return nil, err
			}
			soal.Pembahasan = pembahasan

			// Get items and slots for this soal
			r.loadItemsAndSlotsForSoal(&soal)
			soals = append(soals, soal)
		}
	}

	return soals, nil
}

// loadItemsAndSlotsForSoal loads items and slots for a soal (helper method)
func (r *repository) loadItemsAndSlotsForSoal(soal *entity.SoalDragDrop) error {
	// Get items
	itemsQuery := `
		SELECT id, id_soal_drag_drop, label, image_url, urutan, created_at
		FROM drag_item
		WHERE id_soal_drag_drop = $1
		ORDER BY urutan ASC`
	itemsRows, err := r.db.Query(itemsQuery, soal.ID)
	if err != nil {
		return err
	}
	defer itemsRows.Close()

	for itemsRows.Next() {
		var item entity.DragItem
		err := itemsRows.Scan(&item.ID, &item.IDSoalDragDrop, &item.Label, &item.ImageURL, &item.Urutan, &item.CreatedAt)
		if err != nil {
			return err
		}
		soal.Items = append(soal.Items, item)
	}

	// Get slots
	slotsQuery := `
		SELECT id, id_soal_drag_drop, label, image_url, urutan, created_at
		FROM drag_slot
		WHERE id_soal_drag_drop = $1
		ORDER BY urutan ASC`
	slotsRows, err := r.db.Query(slotsQuery, soal.ID)
	if err != nil {
		return err
	}
	defer slotsRows.Close()

	for slotsRows.Next() {
		var slot entity.DragSlot
		err := slotsRows.Scan(&slot.ID, &slot.IDSoalDragDrop, &slot.Label, &slot.ImageURL, &slot.Urutan, &slot.CreatedAt)
		if err != nil {
			return err
		}
		soal.Slots = append(soal.Slots, slot)
	}

	return nil
}

// CountByMateri counts active drag-drop questions for a materi
func (r *repository) CountByMateri(idMateri int) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM soal_drag_drop WHERE is_active = true AND id_materi = $1`
	err := r.db.QueryRow(query, idMateri).Scan(&count)
	return count, err
}
