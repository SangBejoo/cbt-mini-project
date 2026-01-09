package soal_drag_drop

import (
	"bytes"
	"cbt-test-mini-project/internal/entity"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Helper to upload image string (Base64) to Cloudinary
func (u *usecase) uploadImageToCloudinary(imageStr string) (string, error) {
	// If it's already a URL (not base64), return as is
	if !strings.HasPrefix(imageStr, "data:image") {
		return imageStr, nil
	}

	// Format: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAA..."
	parts := strings.Split(imageStr, ",")
	if len(parts) != 2 {
		return "", errors.New("invalid base64 image format")
	}

	// Decode base64
	imageBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 image: %v", err)
	}

	// Validate image type
	mimeType := http.DetectContentType(imageBytes)
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/webp" {
		return "", fmt.Errorf("invalid image type: %s, only JPG, PNG and WEBP are allowed", mimeType)
	}

	// Initialize Cloudinary
	if u.config == nil {
		return "", errors.New("configuration is nil")
	}
	cld, err := cloudinary.NewFromParams(u.config.Cloudinary.Name, u.config.Cloudinary.Key, u.config.Cloudinary.Secret)
	if err != nil {
		return "", fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}

	// Upload to Cloudinary
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	uploadParams := uploader.UploadParams{
		Folder: "cbt/drag_drop_items",
	}

	resp, err := cld.Upload.Upload(ctx, bytes.NewReader(imageBytes), uploadParams)
	if err != nil {
		return "", fmt.Errorf("failed to upload image to Cloudinary: %v", err)
	}

	return resp.SecureURL, nil
}

// Create creates a new drag-drop question
func (u *usecase) Create(req *CreateRequest) (*entity.SoalDragDrop, error) {
	// Validate request
	if req.Pertanyaan == "" {
		return nil, errors.New("pertanyaan is required")
	}
	if len(req.Items) < 2 {
		return nil, errors.New("at least 2 items are required")
	}

	// Different validation for ORDERING vs MATCHING
	if req.DragType == entity.DragTypeMatching {
		// MATCHING requires manually defined slots and correct answers
		if len(req.Slots) < 2 {
			return nil, errors.New("at least 2 slots are required for matching type")
		}
		if len(req.CorrectAnswers) != len(req.Items) {
			return nil, errors.New("each item must have exactly one correct slot for matching type")
		}
	}
	// ORDERING: slots and correct answers will be auto-generated

	// Create main question entity
	soal := &entity.SoalDragDrop{
		IDMateri:   req.IDMateri,
		IDTingkat:  req.IDTingkat,
		Pertanyaan: req.Pertanyaan,
		DragType:   req.DragType,
		Pembahasan: req.Pembahasan,
		IsActive:   true,
	}

	// Create item entities
	items := make([]entity.DragItem, len(req.Items))
	for i, item := range req.Items {
		var finalImageURL *string
		if item.ImageURL != nil && *item.ImageURL != "" {
			uploadedURL, err := u.uploadImageToCloudinary(*item.ImageURL)
			if err != nil {
				return nil, fmt.Errorf("failed to upload image for item %d: %v", i+1, err)
			}
			finalImageURL = &uploadedURL
		}

		items[i] = entity.DragItem{
			Label:    item.Label,
			ImageURL: finalImageURL,
			Urutan:   item.Urutan,
		}
	}

	var slots []entity.DragSlot
	var correctAnswers []entity.DragCorrectAnswer

	if req.DragType == entity.DragTypeOrdering {
		// ORDERING: Auto-generate numbered slots (1, 2, 3...) based on items count
		// Each slot represents a position where students drop items
		slots = make([]entity.DragSlot, len(req.Items))
		for i := range req.Items {
			slots[i] = entity.DragSlot{
				Label:  fmt.Sprintf("Posisi %d", i+1), // "Position 1", "Position 2", etc.
				Urutan: i + 1,
			}
		}

		// Auto-generate correct answers: item with urutan N belongs to slot with urutan N
		// This means the correct order is the order defined by the admin
		correctAnswers = make([]entity.DragCorrectAnswer, len(req.Items))
		for i, item := range req.Items {
			correctAnswers[i] = entity.DragCorrectAnswer{
				IDDragItem: item.Urutan, // Pass urutan, repository will map to ID
				IDDragSlot: item.Urutan, // Same urutan = correct position
			}
		}
	} else {
		// MATCHING: Use provided slots and correct answers
		slots = make([]entity.DragSlot, len(req.Slots))
		for i, slot := range req.Slots {
			var finalImageURL *string
			if slot.ImageURL != nil && *slot.ImageURL != "" {
				uploadedURL, err := u.uploadImageToCloudinary(*slot.ImageURL)
				if err != nil {
					return nil, fmt.Errorf("failed to upload image for slot %d: %v", i+1, err)
				}
				finalImageURL = &uploadedURL
			}

			slots[i] = entity.DragSlot{
				Label:    slot.Label,
				ImageURL: finalImageURL,
				Urutan:   slot.Urutan,
			}
		}

		// Build correct answers with urutan values (repository will map to IDs)
		correctAnswers = make([]entity.DragCorrectAnswer, len(req.CorrectAnswers))
		for i, ca := range req.CorrectAnswers {
			correctAnswers[i] = entity.DragCorrectAnswer{
				IDDragItem: ca.ItemUrutan, // Pass urutan, repository will map to ID
				IDDragSlot: ca.SlotUrutan, // Pass urutan, repository will map to ID
			}
		}
	}

	// Create all in one transaction
	if err := u.repo.Create(soal, items, slots, correctAnswers); err != nil {
		return nil, err
	}

	return u.repo.GetByID(soal.ID)
}

// GetByID retrieves a drag-drop question by ID
func (u *usecase) GetByID(id int) (*entity.SoalDragDrop, error) {
	return u.repo.GetByID(id)
}

// GetByIDWithCorrectAnswers retrieves with correct answers (admin only)
func (u *usecase) GetByIDWithCorrectAnswers(id int) (*entity.SoalDragDrop, []entity.DragCorrectAnswer, error) {
	return u.repo.GetByIDWithCorrectAnswers(id)
}

// Update updates a drag-drop question
func (u *usecase) Update(id int, req *UpdateRequest) (*entity.SoalDragDrop, error) {
	// Validate request
	if req.Pertanyaan == "" {
		return nil, errors.New("pertanyaan is required")
	}
	if len(req.Items) < 2 {
		return nil, errors.New("at least 2 items are required")
	}

	// Different validation for ORDERING vs MATCHING
	if req.DragType == entity.DragTypeMatching {
		// MATCHING requires manually defined slots and correct answers
		if len(req.Slots) < 2 {
			return nil, errors.New("at least 2 slots are required for matching type")
		}
	}

	// Get existing question
	existing, err := u.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, errors.New("question not found")
	}

	// Update main question
	existing.IDMateri = req.IDMateri
	existing.IDTingkat = req.IDTingkat
	existing.Pertanyaan = req.Pertanyaan
	existing.DragType = req.DragType
	existing.Pembahasan = req.Pembahasan
	existing.IsActive = req.IsActive

	// Create new items
	items := make([]entity.DragItem, len(req.Items))
	for i, item := range req.Items {
		var finalImageURL *string
		if item.ImageURL != nil && *item.ImageURL != "" {
			uploadedURL, err := u.uploadImageToCloudinary(*item.ImageURL)
			if err != nil {
				return nil, fmt.Errorf("failed to upload image for item %d: %v", i+1, err)
			}
			finalImageURL = &uploadedURL
		}

		items[i] = entity.DragItem{
			Label:    item.Label,
			ImageURL: finalImageURL,
			Urutan:   item.Urutan,
		}
	}

	var slots []entity.DragSlot
	var correctAnswers []entity.DragCorrectAnswer

	if req.DragType == entity.DragTypeOrdering {
		// ORDERING: Auto-generate numbered slots (1, 2, 3...) based on items count
		slots = make([]entity.DragSlot, len(req.Items))
		for i := range req.Items {
			slots[i] = entity.DragSlot{
				Label:  fmt.Sprintf("Posisi %d", i+1),
				Urutan: i + 1,
			}
		}

		// Auto-generate correct answers: item with urutan N belongs to slot with urutan N
		correctAnswers = make([]entity.DragCorrectAnswer, len(req.Items))
		for i, item := range req.Items {
			correctAnswers[i] = entity.DragCorrectAnswer{
				IDDragItem: item.Urutan,
				IDDragSlot: item.Urutan,
			}
		}
	} else {
		// MATCHING: Use provided slots and correct answers
		slots = make([]entity.DragSlot, len(req.Slots))
		for i, slot := range req.Slots {
			var finalImageURL *string
			if slot.ImageURL != nil && *slot.ImageURL != "" {
				uploadedURL, err := u.uploadImageToCloudinary(*slot.ImageURL)
				if err != nil {
					return nil, fmt.Errorf("failed to upload image for slot %d: %v", i+1, err)
				}
				finalImageURL = &uploadedURL
			}

			slots[i] = entity.DragSlot{
				Label:    slot.Label,
				ImageURL: finalImageURL,
				Urutan:   slot.Urutan,
			}
		}

		// Build correct answers with urutan values (repository will map to IDs)
		correctAnswers = make([]entity.DragCorrectAnswer, len(req.CorrectAnswers))
		for i, ca := range req.CorrectAnswers {
			correctAnswers[i] = entity.DragCorrectAnswer{
				IDDragItem: ca.ItemUrutan,
				IDDragSlot: ca.SlotUrutan,
			}
		}
	}

	// Update all in one transaction
	if err := u.repo.Update(existing, items, slots, correctAnswers); err != nil {
		return nil, err
	}

	return u.repo.GetByID(id)
}

// Delete deletes a drag-drop question
func (u *usecase) Delete(id int) error {
	existing, err := u.repo.GetByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("question not found")
	}
	return u.repo.Delete(id)
}

// List retrieves drag-drop questions with pagination
func (u *usecase) List(idMateri, idTingkat int, page, pageSize int) ([]entity.SoalDragDrop, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return u.repo.List(idMateri, idTingkat, page, pageSize)
}

// GetActiveByMateri retrieves active questions for test sessions
func (u *usecase) GetActiveByMateri(idMateri int, limit int) ([]entity.SoalDragDrop, error) {
	return u.repo.GetActiveByMateri(idMateri, limit)
}

// CheckDragDropAnswer checks if user's answer is correct (all-or-nothing)
func (u *usecase) CheckDragDropAnswer(soalID int, userAnswer map[int]int) (bool, error) {
	correctAnswers, err := u.repo.GetCorrectAnswersBySoalID(soalID)
	if err != nil {
		return false, err
	}

	if len(userAnswer) != len(correctAnswers) {
		return false, nil // Not all items were answered
	}

	// Check each correct answer
	for _, correct := range correctAnswers {
		userSlot, exists := userAnswer[correct.IDDragItem]
		if !exists || userSlot != correct.IDDragSlot {
			return false, nil // Wrong or missing answer
		}
	}

	return true, nil // All correct!
}

// CountByMateri counts questions for a materi
func (u *usecase) CountByMateri(idMateri int) (int64, error) {
	return u.repo.CountByMateri(idMateri)
}
