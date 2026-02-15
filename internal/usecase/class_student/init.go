package class_student

import "cbt-test-mini-project/internal/entity"

type ClassStudentUsecase interface {
	ListClassStudents(lmsClassID int64) ([]entity.ClassStudent, error)
}
