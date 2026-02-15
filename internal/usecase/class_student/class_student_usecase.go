package class_student

import (
	"cbt-test-mini-project/internal/entity"
	classStudentRepo "cbt-test-mini-project/internal/repository/class_student"
	"errors"
)

type classStudentUsecaseImpl struct {
	repo classStudentRepo.ClassStudentRepository
}

func NewClassStudentUsecase(repo classStudentRepo.ClassStudentRepository) ClassStudentUsecase {
	return &classStudentUsecaseImpl{repo: repo}
}

func (u *classStudentUsecaseImpl) ListClassStudents(lmsClassID int64) ([]entity.ClassStudent, error) {
	if lmsClassID <= 0 {
		return nil, errors.New("lms_class_id is required")
	}

	return u.repo.ListByClassID(lmsClassID)
}
