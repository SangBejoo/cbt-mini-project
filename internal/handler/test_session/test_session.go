package test_session

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	userLimitUsecase "cbt-test-mini-project/internal/usecase"
	"cbt-test-mini-project/internal/usecase/test_session"
	tingkatUsecase "cbt-test-mini-project/internal/usecase/tingkat"
	"cbt-test-mini-project/util/interceptor"
	"context"
	"errors"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// testSessionHandler implements base.TestSessionServiceServer
type testSessionHandler struct {
	base.UnimplementedTestSessionServiceServer
	usecase           test_session.TestSessionUsecase
	tingkatUsecase    tingkatUsecase.TingkatUsecase
	userLimitUsecase  userLimitUsecase.UserLimitUsecase
}

// NewTestSessionHandler creates a new TestSessionHandler
func NewTestSessionHandler(usecase test_session.TestSessionUsecase, tingkatUsecase tingkatUsecase.TingkatUsecase, userLimitUsecase userLimitUsecase.UserLimitUsecase) base.TestSessionServiceServer {
	return &testSessionHandler{
		usecase:          usecase,
		tingkatUsecase:   tingkatUsecase,
		userLimitUsecase: userLimitUsecase,
	}
}

// CreateTestSession creates a new test session
func (h *testSessionHandler) CreateTestSession(ctx context.Context, req *base.CreateTestSessionRequest) (*base.TestSessionResponse, error) {
	// Get user_id from JWT context
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

	// Check user limits before creating test session
	if err := h.userLimitUsecase.IncrementUsage(ctx, int(userID), entity.LimitTypeTestSessionsPerDay, nil); err != nil {
		return nil, status.Error(codes.ResourceExhausted, "Daily test session limit exceeded. Please try again tomorrow.")
	}

	session, err := h.usecase.CreateTestSession(int(userID), int(req.IdTingkat), int(req.IdMataPelajaran), int(req.DurasiMenit), int(req.JumlahSoal))
	if err != nil {
		// If session creation fails, we should decrement the usage counter
		// For now, we'll rely on the rate limit middleware to handle this
		// In a production system, you might want to implement a rollback mechanism
		return nil, err
	}

	return &base.TestSessionResponse{
		TestSession: h.convertToProtoTestSession(session),
	}, nil
}

// GetTestSession gets session by token
func (h *testSessionHandler) GetTestSession(ctx context.Context, req *base.GetTestSessionRequest) (*base.TestSessionResponse, error) {
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

	session, err := h.usecase.GetTestSession(req.SessionToken)
	if err != nil {
		return nil, err
	}

	// Check if the session belongs to the authenticated user
	if session.UserID == nil || *session.UserID != int(user.Id) {
		return nil, status.Error(codes.PermissionDenied, "you do not have permission to access this session")
	}

	return &base.TestSessionResponse{
		TestSession: h.convertToProtoTestSession(session),
	}, nil
}

// GetTestQuestions gets all questions for the session
func (h *testSessionHandler) GetTestQuestions(ctx context.Context, req *base.GetTestQuestionsRequest) (*base.TestQuestionsResponse, error) {
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

	session, err := h.usecase.GetTestSession(req.SessionToken)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, errors.New("session not found")
	}

	// Check if the session belongs to the authenticated user
	if session.UserID == nil || *session.UserID != int(user.Id) {
		return nil, status.Error(codes.PermissionDenied, "you do not have permission to access this session")
	}

	soals, err := h.usecase.GetAllTestQuestions(req.SessionToken)
	if err != nil {
		return nil, err
	}

	// Get answers status
	answers, _ := h.usecase.GetSessionAnswers(req.SessionToken)
	isAnsweredStatus := make([]bool, len(soals))
	for i := range soals {
		for _, ans := range answers {
			if ans.TestSessionSoal.NomorUrut == soals[i].NomorUrut {
				isAnsweredStatus[i] = true
				break
			}
		}
	}

	var protoSoals []*base.SoalForStudent
	for _, s := range soals {
		var jawabanDipilih base.JawabanOption
		if s.JawabanDipilih != nil {
			jawabanDipilih = base.JawabanOption(base.JawabanOption_value[string(*s.JawabanDipilih)])
		}

		protoSoals = append(protoSoals, &base.SoalForStudent{
			Id:             int32(s.ID),
			NomorUrut:      int32(s.NomorUrut),
			Pertanyaan:     s.Pertanyaan,
			OpsiA:          s.OpsiA,
			OpsiB:          s.OpsiB,
			OpsiC:          s.OpsiC,
			OpsiD:          s.OpsiD,
			JawabanDipilih: jawabanDipilih,
			IsAnswered:     s.IsAnswered,
			Materi: &base.Materi{
				Id:             int32(s.Materi.ID),
				Nama:           s.Materi.Nama,
				MataPelajaran:  &base.MataPelajaran{Id: int32(s.Materi.MataPelajaran.ID), Nama: s.Materi.MataPelajaran.Nama},
				Tingkat:        &base.Tingkat{Id: int32(s.Materi.Tingkat.ID), Nama: s.Materi.Tingkat.Nama},
			},
			Gambar:         convertSoalGambarToProto(s.Gambar),
		})
	}

	return &base.TestQuestionsResponse{
		SessionToken:      req.SessionToken,
		Soal:              protoSoals,
		TotalSoal:         int32(len(protoSoals)),
		CurrentNomorUrut:  1, // Not used
		DijawabCount:      int32(len(answers)),
		IsAnsweredStatus:  isAnsweredStatus,
		BatasWaktu:        timestamppb.New(session.BatasWaktu()),
	}, nil
}

// SubmitAnswer submits an answer
func (h *testSessionHandler) SubmitAnswer(ctx context.Context, req *base.SubmitAnswerRequest) (*base.SubmitAnswerResponse, error) {
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

	// Get session to check ownership
	session, err := h.usecase.GetTestSession(req.SessionToken)
	if err != nil {
		return nil, err
	}

	// Check if the session belongs to the authenticated user
	if session.UserID == nil || *session.UserID != int(user.Id) {
		return nil, status.Error(codes.PermissionDenied, "you do not have permission to access this session")
	}

	jawaban := entity.JawabanOption(req.JawabanDipilih.String()[0])
	err = h.usecase.SubmitAnswer(req.SessionToken, int(req.NomorUrut), jawaban)
	if err != nil {
		return nil, err
	}

	return &base.SubmitAnswerResponse{
		SessionToken:    req.SessionToken,
		NomorUrut:       req.NomorUrut,
		JawabanDipilih:  req.JawabanDipilih,
		IsCorrect:       true, // TODO: get from usecase
		DijawabPada:     timestamppb.Now(),
	}, nil
}

// ClearAnswer clears an answer
func (h *testSessionHandler) ClearAnswer(ctx context.Context, req *base.ClearAnswerRequest) (*base.ClearAnswerResponse, error) {
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

	// Get session to check ownership
	session, err := h.usecase.GetTestSession(req.SessionToken)
	if err != nil {
		return nil, err
	}

	// Check if the session belongs to the authenticated user
	if session.UserID == nil || *session.UserID != int(user.Id) {
		return nil, status.Error(codes.PermissionDenied, "you do not have permission to access this session")
	}

	err = h.usecase.ClearAnswer(req.SessionToken, int(req.NomorUrut))
	if err != nil {
		return nil, err
	}

	return &base.ClearAnswerResponse{
		SessionToken:   req.SessionToken,
		NomorUrut:      req.NomorUrut,
		DibatalkanPada: timestamppb.Now(),
	}, nil
}

// CompleteSession completes the session
func (h *testSessionHandler) CompleteSession(ctx context.Context, req *base.CompleteSessionRequest) (*base.TestSessionResponse, error) {
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

	// Get session to check ownership
	session, err := h.usecase.GetTestSession(req.SessionToken)
	if err != nil {
		return nil, err
	}

	// Check if the session belongs to the authenticated user
	if session.UserID == nil || *session.UserID != int(user.Id) {
		return nil, status.Error(codes.PermissionDenied, "you do not have permission to access this session")
	}

	session, err = h.usecase.CompleteSession(req.SessionToken)
	if err != nil {
		return nil, err
	}

	return &base.TestSessionResponse{
		TestSession: h.convertToProtoTestSession(session),
	}, nil
}

// GetTestResult gets test result
func (h *testSessionHandler) GetTestResult(ctx context.Context, req *base.GetTestResultRequest) (*base.TestResultResponse, error) {
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

	session, details, err := h.usecase.GetTestResult(req.SessionToken)
	if err != nil {
		return nil, err
	}

	// Check if the session belongs to the authenticated user
	if session.UserID == nil || *session.UserID != int(user.Id) {
		return nil, status.Error(codes.PermissionDenied, "you do not have permission to access this session result")
	}

	// Get all tingkat
	tingkatList, _, err := h.tingkatUsecase.ListTingkat(1, 100) // Assuming max 100 tingkat
	if err != nil {
		return nil, err
	}

	var jawabanDetails []*base.JawabanDetail
	for _, d := range details {
		var jawabanDipilih base.JawabanOption
		if d.JawabanDipilih != nil {
			jawabanDipilih = base.JawabanOption(base.JawabanOption_value[string(*d.JawabanDipilih)])
		}

		var pembahasan string
		if d.Pembahasan != nil {
			pembahasan = *d.Pembahasan
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
			IsAnswered:     d.IsAnswered,
			Pembahasan:     pembahasan,
			Gambar:         convertSoalGambarToProto(d.Gambar),
		})
	}

	var protoTingkat []*base.Tingkat
	for _, t := range tingkatList {
		protoTingkat = append(protoTingkat, &base.Tingkat{
			Id:   int32(t.ID),
			Nama: t.Nama,
		})
	}

	return &base.TestResultResponse{
		SessionInfo:   h.convertToProtoTestSession(session),
		DetailJawaban: jawabanDetails,
		Tingkat:       protoTingkat,
	}, nil
}

// ListTestSessions lists sessions
func (h *testSessionHandler) ListTestSessions(ctx context.Context, req *base.ListTestSessionsRequest) (*base.ListTestSessionsResponse, error) {
	// Get user from JWT context for admin access
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
		return nil, status.Error(codes.PermissionDenied, "only admin can access this endpoint")
	}

	var tingkatan, idMataPelajaran *int
	var status *entity.TestStatus

	if req.IdTingkat != 0 {
		t := int(req.IdTingkat)
		tingkatan = &t
	}
	if req.IdMataPelajaran != 0 {
		i := int(req.IdMataPelajaran)
		idMataPelajaran = &i
	}
	if req.Status != base.TestStatus_STATUS_INVALID {
		s := entity.TestStatus(req.Status.String())
		status = &s
	}

	page := 1
	pageSize := 1000
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = int(req.Pagination.Page)
		}
		if req.Pagination.PageSize > 0 {
			pageSize = int(req.Pagination.PageSize)
		}
	}

	sessions, pagination, err := h.usecase.ListTestSessions(tingkatan, idMataPelajaran, status, page, pageSize)
	if err != nil {
		return nil, err
	}

	var sessionList []*base.TestSession
	for _, s := range sessions {
		sessionList = append(sessionList, h.convertToProtoTestSession(&s))
	}

	return &base.ListTestSessionsResponse{
		TestSessions: sessionList,
		Pagination: &base.PaginationResponse{
			TotalCount:  int32(pagination.TotalCount),
			TotalPages:  int32(pagination.TotalPages),
			CurrentPage: int32(pagination.CurrentPage),
			PageSize:    int32(pagination.PageSize),
		},
	}, nil
}

// Helper function to convert entity to proto
func (h *testSessionHandler) convertToProtoTestSession(session *entity.TestSession) *base.TestSession {
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
		NamaPeserta:     session.NamaPeserta,
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
func (h *testSessionHandler) convertUserToProto(user *entity.User) *base.User {
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
func convertSoalGambarToProto(gambar []entity.SoalGambar) []*base.SoalGambar {
	if len(gambar) == 0 {
		return nil
	}
	
	var protoGambar []*base.SoalGambar
	for _, g := range gambar {
		protoGambar = append(protoGambar, &base.SoalGambar{
			Id:       int32(g.ID),
			FilePath: g.FilePath,
			Urutan:   int32(g.Urutan),
		})
	}
	return protoGambar
}