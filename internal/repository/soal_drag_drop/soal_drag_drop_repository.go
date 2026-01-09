package soal_drag_drop

import (
	"cbt-test-mini-project/internal/entity"
	"errors"

	"gorm.io/gorm"
)

// Create creates a new drag-drop question with items, slots, and correct answers
func (r *repository) Create(soal *entity.SoalDragDrop, items []entity.DragItem, slots []entity.DragSlot, correctAnswers []entity.DragCorrectAnswer) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create the main question
		if err := tx.Create(soal).Error; err != nil {
			return err
		}

		// Create items
		for i := range items {
			items[i].IDSoalDragDrop = soal.ID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}

		// Create slots
		for i := range slots {
			slots[i].IDSoalDragDrop = soal.ID
			if err := tx.Create(&slots[i]).Error; err != nil {
				return err
			}
		}

		// Create correct answers (need to map temp IDs to real IDs)
		// Assumes items and slots are passed in order with urutan matching
		itemMap := make(map[int]int)  // urutan -> id
		slotMap := make(map[int]int)  // urutan -> id
		for _, item := range items {
			itemMap[item.Urutan] = item.ID
		}
		for _, slot := range slots {
			slotMap[slot.Urutan] = slot.ID
		}

		for _, ca := range correctAnswers {
			// Map urutan values to actual database IDs
			itemID, itemExists := itemMap[ca.IDDragItem]
			slotID, slotExists := slotMap[ca.IDDragSlot]

			if !itemExists || !slotExists {
				return errors.New("invalid item or slot urutan in correct answers")
			}

			newCA := entity.DragCorrectAnswer{
				IDDragItem: itemID,
				IDDragSlot: slotID,
			}
			if err := tx.Create(&newCA).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetByID retrieves a drag-drop question by ID with items and slots
func (r *repository) GetByID(id int) (*entity.SoalDragDrop, error) {
	var soal entity.SoalDragDrop
	err := r.db.
		Preload("Materi.MataPelajaran").
		Preload("Materi.Tingkat").
		Preload("Tingkat").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("urutan ASC")
		}).
		Preload("Slots", func(db *gorm.DB) *gorm.DB {
			return db.Order("urutan ASC")
		}).
		First(&soal, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
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

	// Join through drag_item to get correct answers for this soal
	err := r.db.
		Joins("JOIN drag_item ON drag_correct_answer.id_drag_item = drag_item.id").
		Where("drag_item.id_soal_drag_drop = ?", soalID).
		Find(&correctAnswers).Error

	return correctAnswers, err
}

// Update updates a drag-drop question with items, slots, and correct answers
func (r *repository) Update(soal *entity.SoalDragDrop, items []entity.DragItem, slots []entity.DragSlot, correctAnswers []entity.DragCorrectAnswer) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Update main question
		if err := tx.Model(soal).Updates(map[string]interface{}{
			"id_materi":   soal.IDMateri,
			"id_tingkat":  soal.IDTingkat,
			"pertanyaan":  soal.Pertanyaan,
			"drag_type":   soal.DragType,
			"pembahasan":  soal.Pembahasan,
			"is_active":   soal.IsActive,
		}).Error; err != nil {
			return err
		}

		// Delete old items (will cascade delete correct_answers)
		if err := tx.Where("id_soal_drag_drop = ?", soal.ID).Delete(&entity.DragItem{}).Error; err != nil {
			return err
		}

		// Delete old slots
		if err := tx.Where("id_soal_drag_drop = ?", soal.ID).Delete(&entity.DragSlot{}).Error; err != nil {
			return err
		}

		// Create new items
		for i := range items {
			items[i].ID = 0 // Reset ID for new creation
			items[i].IDSoalDragDrop = soal.ID
			if err := tx.Create(&items[i]).Error; err != nil {
				return err
			}
		}

		// Create new slots
		for i := range slots {
			slots[i].ID = 0 // Reset ID for new creation
			slots[i].IDSoalDragDrop = soal.ID
			if err := tx.Create(&slots[i]).Error; err != nil {
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
			
			newCA := entity.DragCorrectAnswer{
				IDDragItem: itemID,
				IDDragSlot: slotID,
			}
			if err := tx.Create(&newCA).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// Delete soft-deletes a drag-drop question by setting is_active to false
func (r *repository) Delete(id int) error {
	return r.db.Model(&entity.SoalDragDrop{}).Where("id = ?", id).Update("is_active", false).Error
}

// List retrieves drag-drop questions with pagination and optional filters
func (r *repository) List(idMateri, idTingkat int, page, pageSize int) ([]entity.SoalDragDrop, int64, error) {
	var soals []entity.SoalDragDrop
	var total int64

	query := r.db.Model(&entity.SoalDragDrop{}).Where("is_active = ?", true)

	if idMateri > 0 {
		query = query.Where("id_materi = ?", idMateri)
	}
	if idTingkat > 0 {
		query = query.Where("id_tingkat = ?", idTingkat)
	}

	// Count total - simple query without preloads for speed
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results - load all necessary relations including nested ones
	offset := (page - 1) * pageSize
	err := query.
		Preload("Materi.MataPelajaran").
		Preload("Materi.Tingkat").
		Preload("Tingkat").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("urutan ASC")
		}).
		Preload("Slots", func(db *gorm.DB) *gorm.DB {
			return db.Order("urutan ASC")
		}).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&soals).Error

	return soals, total, err
}

// GetActiveByMateri retrieves active drag-drop questions for a materi (for test sessions)
func (r *repository) GetActiveByMateri(idMateri int, limit int) ([]entity.SoalDragDrop, error) {
	var soals []entity.SoalDragDrop

	query := r.db.
		Where("is_active = ? AND id_materi = ?", true, idMateri).
		Preload("Materi").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("urutan ASC")
		}).
		Preload("Slots", func(db *gorm.DB) *gorm.DB {
			return db.Order("urutan ASC")
		}).
		Order("RAND()")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&soals).Error
	return soals, err
}

// CountByMateri counts active drag-drop questions for a materi
func (r *repository) CountByMateri(idMateri int) (int64, error) {
	var count int64
	err := r.db.Model(&entity.SoalDragDrop{}).
		Where("is_active = ? AND id_materi = ?", true, idMateri).
		Count(&count).Error
	return count, err
}
