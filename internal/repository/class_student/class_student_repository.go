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

func (r *classStudentRepository) resolveMembershipID(lmsClassID, lmsUserID int64) (int64, error) {
	query := `
		SELECT sm.id
		FROM classes c
		JOIN school_memberships sm ON sm.school_id = c.school_id
		WHERE c.id = $1
		  AND sm.user_id = $2
		  AND sm.deleted_at IS NULL
		  AND COALESCE(sm.status, 'active') = 'active'
		ORDER BY CASE sm.role::text
			WHEN 'student' THEN 1
			WHEN 'parent' THEN 2
			WHEN 'teacher' THEN 3
			WHEN 'school_admin' THEN 4
			ELSE 5
		END,
		sm.id
		LIMIT 1`

	var membershipID int64
	err := r.db.QueryRow(query, lmsClassID, lmsUserID).Scan(&membershipID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return membershipID, nil
}

// NewClassStudentRepository creates a new ClassStudentRepository instance
func NewClassStudentRepository(db *sql.DB) ClassStudentRepository {
	return &classStudentRepository{db: db}
}

// AddStudent adds a student to a class (upsert pattern for idempotency)
func (r *classStudentRepository) AddStudent(lmsClassID, lmsUserID int64) error {
	membershipID, err := r.resolveMembershipID(lmsClassID, lmsUserID)
	if err != nil {
		return err
	}
	if membershipID == 0 {
		return nil
	}

	reactivateQuery := `
		UPDATE class_students
		SET deleted_at = NULL,
		    status = 'active',
		    joined_at = COALESCE(joined_at, CURRENT_TIMESTAMP),
		    lms_class_id = $1,
		    lms_user_id = $2
		WHERE class_id = $1
		  AND student_membership_id = $3`
	res, err := r.db.Exec(reactivateQuery, lmsClassID, lmsUserID, membershipID)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected > 0 {
		return nil
	}

	insertQuery := `
		INSERT INTO class_students (class_id, student_membership_id, lms_class_id, lms_user_id, joined_at, status, join_method)
		VALUES ($1, $2, $1, $3, CURRENT_TIMESTAMP, 'active', 'manual')`
	_, err = r.db.Exec(insertQuery, lmsClassID, membershipID, lmsUserID)
	return err
}

// RemoveStudent removes a student from a class
func (r *classStudentRepository) RemoveStudent(lmsClassID, lmsUserID int64) error {
	membershipID, err := r.resolveMembershipID(lmsClassID, lmsUserID)
	if err != nil {
		return err
	}
	if membershipID == 0 {
		return nil
	}

	query := `
		UPDATE class_students
		SET deleted_at = NOW(), status = 'inactive'
		WHERE class_id = $1
		  AND student_membership_id = $2
		  AND deleted_at IS NULL`
	_, err = r.db.Exec(query, lmsClassID, membershipID)
	return err
}

// IsStudentInClass checks if a student is enrolled in a class
func (r *classStudentRepository) IsStudentInClass(lmsClassID, lmsUserID int64) (bool, error) {
	var exists bool
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM class_students cs
			JOIN school_memberships sm ON sm.id = cs.student_membership_id
			WHERE cs.class_id = $1
			  AND sm.user_id = $2
			  AND cs.deleted_at IS NULL
			  AND sm.deleted_at IS NULL
			  AND COALESCE(sm.status, 'active') = 'active'
		)`
	err := r.db.QueryRow(query, lmsClassID, lmsUserID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// GetStudentClasses returns all class IDs a student is enrolled in
func (r *classStudentRepository) GetStudentClasses(lmsUserID int64) ([]int64, error) {
	query := `
		SELECT cs.class_id
		FROM class_students cs
		JOIN school_memberships sm ON sm.id = cs.student_membership_id
		WHERE sm.user_id = $1
		  AND cs.deleted_at IS NULL
		  AND sm.deleted_at IS NULL
		  AND COALESCE(sm.status, 'active') = 'active'`
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
	query := `
		SELECT cs.id,
		       cs.class_id AS lms_class_id,
		       sm.user_id AS lms_user_id,
		       COALESCE(cs.joined_at, CURRENT_TIMESTAMP) AS joined_at
		FROM class_students cs
		JOIN school_memberships sm ON sm.id = cs.student_membership_id
		WHERE cs.class_id = $1
		  AND sm.user_id = $2
		  AND cs.deleted_at IS NULL
		  AND sm.deleted_at IS NULL
		  AND COALESCE(sm.status, 'active') = 'active'`
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
	query := `
		SELECT sm.user_id
		FROM class_students cs
		JOIN school_memberships sm ON sm.id = cs.student_membership_id
		WHERE cs.class_id = $1
		  AND cs.deleted_at IS NULL
		  AND sm.deleted_at IS NULL
		  AND COALESCE(sm.status, 'active') = 'active'`
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
	query := `
		SELECT cs.id,
		       cs.class_id AS lms_class_id,
		       sm.user_id AS lms_user_id,
		       COALESCE(cs.joined_at, CURRENT_TIMESTAMP) AS joined_at
		FROM class_students cs
		JOIN school_memberships sm ON sm.id = cs.student_membership_id
		WHERE cs.class_id = $1
		  AND cs.deleted_at IS NULL
		  AND sm.deleted_at IS NULL
		  AND COALESCE(sm.status, 'active') = 'active'
		ORDER BY COALESCE(cs.joined_at, CURRENT_TIMESTAMP) ASC`
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
