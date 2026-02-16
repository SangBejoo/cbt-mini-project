package contracts

// EventType is the canonical event name shared between LMS and CBT.
type EventType string

const (
	ExamResultCompleted EventType = "exam_result_completed"
	ExamAssignmentCreated EventType = "exam_assignment_created"
	ExamAssignmentUpdated EventType = "exam_assignment_updated"
	ExamAssignmentDeleted EventType = "exam_assignment_deleted"
	ModuleUpsert EventType = "module_upsert"
	ModuleDeleted EventType = "module_deleted"
	ClassUpsert EventType = "class_upsert"
	ClassDeleted EventType = "class_deleted"
	ClassStudentJoined EventType = "class_student_joined"
	ClassStudentLeft EventType = "class_student_left"
)

// ExamResultPayload is emitted by CBT and consumed by LMS.
type ExamResultPayload struct {
	SessionID       int     `json:"session_id"`
	LMSAssignmentID int64   `json:"lms_assignment_id"`
	LMSUserID       int64   `json:"lms_user_id"`
	LMSClassID      int64   `json:"lms_class_id"`
	Score           float64 `json:"score"`
	CorrectCount    int     `json:"correct_count"`
	TotalCount      int     `json:"total_count"`
	CompletedAt     string  `json:"completed_at"`
}

// ExamAssignmentPayload is emitted by LMS and consumed by CBT.
type ExamAssignmentPayload struct {
	LMSAssignmentID int64   `json:"lms_assignment_id"`
	LMSClassID      int64   `json:"lms_class_id"`
	Title           string  `json:"title"`
	MaxScore        float64 `json:"max_score"`
	ModuleID        int64   `json:"module_id"`
	ModuleRefType   string  `json:"module_ref_type,omitempty"`
	ScheduledTime   string  `json:"scheduled_time"`
}

// ModuleUpsertPayload is emitted by LMS and consumed by CBT.
type ModuleUpsertPayload struct {
	ID        int64  `json:"id"`
	ClassID   int64  `json:"class_id"`
	SubjectID int64  `json:"subject_id"`
	LevelID   int64  `json:"level_id"`
	Name      string `json:"name"`
}

// ClassPayload is emitted by LMS and consumed by CBT.
type ClassPayload struct {
	ID       int64  `json:"id"`
	SchoolID int64  `json:"school_id"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

// ClassStudentPayload is emitted by LMS and consumed by CBT.
type ClassStudentPayload struct {
	LMSClassID int64 `json:"lms_class_id"`
	LMSUserID  int64 `json:"lms_user_id"`
}

// DeletePayload is emitted by LMS for delete sync operations.
type DeletePayload struct {
	ID int64 `json:"id"`
}
