package class

import "cbt-test-mini-project/internal/entity"

type ClassUsecase interface {
	ListClasses(lmsSchoolID int64) ([]entity.Class, error)
}
