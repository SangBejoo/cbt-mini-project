package class_student

import (
	"database/sql"

	"cbt-test-mini-project/internal/entity"
)

// ClassStudentRepository defines the interface for class student operations
type ClassStudentRepository interface {
	AddStudent(lmsClassID, lmsUserID int64) error
	RemoveStudent(lmsClassID, lmsUserID int64) error
	IsStudentInClass(lmsClassID, lmsUserID int64) (bool, error)
	GetStudentClasses(lmsUserID int64) ([]int64, error)
	GetByClassAndUser(lmsClassID, lmsUserID int64) (*entity.ClassStudent, error)
	GetStudentIDsByClassID(lmsClassID int64) ([]int64, error)
	ListByClassID(lmsClassID int64) ([]entity.ClassStudent, error)
}

type classStudentRepository struct {
	db *sql.DB
}

// NewClassStudentRepository creates a new ClassStudentRepository instance
func NewClassStudentRepository(db *sql.DB) ClassStudentRepository {
	return &classStudentRepository{db: db}
}

// AddStudent adds a student to a class (upsert pattern for idempotency)
func (r *classStudentRepository) AddStudent(lmsClassID, lmsUserID int64) error {
	query := `
		INSERT INTO class_students (lms_class_id, lms_user_id, joined_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (lms_class_id, lms_user_id) DO NOTHING
	`
	_, err := r.db.Exec(query, lmsClassID, lmsUserID)
	return err
}

// RemoveStudent removes a student from a class
func (r *classStudentRepository) RemoveStudent(lmsClassID, lmsUserID int64) error {
	query := `DELETE FROM class_students WHERE lms_class_id = $1 AND lms_user_id = $2`
	_, err := r.db.Exec(query, lmsClassID, lmsUserID)
	return err
}

// IsStudentInClass checks if a student is enrolled in a class
func (r *classStudentRepository) IsStudentInClass(lmsClassID, lmsUserID int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM class_students WHERE lms_class_id = $1 AND lms_user_id = $2)`
	err := r.db.QueryRow(query, lmsClassID, lmsUserID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetStudentClasses returns all class IDs a student is enrolled in
func (r *classStudentRepository) GetStudentClasses(lmsUserID int64) ([]int64, error) {
	query := `SELECT lms_class_id FROM class_students WHERE lms_user_id = $1`
	rows, err := r.db.Query(query, lmsUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classIDs []int64
	for rows.Next() {
		var classID int64
		if err := rows.Scan(&classID); err != nil {
			return nil, err
		}
		classIDs = append(classIDs, classID)
	}
	return classIDs, rows.Err()
}

// GetByClassAndUser retrieves a class student record by class and user
func (r *classStudentRepository) GetByClassAndUser(lmsClassID, lmsUserID int64) (*entity.ClassStudent, error) {
	var cs entity.ClassStudent
	query := `SELECT id, lms_class_id, lms_user_id, joined_at FROM class_students WHERE lms_class_id = $1 AND lms_user_id = $2`
	err := r.db.QueryRow(query, lmsClassID, lmsUserID).Scan(&cs.ID, &cs.LMSClassID, &cs.LMSUserID, &cs.JoinedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &cs, nil
}

// GetStudentIDsByClassID returns all student IDs enrolled in a class
func (r *classStudentRepository) GetStudentIDsByClassID(lmsClassID int64) ([]int64, error) {
	query := `SELECT lms_user_id FROM class_students WHERE lms_class_id = $1`
	rows, err := r.db.Query(query, lmsClassID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var studentIDs []int64
	for rows.Next() {
		var studentID int64
		if err := rows.Scan(&studentID); err != nil {
			return nil, err
		}
		studentIDs = append(studentIDs, studentID)
	}
	return studentIDs, rows.Err()
}

// ListByClassID returns detailed class-student rows for a class
func (r *classStudentRepository) ListByClassID(lmsClassID int64) ([]entity.ClassStudent, error) {
	query := `SELECT id, lms_class_id, lms_user_id, joined_at FROM class_students WHERE lms_class_id = $1 ORDER BY joined_at ASC`
	rows, err := r.db.Query(query, lmsClassID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	students := make([]entity.ClassStudent, 0)
	for rows.Next() {
		var item entity.ClassStudent
		if err := rows.Scan(&item.ID, &item.LMSClassID, &item.LMSUserID, &item.JoinedAt); err != nil {
			return nil, err
		}
		students = append(students, item)
	}

	return students, rows.Err()
}
