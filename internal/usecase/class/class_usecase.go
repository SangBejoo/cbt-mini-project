package class

import (
	"cbt-test-mini-project/internal/entity"
	classRepo "cbt-test-mini-project/internal/repository/class"
	"errors"
)

type classUsecaseImpl struct {
	repo classRepo.ClassRepository
}

func NewClassUsecase(repo classRepo.ClassRepository) ClassUsecase {
	return &classUsecaseImpl{repo: repo}
}

func (u *classUsecaseImpl) ListClasses(lmsSchoolID int64) ([]entity.Class, error) {
	if lmsSchoolID < 0 {
		return nil, errors.New("lms_school_id must be greater than or equal to 0")
	}

	classes, err := u.repo.List()
	if err != nil {
		return nil, err
	}

	if lmsSchoolID == 0 {
		return classes, nil
	}

	filtered := make([]entity.Class, 0, len(classes))
	for _, item := range classes {
		if item.LMSSchoolID == lmsSchoolID {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}
