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
	return &base.Materi{
		Id:                    int32(m.ID),
		MataPelajaran:         &base.MataPelajaran{Id: int32(m.MataPelajaran.ID), Nama: m.MataPelajaran.Nama},
		Tingkat:               &base.Tingkat{Id: int32(m.Tingkat.ID), Nama: m.Tingkat.Nama},
		Nama:                  m.Nama,
		IsActive:              m.IsActive,
		DefaultDurasiMenit:    int32(m.DefaultDurasiMenit),
		DefaultJumlahSoal:     int32(m.DefaultJumlahSoal),
		JumlahSoalReal:        int32(questionCount),
		OwnerUserId:           owner,
		SchoolId:              school,
		Labels:                labels,
	}
}

// CreateMateri creates a new materi
func (h *materiHandler) CreateMateri(ctx context.Context, req *base.CreateMateriRequest) (*base.MateriResponse, error) {
	// Get user from context
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		// REST gateway fallback
		token, extractErr := interceptor.ExtractTokenFromContext(ctx)
		if extractErr != nil {
			return nil, status.Error(codes.Unauthenticated, "user not authenticated")
		}
		claims, validateErr := interceptor.ValidateToken(token)
		if validateErr != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		user = &base.User{Id: claims.UserID}
	}
	ownerUserID := int(user.Id)

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

	m, err := h.usecase.CreateMateri(int(req.IdMataPelajaran), req.Nama, int(req.IdTingkat), req.IsActive, int(req.DefaultDurasiMenit), int(req.DefaultJumlahSoal), ownerUserID, schoolID, labels)
	if err != nil {
		return nil, err
	}

	return &base.MateriResponse{
		Materi: h.convertToProtoMateri(m, 0),
	}, nil
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
	m, err := h.usecase.UpdateMateri(int(req.Id), int(req.IdMataPelajaran), req.Nama, int(req.IdTingkat), req.IsActive, int(req.DefaultDurasiMenit), int(req.DefaultJumlahSoal))
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
		Materi:     materiList,
		Pagination: &base.PaginationResponse{
			TotalCount:  int32(pagination.TotalCount),
			TotalPages:  int32(pagination.TotalPages),
			CurrentPage: int32(pagination.CurrentPage),
			PageSize:    int32(pagination.PageSize),
		},
	}, nil
}