package history

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	"cbt-test-mini-project/internal/usecase/history"
	"cbt-test-mini-project/util/interceptor"
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// historyHandler implements base.HistoryServiceServer
type historyHandler struct {
	base.UnimplementedHistoryServiceServer
	usecase history.HistoryUsecase
}

// NewHistoryHandler creates a new HistoryHandler
func NewHistoryHandler(usecase history.HistoryUsecase) base.HistoryServiceServer {
	return &historyHandler{usecase: usecase}
}

// GetStudentHistory gets student history
func (h *historyHandler) GetStudentHistory(ctx context.Context, req *base.StudentHistoryRequest) (*base.StudentHistoryResponse, error) {
	// Try to get user from context first (for gRPC calls)
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		// For REST gateway, extract token from metadata
		token, extractErr := interceptor.ExtractTokenFromContext(ctx)
		if extractErr != nil {
			return nil, status.Error(codes.Unauthenticated, "user not authenticated")
		}
		claims, validateErr := interceptor.ValidateToken(token)
		if validateErr != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		user = &base.User{
			Id:    claims.UserID,
			Email: claims.Email,
			Role:  base.UserRole(claims.Role),
		}
		// Add to context for consistency
		ctx = interceptor.AddUserToContext(ctx, claims)
	}

	userID := user.Id

	var tingkatan, idMataPelajaran *int
	if req.Tingkatan != 0 {
		t := int(req.Tingkatan)
		tingkatan = &t
	}
	if req.IdMataPelajaran != 0 {
		i := int(req.IdMataPelajaran)
		idMataPelajaran = &i
	}

	page := 1
	pageSize := 20
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = int(req.Pagination.Page)
		}
		if req.Pagination.PageSize > 0 {
			pageSize = int(req.Pagination.PageSize)
			// Cap at 100 for single student history (lighter query)
			if pageSize > 100 {
				pageSize = 100
			}
		}
	}

	response, err := h.usecase.GetStudentHistory(int(userID), tingkatan, idMataPelajaran, page, pageSize)
	if err != nil {
		return nil, err
	}

	var histories []*base.HistorySummary
	for _, h := range response.History {
		var waktuSelesai *timestamppb.Timestamp
		if h.WaktuSelesai != nil {
			waktuSelesai = timestamppb.New(*h.WaktuSelesai)
		}

		histories = append(histories, &base.HistorySummary{
			Id:                    int32(h.ID),
			SessionToken:          h.SessionToken,
			NamaPeserta:           h.NamaPeserta,
			MataPelajaran:         &base.MataPelajaran{Id: int32(h.MataPelajaran.ID), Nama: h.MataPelajaran.Nama},
			Tingkat:               &base.Tingkat{Id: int32(h.Tingkat.ID), Nama: h.Tingkat.Nama},
			WaktuMulai:            timestamppb.New(*h.WaktuMulai),
			WaktuSelesai:          waktuSelesai,
			DurasiPengerjaanDetik: int32(h.DurasiPengerjaanDetik),
			NilaiAkhir:            h.NilaiAkhir,
			JumlahBenar:           int32(h.JumlahBenar),
			TotalSoal:             int32(h.TotalSoal),
			Status:                base.TestStatus(base.TestStatus_value[strings.ToUpper(string(h.Status))]),
		})
	}

	return &base.StudentHistoryResponse{
		User:              h.convertUserToProto(response.User),
		Tingkatan:         req.Tingkatan,
		History:           histories,
		RataRataNilai:     response.RataRataNilai,
		TotalTestCompleted: int32(response.TotalTestCompleted),
		Pagination: &base.PaginationResponse{
			TotalCount:  int32(response.Pagination.TotalCount),
			TotalPages:  int32(response.Pagination.TotalPages),
			CurrentPage: int32(response.Pagination.CurrentPage),
			PageSize:    int32(response.Pagination.PageSize),
		},
	}, nil
}

// GetHistoryDetail gets detailed history
func (h *historyHandler) GetHistoryDetail(ctx context.Context, req *base.GetHistoryDetailRequest) (*base.HistoryDetailResponse, error) {
	// Get user from JWT context
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		// For REST gateway, extract token from metadata
		token, extractErr := interceptor.ExtractTokenFromContext(ctx)
		if extractErr != nil {
			return nil, status.Error(codes.Unauthenticated, "user not authenticated")
		}
		claims, validateErr := interceptor.ValidateToken(token)
		if validateErr != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		user = &base.User{
			Id:    claims.UserID,
			Email: claims.Email,
			Role:  base.UserRole(claims.Role),
		}
		// Add to context for consistency
		ctx = interceptor.AddUserToContext(ctx, claims)
	}

	response, err := h.usecase.GetHistoryDetail(req.SessionToken)
	if err != nil {
		return nil, err
	}

	// Check if the session belongs to the authenticated user
	if response.SessionInfo.UserID == nil || *response.SessionInfo.UserID != int(user.Id) {
		return nil, status.Error(codes.PermissionDenied, "you do not have permission to access this session history")
	}

	var jawabanDetails []*base.JawabanDetail
	for _, d := range response.DetailJawaban {
		var jawabanDipilih base.JawabanOption
		if d.JawabanDipilih != nil {
			jawabanDipilih = base.JawabanOption(base.JawabanOption_value[string(*d.JawabanDipilih)])
		}

		jawabanDetails = append(jawabanDetails, &base.JawabanDetail{
			NomorUrut:      int32(d.NomorUrut),
			Pertanyaan:     d.Pertanyaan,
			OpsiA:          d.OpsiA,
			OpsiB:          d.OpsiB,
			OpsiC:          d.OpsiC,
			OpsiD:          d.OpsiD,
			JawabanDipilih: jawabanDipilih,
			JawabanBenar:   base.JawabanOption(base.JawabanOption_value[string(d.JawabanBenar)]),
			IsCorrect:      d.IsCorrect,
		})
	}

	var breakdownMateri []*base.MateriBreakdown
	for _, b := range response.BreakdownMateri {
		breakdownMateri = append(breakdownMateri, &base.MateriBreakdown{
			NamaMateri:      b.NamaMateri,
			JumlahSoal:      int32(b.JumlahSoal),
			JumlahBenar:     int32(b.JumlahBenar),
			PersentaseBenar: b.PersentaseBenar,
		})
	}

	return &base.HistoryDetailResponse{
		SessionInfo:     h.convertToProtoTestSession(response.SessionInfo),
		DetailJawaban:   jawabanDetails,
		BreakdownMateri: breakdownMateri,
	}, nil
}

// Helper function to convert entity to proto
func (h *historyHandler) convertToProtoTestSession(session *entity.TestSession) *base.TestSession {
	var waktuSelesai, batasWaktu *timestamppb.Timestamp
	if session.WaktuSelesai != nil {
		waktuSelesai = timestamppb.New(*session.WaktuSelesai)
	}
	batasWaktu = timestamppb.New(session.BatasWaktu())

	var nilaiAkhir float64
	if session.NilaiAkhir != nil {
		nilaiAkhir = *session.NilaiAkhir
	}

	var jumlahBenar, totalSoal int32
	if session.JumlahBenar != nil {
		jumlahBenar = int32(*session.JumlahBenar)
	}
	if session.TotalSoal != nil {
		totalSoal = int32(*session.TotalSoal)
	}

	status := base.TestStatus(base.TestStatus_value[strings.ToUpper(string(session.Status))])

	return &base.TestSession{
		Id:              int32(session.ID),
		SessionToken:    session.SessionToken,
		User:            h.convertUserToProto(session.User),
		Tingkat:         &base.Tingkat{Id: int32(session.Tingkat.ID), Nama: session.Tingkat.Nama},
		MataPelajaran:   &base.MataPelajaran{Id: int32(session.MataPelajaran.ID), Nama: session.MataPelajaran.Nama},
		WaktuMulai:      timestamppb.New(session.WaktuMulai),
		WaktuSelesai:    waktuSelesai,
		BatasWaktu:      batasWaktu,
		DurasiMenit:     int32(session.DurasiMenit),
		NilaiAkhir:      nilaiAkhir,
		JumlahBenar:     jumlahBenar,
		TotalSoal:       totalSoal,
		Status:          status,
	}
}

// convertUserToProto converts entity.User to proto User
func (h *historyHandler) convertUserToProto(user *entity.User) *base.User {
	if user == nil {
		return nil
	}

	role := base.UserRole(base.UserRole_value[strings.ToUpper(user.Role)])

	return &base.User{
		Id:       int32(user.ID),
		Email:    user.Email,
		Nama:     user.Nama,
		Role:     role,
		IsActive: user.IsActive,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

// convertHistoriesToProto converts []entity.HistorySummary to []*base.HistorySummary
func (h *historyHandler) convertHistoriesToProto(histories []entity.HistorySummary) []*base.HistorySummary {
	protoHistories := make([]*base.HistorySummary, len(histories))
	for i, h := range histories {
		var waktuMulai, waktuSelesai *timestamppb.Timestamp
		if h.WaktuMulai != nil {
			waktuMulai = timestamppb.New(*h.WaktuMulai)
		}
		if h.WaktuSelesai != nil {
			waktuSelesai = timestamppb.New(*h.WaktuSelesai)
		}

		status := base.TestStatus(base.TestStatus_value[strings.ToUpper(string(h.Status))])

		protoHistories[i] = &base.HistorySummary{
			Id:                    int32(h.ID),
			SessionToken:          h.SessionToken,
			MataPelajaran:         &base.MataPelajaran{Id: int32(h.MataPelajaran.ID), Nama: h.MataPelajaran.Nama},
			Tingkat:               &base.Tingkat{Id: int32(h.Tingkat.ID), Nama: h.Tingkat.Nama},
			WaktuMulai:            waktuMulai,
			WaktuSelesai:          waktuSelesai,
			DurasiPengerjaanDetik: int32(h.DurasiPengerjaanDetik),
			NilaiAkhir:            h.NilaiAkhir,
			JumlahBenar:           int32(h.JumlahBenar),
			TotalSoal:             int32(h.TotalSoal),
			Status:                status,
			NamaPeserta:           h.NamaPeserta,
		}
	}
	return protoHistories
}

// ListStudentHistories lists all student histories (admin only)
func (h *historyHandler) ListStudentHistories(ctx context.Context, req *base.ListStudentHistoriesRequest) (*base.ListStudentHistoriesResponse, error) {
	// Get user from context
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		// For REST gateway, extract token from metadata
		token, extractErr := interceptor.ExtractTokenFromContext(ctx)
		if extractErr != nil {
			return nil, status.Error(codes.Unauthenticated, "user not authenticated")
		}
		claims, validateErr := interceptor.ValidateToken(token)
		if validateErr != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		user = &base.User{
			Id:    claims.UserID,
			Email: claims.Email,
			Role:  base.UserRole(claims.Role),
		}
		// Add to context for consistency
		ctx = interceptor.AddUserToContext(ctx, claims)
	}

	// Check if user is admin
	if user.Role != base.UserRole_ADMIN {
		return nil, status.Error(codes.PermissionDenied, "admin access required")
	}

	page := 1
	pageSize := 20
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = int(req.Pagination.Page)
		}
		if req.Pagination.PageSize > 0 {
			pageSize = int(req.Pagination.PageSize)
			// Cap at 50 for list histories (heavy N+ query)
			if pageSize > 50 {
				pageSize = 50
			}
		}
	}

	var userID, tingkatan, idMataPelajaran *int
	if req.UserId > 0 {
		userID = &[]int{int(req.UserId)}[0]
	}
	if req.Tingkatan > 0 {
		tingkatan = &[]int{int(req.Tingkatan)}[0]
	}
	if req.IdMataPelajaran > 0 {
		idMataPelajaran = &[]int{int(req.IdMataPelajaran)}[0]
	}

	histories, total, err := h.usecase.ListStudentHistories(userID, tingkatan, idMataPelajaran, page, pageSize)
	if err != nil {
		return &base.ListStudentHistoriesResponse{
			HistoryPerStudent: []*base.StudentHistoryWithUser{},
			Pagination: &base.PaginationResponse{
				TotalCount:  0,
				TotalPages:  0,
				CurrentPage: int32(page),
				PageSize:    int32(pageSize),
			},
		}, nil
	}

	// Convert to proto
	protoHistories := make([]*base.StudentHistoryWithUser, len(histories))
	for i, hist := range histories {
		protoHistories[i] = &base.StudentHistoryWithUser{
			User:               h.convertUserToProto(&hist.User),
			History:            h.convertHistoriesToProto(hist.History),
			RataRataNilai:      hist.RataRataNilai,
			TotalTestCompleted: int32(hist.TotalTestCompleted),
		}
	}

	totalPages := (total + pageSize - 1) / pageSize

	return &base.ListStudentHistoriesResponse{
		HistoryPerStudent: protoHistories,
		Pagination: &base.PaginationResponse{
			TotalCount:  int32(total),
			TotalPages:  int32(totalPages),
			CurrentPage: int32(page),
			PageSize:    int32(pageSize),
		},
	}, nil
}