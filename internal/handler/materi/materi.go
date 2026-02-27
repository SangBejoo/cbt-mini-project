package materi

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/usecase/mata_pelajaran"
	"cbt-test-mini-project/internal/usecase/materi"
	"cbt-test-mini-project/internal/usecase/soal"
	"cbt-test-mini-project/util/interceptor"
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// materiHandler implements base.MateriServiceServer
type materiHandler struct {
	base.UnimplementedMateriServiceServer
	usecase     materi.MateriUsecase
	soalUsecase soal.SoalUsecase
	mataUsecase mata_pelajaran.MataPelajaranUsecase
}

// NewMateriHandler creates a new MateriHandler
func NewMateriHandler(usecase materi.MateriUsecase, soalUsecase soal.SoalUsecase, mataUsecase mata_pelajaran.MataPelajaranUsecase) base.MateriServiceServer {
	return &materiHandler{usecase: usecase, soalUsecase: soalUsecase, mataUsecase: mataUsecase}
}

func validateOwnershipSource(roleName string, lmsBookID, lmsTeacherMaterialID int64) error {
	switch roleName {
	case "superadmin":
		if lmsBookID <= 0 {
			return status.Error(codes.InvalidArgument, "superadmin wajib menggunakan lms_book_id")
		}
		if lmsTeacherMaterialID > 0 {
			return status.Error(codes.InvalidArgument, "superadmin tidak boleh menggunakan lms_teacher_material_id")
		}
	case "teacher":
		if lmsTeacherMaterialID <= 0 {
			return status.Error(codes.InvalidArgument, "guru wajib menggunakan lms_teacher_material_id")
		}
		if lmsBookID > 0 {
			return status.Error(codes.InvalidArgument, "guru tidak boleh menggunakan lms_book_id")
		}
	default:
		return status.Error(codes.PermissionDenied, "hanya superadmin atau guru yang boleh membuat materi CBT")
	}
	return nil
}

// Helper function to convert entity.Materi to proto.Materi
func (h *materiHandler) convertToProtoMateri(m *entity.Materi, questionCount int) *base.Materi {
	var labels []string
	if m.Labels != nil {
		labels = m.Labels
	}
	var owner int64
	if m.OwnerUserID != nil {
		owner = int64(*m.OwnerUserID)
	}
	var school int64
	if m.SchoolID != nil {
		school = *m.SchoolID
	}
	var lmsModuleID int64
	if m.LmsModuleID != nil {
		lmsModuleID = *m.LmsModuleID
	}
	var lmsBookID int64
	if m.LmsBookID != nil {
		lmsBookID = *m.LmsBookID
	}
	var lmsTeacherMaterialID int64
	if m.LmsTeacherMaterialID != nil {
		lmsTeacherMaterialID = *m.LmsTeacherMaterialID
	}
	return &base.Materi{
		Id:                   int32(m.ID),
		MataPelajaran:        &base.MataPelajaran{Id: int32(m.MataPelajaran.ID), Nama: m.MataPelajaran.Nama},
		Tingkat:              &base.Tingkat{Id: int32(m.Tingkat.ID), Nama: m.Tingkat.Nama},
		Nama:                 m.Nama,
		IsActive:             m.IsActive,
		DefaultDurasiMenit:   int32(m.DefaultDurasiMenit),
		DefaultJumlahSoal:    int32(m.DefaultJumlahSoal),
		JumlahSoalReal:       int32(questionCount),
		OwnerUserId:          owner,
		SchoolId:             school,
		Labels:               labels,
		LmsModuleId:          lmsModuleID,
		LmsBookId:            lmsBookID,
		LmsTeacherMaterialId: lmsTeacherMaterialID,
		RandomizeQuestions:   m.RandomizeQuestions,
	}
}

// CreateMateri creates a new materi
func (h *materiHandler) CreateMateri(ctx context.Context, req *base.CreateMateriRequest) (*base.MateriResponse, error) {
	// Get user from context
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	ownerUserID := int(user.Id)
	roleName := interceptor.GetRoleNameFromContext(ctx)
	if err := validateOwnershipSource(roleName, req.LmsBookId, req.LmsTeacherMaterialId); err != nil {
		return nil, err
	}

	// Resolve school_id from mata pelajaran if available
	var schoolID int64
	if req.IdMataPelajaran != 0 {
		mp, err := h.mataUsecase.GetMataPelajaran(int(req.IdMataPelajaran))
		if err == nil && mp != nil && mp.LmsSchoolID != nil {
			schoolID = *mp.LmsSchoolID
		}
	}

	// Collect labels from request
	var labels []string
	if len(req.Labels) > 0 {
		labels = req.Labels
	}

	var lmsModuleID *int64
	if req.LmsModuleId > 0 {
		v := req.LmsModuleId
		lmsModuleID = &v
	}
	var lmsBookID *int64
	if req.LmsBookId > 0 {
		v := req.LmsBookId
		lmsBookID = &v
	}
	var lmsTeacherMaterialID *int64
	if req.LmsTeacherMaterialId > 0 {
		v := req.LmsTeacherMaterialId
		lmsTeacherMaterialID = &v
	}

	m, err := h.usecase.CreateMateri(int(req.IdMataPelajaran), req.Nama, int(req.IdTingkat), req.IsActive, int(req.DefaultDurasiMenit), int(req.DefaultJumlahSoal), ownerUserID, schoolID, labels, req.RandomizeQuestions, lmsModuleID, lmsBookID, lmsTeacherMaterialID)
	if err != nil {
		return nil, err
	}

	return &base.MateriResponse{
		Materi: h.convertToProtoMateri(m, 0),
	}, nil
}

func (h *materiHandler) CreateMateriSuperadmin(ctx context.Context, req *base.CreateMateriSuperadminRequest) (*base.MateriResponse, error) {
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if interceptor.GetRoleNameFromContext(ctx) != "superadmin" {
		return nil, status.Error(codes.PermissionDenied, "hanya superadmin yang boleh akses endpoint ini")
	}
	if req.LmsBookId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "lms_book_id wajib diisi")
	}

	var schoolID int64
	if req.IdMataPelajaran != 0 {
		mp, err := h.mataUsecase.GetMataPelajaran(int(req.IdMataPelajaran))
		if err == nil && mp != nil && mp.LmsSchoolID != nil {
			schoolID = *mp.LmsSchoolID
		}
	}

	lmsBookID := req.LmsBookId
	m, err := h.usecase.CreateMateri(
		int(req.IdMataPelajaran),
		req.Nama,
		int(req.IdTingkat),
		req.IsActive,
		int(req.DefaultDurasiMenit),
		int(req.DefaultJumlahSoal),
		int(user.Id),
		schoolID,
		req.Labels,
		req.RandomizeQuestions,
		nil,
		&lmsBookID,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &base.MateriResponse{Materi: h.convertToProtoMateri(m, 0)}, nil
}

func (h *materiHandler) CreateMateriTeacher(ctx context.Context, req *base.CreateMateriTeacherRequest) (*base.MateriResponse, error) {
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}
	if interceptor.GetRoleNameFromContext(ctx) != "teacher" {
		return nil, status.Error(codes.PermissionDenied, "hanya guru yang boleh akses endpoint ini")
	}
	if req.LmsTeacherMaterialId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "lms_teacher_material_id wajib diisi")
	}

	var schoolID int64
	if req.IdMataPelajaran != 0 {
		mp, err := h.mataUsecase.GetMataPelajaran(int(req.IdMataPelajaran))
		if err == nil && mp != nil && mp.LmsSchoolID != nil {
			schoolID = *mp.LmsSchoolID
		}
	}

	lmsTeacherMaterialID := req.LmsTeacherMaterialId
	m, err := h.usecase.CreateMateri(
		int(req.IdMataPelajaran),
		req.Nama,
		int(req.IdTingkat),
		req.IsActive,
		int(req.DefaultDurasiMenit),
		int(req.DefaultJumlahSoal),
		int(user.Id),
		schoolID,
		req.Labels,
		req.RandomizeQuestions,
		nil,
		nil,
		&lmsTeacherMaterialID,
	)
	if err != nil {
		return nil, err
	}

	return &base.MateriResponse{Materi: h.convertToProtoMateri(m, 0)}, nil
}

// GetMateri gets materi by ID
func (h *materiHandler) GetMateri(ctx context.Context, req *base.GetMateriRequest) (*base.MateriResponse, error) {
	m, err := h.usecase.GetMateri(int(req.Id))
	if err != nil {
		return nil, err
	}

	// Get question count for this materi
	counts, _ := h.soalUsecase.GetQuestionCountsByTopic()
	questionCount := 0
	if counts != nil {
		questionCount = counts[m.ID]
	}

	return &base.MateriResponse{
		Materi: h.convertToProtoMateri(m, questionCount),
	}, nil
}

// UpdateMateri updates materi
func (h *materiHandler) UpdateMateri(ctx context.Context, req *base.UpdateMateriRequest) (*base.MateriResponse, error) {
	roleName := interceptor.GetRoleNameFromContext(ctx)
	existing, err := h.usecase.GetMateri(int(req.Id))
	if err != nil {
		return nil, err
	}

	var effectiveBookID int64
	if existing.LmsBookID != nil {
		effectiveBookID = *existing.LmsBookID
	}
	if req.LmsBookId > 0 {
		effectiveBookID = req.LmsBookId
	}

	var effectiveTeacherMaterialID int64
	if existing.LmsTeacherMaterialID != nil {
		effectiveTeacherMaterialID = *existing.LmsTeacherMaterialID
	}
	if req.LmsTeacherMaterialId > 0 {
		effectiveTeacherMaterialID = req.LmsTeacherMaterialId
	}

	if err := validateOwnershipSource(roleName, effectiveBookID, effectiveTeacherMaterialID); err != nil {
		return nil, err
	}

	var lmsModuleID *int64
	if req.LmsModuleId > 0 {
		v := req.LmsModuleId
		lmsModuleID = &v
	}
	var lmsBookID *int64
	if req.LmsBookId > 0 {
		v := req.LmsBookId
		lmsBookID = &v
	}
	var lmsTeacherMaterialID *int64
	if req.LmsTeacherMaterialId > 0 {
		v := req.LmsTeacherMaterialId
		lmsTeacherMaterialID = &v
	}

	m, err := h.usecase.UpdateMateri(int(req.Id), int(req.IdMataPelajaran), req.Nama, int(req.IdTingkat), req.IsActive, int(req.DefaultDurasiMenit), int(req.DefaultJumlahSoal), req.RandomizeQuestions, lmsModuleID, lmsBookID, lmsTeacherMaterialID)
	if err != nil {
		return nil, err
	}

	// Get question count for this materi
	counts, _ := h.soalUsecase.GetQuestionCountsByTopic()
	questionCount := 0
	if counts != nil {
		questionCount = counts[m.ID]
	}

	return &base.MateriResponse{
		Materi: h.convertToProtoMateri(m, questionCount),
	}, nil
}

// DeleteMateri deletes materi
func (h *materiHandler) DeleteMateri(ctx context.Context, req *base.DeleteMateriRequest) (*base.MessageStatusResponse, error) {
	err := h.usecase.DeleteMateri(int(req.Id))
	if err != nil {
		return &base.MessageStatusResponse{
			Message: "Failed to delete materi",
			Status:  "error",
		}, err
	}

	return &base.MessageStatusResponse{
		Message: "Materi deleted successfully",
		Status:  "success",
	}, nil
}

// ListMateri lists materi
func (h *materiHandler) ListMateri(ctx context.Context, req *base.ListMateriRequest) (*base.ListMateriResponse, error) {
	page := 1
	pageSize := 100 // Changed from 10 to 100 for better default
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = int(req.Pagination.Page)
		}
		if req.Pagination.PageSize > 0 {
			pageSize = int(req.Pagination.PageSize)
		}
	}
	// Cap pageSize to prevent abuse
	if pageSize > 1000 {
		pageSize = 1000
	}
	materis, pagination, err := h.usecase.ListMateri(int(req.IdMataPelajaran), int(req.IdTingkat), page, pageSize)
	if err != nil {
		return nil, err
	}

	// Get all question counts in one query
	counts, _ := h.soalUsecase.GetQuestionCountsByTopic()
	if counts == nil {
		counts = make(map[int]int)
	}

	var materiList []*base.Materi
	for _, m := range materis {
		questionCount := counts[m.ID]
		materiList = append(materiList, h.convertToProtoMateri(&m, questionCount))
	}

	return &base.ListMateriResponse{
		Materi: materiList,
		Pagination: &base.PaginationResponse{
			TotalCount:  int32(pagination.TotalCount),
			TotalPages:  int32(pagination.TotalPages),
			CurrentPage: int32(pagination.CurrentPage),
			PageSize:    int32(pagination.PageSize),
		},
	}, nil
}
