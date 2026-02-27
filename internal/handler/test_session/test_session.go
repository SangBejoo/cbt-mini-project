package test_session

import (
	base "cbt-test-mini-project/gen/proto"
	"cbt-test-mini-project/internal/entity"
	userLimitUsecase "cbt-test-mini-project/internal/usecase"
	"cbt-test-mini-project/internal/usecase/materi"
	"cbt-test-mini-project/internal/usecase/test_session"
	tingkatUsecase "cbt-test-mini-project/internal/usecase/tingkat"
	"cbt-test-mini-project/util/interceptor"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// testSessionHandler implements base.TestSessionServiceServer
type testSessionHandler struct {
	base.UnimplementedTestSessionServiceServer
	usecase          test_session.TestSessionUsecase
	materiUsecase    materi.MateriUsecase
	tingkatUsecase   tingkatUsecase.TingkatUsecase
	userLimitUsecase userLimitUsecase.UserLimitUsecase
}

// NewTestSessionHandler creates a new TestSessionHandler
func NewTestSessionHandler(usecase test_session.TestSessionUsecase, materiUsecase materi.MateriUsecase, tingkatUsecase tingkatUsecase.TingkatUsecase, userLimitUsecase userLimitUsecase.UserLimitUsecase) base.TestSessionServiceServer {
	return &testSessionHandler{
		usecase:          usecase,
		materiUsecase:    materiUsecase,
		tingkatUsecase:   tingkatUsecase,
		userLimitUsecase: userLimitUsecase,
	}
}

// CreateTestSession creates a new test session
func (h *testSessionHandler) CreateTestSession(ctx context.Context, req *base.CreateTestSessionRequest) (*base.TestSessionResponse, error) {
	// DEBUG: Catch any panics and log them
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC in CreateTestSession: %v\n", r)
		}
	}()

	fmt.Printf("=== CreateTestSession called with req: %+v ===\n", req)

	// Get user_id from JWT context
	user, err := interceptor.GetUserFromContext(ctx)
	fmt.Printf("=== GetUserFromContext result: user=%+v, err=%v ===\n", user, err)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	userID := user.Id
	fmt.Printf("=== UserID: %d ===\n", userID)

	// Get durasi_menit and jumlah_soal from materi if not provided (or use defaults if provided)
	// For now we always use materi defaults - siswa tidak bisa custom durasi/jumlah soal
	durasiMenit := int(req.DurasiMenit)
	jumlahSoal := int(req.JumlahSoal)

	// If client provides 0, get from materi defaults
	// Query to get a materi with these tingkat dan mataPelajaran to get its defaults
	// For now, we'll use the request values or defaults
	if durasiMenit == 0 {
		durasiMenit = 60 // fallback default
	}
	if jumlahSoal == 0 {
		jumlahSoal = 20 // fallback default
	}

	includeTypes := make([]entity.QuestionType, 0, len(req.IncludeQuestionTypes))
	for _, questionType := range req.IncludeQuestionTypes {
		switch questionType {
		case base.QuestionType_MULTIPLE_CHOICE:
			includeTypes = append(includeTypes, entity.QuestionTypeMultipleChoice)
		case base.QuestionType_DRAG_DROP:
			includeTypes = append(includeTypes, entity.QuestionTypeDragDrop)
		case base.QuestionType_ESSAY:
			includeTypes = append(includeTypes, entity.QuestionTypeEssay)
		case base.QuestionType_MULTIPLE_CHOICES_COMPLEX:
			includeTypes = append(includeTypes, entity.QuestionTypeMultipleChoicesComplex)
		}
	}

	randomize := true
	if req.SelectionMode == base.QuestionSelectionMode_ORDERED {
		randomize = false
	}

	session, err := h.usecase.CreateTestSession(int(userID), int(req.IdTingkat), int(req.IdMataPelajaran), durasiMenit, jumlahSoal, includeTypes, randomize)
	if err != nil {
		fmt.Printf("=== HANDLER: CreateTestSession FAILED: %v ===\n", err)
		return nil, status.Error(codes.Internal, err.Error())
	}
	if session == nil {
		fmt.Printf("=== HANDLER: CreateTestSession returned nil session without error ===\n")
		return nil, status.Error(codes.Internal, "session creation returned nil")
	}
	fmt.Printf("=== HANDLER: CreateTestSession SUCCESS: sessionID=%d, token=%s ===\n", session.ID, session.SessionToken)

	// Check user limits after successful session creation
	fmt.Printf("=== HANDLER: About to increment usage for user %d ===\n", userID)
	if err := h.userLimitUsecase.IncrementUsage(ctx, int(userID), entity.LimitTypeTestSessionsPerDay, &session.ID); err != nil {
		fmt.Printf("=== HANDLER: IncrementUsage failed: %v ===\n", err)
		return nil, status.Error(codes.ResourceExhausted, "Daily test session limit exceeded. Please try again tomorrow.")
	}
	fmt.Printf("=== HANDLER: IncrementUsage success for user %d ===\n", userID)

	return &base.TestSessionResponse{
		TestSession: h.convertToProtoTestSession(session),
	}, nil
}

// GetTestSession gets session by token
func (h *testSessionHandler) GetTestSession(ctx context.Context, req *base.GetTestSessionRequest) (*base.TestSessionResponse, error) {
	// Get user from JWT context
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
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
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC in GetTestQuestions: %v\n", r)
		}
	}()

	// Get user from JWT context
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
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

	// DEBUG: Log session and soals info
	fmt.Printf("DEBUG GetTestQuestions - Token: %s, SessionID: %d, Status: %s, WaktuMulai: %v, BatasWaktu: %v, Soals count: %d, Now: %v\n",
		req.SessionToken, session.ID, session.Status, session.WaktuMulai, session.BatasWaktu(), len(soals), time.Now())

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

	var protoQuestions []*base.QuestionForStudent
	for i, q := range soals {
		// Add more detailed logging
		fmt.Printf("Processing question %d: NomorUrut=%d, QuestionType=%s\n", i, q.NomorUrut, q.QuestionType)

		protoQuestion := &base.QuestionForStudent{
			NomorUrut:    int32(q.NomorUrut),
			QuestionType: base.QuestionType(base.QuestionType_value[strings.ToUpper(string(q.QuestionType))]),
			IsAnswered:   q.IsAnswered,
		}

		// Add Materi
		if q.Materi.ID > 0 {
			protoQuestion.Materi = &base.Materi{
				Id:                 int32(q.Materi.ID),
				MataPelajaran:      &base.MataPelajaran{Id: int32(q.Materi.MataPelajaran.ID), Nama: q.Materi.MataPelajaran.Nama},
				Tingkat:            &base.Tingkat{Id: int32(q.Materi.Tingkat.ID), Nama: q.Materi.Tingkat.Nama},
				Nama:               q.Materi.Nama,
				IsActive:           q.Materi.IsActive,
				DefaultDurasiMenit: int32(q.Materi.DefaultDurasiMenit),
				DefaultJumlahSoal:  int32(q.Materi.DefaultJumlahSoal),
			}
		}

		// Handle multiple choice fields
		if q.QuestionType == entity.QuestionTypeMultipleChoice && q.MCID != nil {
			protoQuestion.McId = int32(*q.MCID)
			if q.MCPertanyaan != nil {
				protoQuestion.McPertanyaan = *q.MCPertanyaan
			}
			if q.MCOpsiA != nil {
				protoQuestion.McOpsiA = *q.MCOpsiA
			}
			if q.MCOpsiB != nil {
				protoQuestion.McOpsiB = *q.MCOpsiB
			}
			if q.MCOpsiC != nil {
				protoQuestion.McOpsiC = *q.MCOpsiC
			}
			if q.MCOpsiD != nil {
				protoQuestion.McOpsiD = *q.MCOpsiD
			}
			if q.MCJawabanDipilih != nil {
				if val, ok := base.JawabanOption_value[string(*q.MCJawabanDipilih)]; ok {
					protoQuestion.McJawabanDipilih = base.JawabanOption(val)
				}
			}
			protoQuestion.McGambar = convertSoalGambarToProto(q.MCGambar)
		}

		if q.QuestionType == entity.QuestionTypeMultipleChoicesComplex && q.MCCID != nil {
			protoQuestion.MccId = int32(*q.MCCID)
			if q.MCCPertanyaan != nil {
				protoQuestion.MccPertanyaan = *q.MCCPertanyaan
			}
			if q.MCCOpsiA != nil {
				protoQuestion.MccOpsiA = *q.MCCOpsiA
			}
			if q.MCCOpsiB != nil {
				protoQuestion.MccOpsiB = *q.MCCOpsiB
			}
			if q.MCCOpsiC != nil {
				protoQuestion.MccOpsiC = *q.MCCOpsiC
			}
			if q.MCCOpsiD != nil {
				protoQuestion.MccOpsiD = *q.MCCOpsiD
			}
			protoQuestion.MccJawabanDipilih = toProtoJawabanOptions(q.MCCJawabanDipilih)
			protoQuestion.MccGambar = convertSoalGambarToProto(q.MCCGambar)
		}

		// Handle drag-drop fields
		if q.QuestionType == entity.QuestionTypeDragDrop && q.DDID != nil {
			protoQuestion.DdId = int32(*q.DDID)
			if q.DDPertanyaan != nil {
				protoQuestion.DdPertanyaan = *q.DDPertanyaan
			}
			if q.DDDDragType != nil {
				protoQuestion.DdDragType = base.DragDropType(base.DragDropType_value[strings.ToUpper(string(*q.DDDDragType))])
			}
			protoQuestion.DdItems = convertDragItemsToProto(q.DDItems)
			protoQuestion.DdSlots = convertDragSlotsToProto(q.DDSlots)
			if q.DDUserAnswer != nil {
				protoQuestion.DdUserAnswer = make(map[int32]int32)
				for k, v := range q.DDUserAnswer {
					protoQuestion.DdUserAnswer[int32(k)] = int32(v)
				}
			}
		}

		if q.QuestionType == entity.QuestionTypeEssay && q.EssayID != nil {
			protoQuestion.EssayId = int32(*q.EssayID)
			if q.EssayPertanyaan != nil {
				protoQuestion.EssayPertanyaan = *q.EssayPertanyaan
			}
			if q.EssayJawaban != nil {
				protoQuestion.EssayJawaban = *q.EssayJawaban
			}
		}

		protoQuestions = append(protoQuestions, protoQuestion)
	}

	// Don't return error if no questions, just log warning
	if len(protoQuestions) == 0 {
		fmt.Printf("WARNING: No valid questions found for session %s\n", req.SessionToken)
	}

	fmt.Printf("DEBUG: Creating response with %d questions\n", len(protoQuestions))

	response := &base.TestQuestionsResponse{
		SessionToken:     req.SessionToken,
		Questions:        protoQuestions,
		TotalSoal:        int32(len(protoQuestions)),
		CurrentNomorUrut: 1, // Not used
		DijawabCount:     int32(len(answers)),
		IsAnsweredStatus: isAnsweredStatus,
	}

	// Add BatasWaktu carefully to avoid panic
	if session.BatasWaktu() != (time.Time{}) {
		response.BatasWaktu = timestamppb.New(session.BatasWaktu())
	} else {
		fmt.Printf("WARNING: Invalid BatasWaktu for session %s\n", req.SessionToken)
		// Use current time + 1 hour as fallback
		response.BatasWaktu = timestamppb.New(time.Now().Add(time.Hour))
	}

	fmt.Printf("DEBUG: Response created successfully\n")
	return response, nil
}

// SubmitAnswer submits an answer
func (h *testSessionHandler) SubmitAnswer(ctx context.Context, req *base.SubmitAnswerRequest) (*base.SubmitAnswerResponse, error) {
	// Get user from JWT context
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
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
		SessionToken:   req.SessionToken,
		NomorUrut:      req.NomorUrut,
		JawabanDipilih: req.JawabanDipilih,
		IsCorrect:      true, // TODO: get from usecase
		DijawabPada:    timestamppb.Now(),
	}, nil
}

func (h *testSessionHandler) SubmitComplexAnswer(ctx context.Context, req *base.SubmitComplexAnswerRequest) (*base.SubmitComplexAnswerResponse, error) {
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	session, err := h.usecase.GetTestSession(req.SessionToken)
	if err != nil {
		return nil, err
	}

	if session.UserID == nil || *session.UserID != int(user.Id) {
		return nil, status.Error(codes.PermissionDenied, "you do not have permission to access this session")
	}

	jawaban := toEntityJawabanOptions(req.JawabanDipilih)
	err = h.usecase.SubmitComplexAnswer(req.SessionToken, int(req.NomorUrut), jawaban)
	if err != nil {
		return nil, err
	}

	return &base.SubmitComplexAnswerResponse{
		SessionToken:   req.SessionToken,
		NomorUrut:      req.NomorUrut,
		JawabanDipilih: req.JawabanDipilih,
		IsCorrect:      true,
		DijawabPada:    timestamppb.Now(),
	}, nil
}

// SubmitDragDropAnswer submits a drag-drop answer
func (h *testSessionHandler) SubmitDragDropAnswer(ctx context.Context, req *base.SubmitDragDropAnswerRequest) (*base.SubmitDragDropAnswerResponse, error) {
	// Get user from JWT context
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
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

	// Convert map[int32]int32 to map[int]int
	answer := make(map[int]int)
	for k, v := range req.Answer {
		answer[int(k)] = int(v)
	}

	err = h.usecase.SubmitDragDropAnswer(req.SessionToken, int(req.NomorUrut), answer)
	if err != nil {
		return nil, err
	}

	return &base.SubmitDragDropAnswerResponse{
		SessionToken: req.SessionToken,
		NomorUrut:    req.NomorUrut,
		Answer:       req.Answer,
		IsCorrect:    true, // Determined by usecase
		DijawabPada:  timestamppb.Now(),
	}, nil
}

func (h *testSessionHandler) SubmitEssayAnswer(ctx context.Context, req *base.SubmitEssayAnswerRequest) (*base.SubmitEssayAnswerResponse, error) {
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	session, err := h.usecase.GetTestSession(req.SessionToken)
	if err != nil {
		return nil, err
	}

	if session.UserID == nil || *session.UserID != int(user.Id) {
		return nil, status.Error(codes.PermissionDenied, "you do not have permission to access this session")
	}

	err = h.usecase.SubmitEssayAnswer(req.SessionToken, int(req.NomorUrut), req.JawabanEssay)
	if err != nil {
		return nil, err
	}

	return &base.SubmitEssayAnswerResponse{
		SessionToken: req.SessionToken,
		NomorUrut:    req.NomorUrut,
		JawabanEssay: req.JawabanEssay,
		DijawabPada:  timestamppb.Now(),
	}, nil
}

// ClearAnswer clears an answer
func (h *testSessionHandler) ClearAnswer(ctx context.Context, req *base.ClearAnswerRequest) (*base.ClearAnswerResponse, error) {
	// Get user from JWT context
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
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
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
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
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
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

		jawabanDetail := &base.JawabanDetail{
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
			QuestionType:   base.QuestionType(base.QuestionType_value[strings.ToUpper(string(d.QuestionType))]),
		}

		if d.QuestionType == entity.QuestionTypeEssay {
			if d.JawabanEssay != nil {
				jawabanDetail.JawabanEssay = *d.JawabanEssay
			}
			if d.NilaiEssay != nil {
				jawabanDetail.NilaiEssay = *d.NilaiEssay
			}
			if d.FeedbackTeacher != nil {
				jawabanDetail.FeedbackTeacher = *d.FeedbackTeacher
			}
		}

		if d.QuestionType == entity.QuestionTypeMultipleChoicesComplex {
			jawabanDetail.JawabanDipilihComplex = toProtoJawabanOptions(d.JawabanDipilihComplex)
			jawabanDetail.JawabanBenarComplex = toProtoJawabanOptions(d.JawabanBenarComplex)
		}

		if d.QuestionType == entity.QuestionTypeDragDrop {
			if d.DragType != nil {
				jawabanDetail.DragType = base.DragDropType(base.DragDropType_value[strings.ToUpper(string(*d.DragType))])
			}
			jawabanDetail.Items = convertDragItemsToProto(d.DragItems)
			jawabanDetail.Slots = convertDragSlotsToProto(d.DragSlots)

			if d.UserDragAnswer != nil {
				jawabanDetail.UserDragAnswer = make(map[int32]int32)
				for k, v := range d.UserDragAnswer {
					jawabanDetail.UserDragAnswer[int32(k)] = int32(v)
				}
			}
			if d.CorrectDragAnswer != nil {
				jawabanDetail.CorrectDragAnswer = make(map[int32]int32)
				for k, v := range d.CorrectDragAnswer {
					jawabanDetail.CorrectDragAnswer[int32(k)] = int32(v)
				}
			}
		}

		jawabanDetails = append(jawabanDetails, jawabanDetail)
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

func toEntityJawabanOptions(options []base.JawabanOption) []entity.JawabanOption {
	result := make([]entity.JawabanOption, 0, len(options))
	for _, option := range options {
		switch option {
		case base.JawabanOption_A:
			result = append(result, entity.JawabanA)
		case base.JawabanOption_B:
			result = append(result, entity.JawabanB)
		case base.JawabanOption_C:
			result = append(result, entity.JawabanC)
		case base.JawabanOption_D:
			result = append(result, entity.JawabanD)
		}
	}
	return result
}

func toProtoJawabanOptions(options []entity.JawabanOption) []base.JawabanOption {
	result := make([]base.JawabanOption, 0, len(options))
	for _, option := range options {
		switch option {
		case entity.JawabanA:
			result = append(result, base.JawabanOption_A)
		case entity.JawabanB:
			result = append(result, base.JawabanOption_B)
		case entity.JawabanC:
			result = append(result, base.JawabanOption_C)
		case entity.JawabanD:
			result = append(result, base.JawabanOption_D)
		}
	}
	return result
}

func (h *testSessionHandler) GradeEssayAnswer(ctx context.Context, req *base.GradeEssayAnswerRequest) (*base.GradeEssayAnswerResponse, error) {
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if user.Role != base.UserRole_ADMIN {
		return nil, status.Error(codes.PermissionDenied, "only admin can grade essay answers")
	}

	if err := h.usecase.GradeEssayAnswer(int(req.AnswerId), req.Score, req.Feedback); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &base.GradeEssayAnswerResponse{
		Success: true,
		Message: "Essay answer graded successfully",
	}, nil
}

// ListMyScheduledSessions lists scheduled sessions for the authenticated student.
func (h *testSessionHandler) ListMyScheduledSessions(ctx context.Context, req *base.ListMyScheduledSessionsRequest) (*base.ListTestSessionsResponse, error) {
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if user.Role != base.UserRole_SISWA {
		return nil, status.Error(codes.PermissionDenied, "only student can access this endpoint")
	}

	page := 1
	pageSize := 20
	if req.Pagination != nil {
		if req.Pagination.Page > 0 {
			page = int(req.Pagination.Page)
		}
		if req.Pagination.PageSize > 0 {
			pageSize = int(req.Pagination.PageSize)
		}
	}

	var lmsClassID *int64
	if req.LmsClassId > 0 {
		v := req.LmsClassId
		lmsClassID = &v
	}

	sessions, pagination, err := h.usecase.ListScheduledSessions(int(user.Id), lmsClassID, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	protoSessions := make([]*base.TestSession, 0, len(sessions))
	for i := range sessions {
		protoSessions = append(protoSessions, h.convertToProtoTestSession(&sessions[i]))
	}

	return &base.ListTestSessionsResponse{
		TestSessions: protoSessions,
		Pagination: &base.PaginationResponse{
			TotalCount:  int32(pagination.TotalCount),
			TotalPages:  int32(pagination.TotalPages),
			CurrentPage: int32(pagination.CurrentPage),
			PageSize:    int32(pagination.PageSize),
		},
	}, nil
}

// StartScheduledSession starts an LMS-scheduled session when the start time is reached.
func (h *testSessionHandler) StartScheduledSession(ctx context.Context, req *base.StartScheduledSessionRequest) (*base.TestSessionResponse, error) {
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
	}

	if user.Role != base.UserRole_SISWA {
		return nil, status.Error(codes.PermissionDenied, "only student can access this endpoint")
	}

	session, err := h.usecase.StartScheduledSession(int(user.Id), req.SessionToken)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "scheduled to start") {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		if strings.Contains(strings.ToLower(err.Error()), "permission") {
			return nil, status.Error(codes.PermissionDenied, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &base.TestSessionResponse{TestSession: h.convertToProtoTestSession(session)}, nil
}

// ListTestSessions lists sessions
func (h *testSessionHandler) ListTestSessions(ctx context.Context, req *base.ListTestSessionsRequest) (*base.ListTestSessionsResponse, error) {
	// Get user from JWT context for admin access
	user, err := interceptor.GetUserFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "user not authenticated")
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
	if session == nil {
		return nil
	}

	var waktuSelesai, batasWaktu *timestamppb.Timestamp
	if session.WaktuSelesai != nil {
		waktuSelesai = timestamppb.New(*session.WaktuSelesai)
	}
	batasWaktu = timestamppb.New(session.BatasWaktu())

	var nilaiAkhir float64
	if session.NilaiAkhir != nil {
		nilaiAkhir = *session.NilaiAkhir
	}

	var lmsClassID int64
	if session.LMSClassID != nil {
		lmsClassID = *session.LMSClassID
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
		Id:            int32(session.ID),
		SessionToken:  session.SessionToken,
		User:          h.convertUserToProto(session.User),
		NamaPeserta:   session.NamaPeserta,
		Tingkat:       &base.Tingkat{Id: int32(session.Tingkat.ID), Nama: session.Tingkat.Nama},
		MataPelajaran: &base.MataPelajaran{Id: int32(session.MataPelajaran.ID), Nama: session.MataPelajaran.Nama},
		WaktuMulai:    timestamppb.New(session.WaktuMulai),
		WaktuSelesai:  waktuSelesai,
		BatasWaktu:    batasWaktu,
		DurasiMenit:   int32(session.DurasiMenit),
		NilaiAkhir:    nilaiAkhir,
		JumlahBenar:   jumlahBenar,
		TotalSoal:     totalSoal,
		Status:        status,
		LmsClassId:    lmsClassID,
	}
}

// convertUserToProto converts entity.User to proto User
func (h *testSessionHandler) convertUserToProto(user *entity.User) *base.User {
	if user == nil {
		return nil
	}

	role := base.UserRole(base.UserRole_value[strings.ToUpper(user.Role)])

	return &base.User{
		Id:        int32(user.ID),
		Email:     user.Email,
		Nama:      user.Nama,
		Role:      role,
		IsActive:  user.IsActive,
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
		keterangan := ""
		if g.Keterangan != nil {
			keterangan = *g.Keterangan
		}

		cloudId := ""
		if g.CloudId != nil {
			cloudId = *g.CloudId
		}

		publicId := ""
		if g.PublicId != nil {
			publicId = *g.PublicId
		}

		protoGambar = append(protoGambar, &base.SoalGambar{
			Id:         int32(g.ID),
			NamaFile:   g.NamaFile,
			FilePath:   g.FilePath,
			FileSize:   int32(g.FileSize),
			MimeType:   g.MimeType,
			Urutan:     int32(g.Urutan),
			Keterangan: keterangan,
			CloudId:    cloudId,
			PublicId:   publicId,
			CreatedAt:  timestamppb.New(g.CreatedAt),
		})
	}
	return protoGambar
}

func convertDragItemsToProto(items []entity.DragItem) []*base.DragItem {
	if len(items) == 0 {
		return nil
	}

	var protoItems []*base.DragItem
	for _, item := range items {
		protoItem := &base.DragItem{
			Id:     int32(item.ID),
			Label:  item.Label,
			Urutan: int32(item.Urutan),
		}
		if item.ImageURL != nil {
			protoItem.ImageUrl = *item.ImageURL
		}
		protoItems = append(protoItems, protoItem)
	}
	return protoItems
}

func convertDragSlotsToProto(slots []entity.DragSlot) []*base.DragSlot {
	if len(slots) == 0 {
		return nil
	}

	var protoSlots []*base.DragSlot
	for _, slot := range slots {
		protoSlot := &base.DragSlot{
			Id:     int32(slot.ID),
			Label:  slot.Label,
			Urutan: int32(slot.Urutan),
		}
		protoSlots = append(protoSlots, protoSlot)
	}
	return protoSlots
}
